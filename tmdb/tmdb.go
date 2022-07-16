package tmdb

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sipt/data2notion/global"
	"github.com/sipt/data2notion/model"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var tmdbClient *tmdb.Client
var tmdbKey string

func init() {
	var err error
	tmdbKey = os.Getenv("TMDB_API_KEY")
	tmdbClient, err = tmdb.Init(tmdbKey)
	if err != nil {
		panic(err)
	}
}

type TMDBFindResult struct {
	Media  *model.Media
	Person *model.Person
}

func FindMediaIDByIMDBID(imdbID string) (int, error) {
	options := map[string]string{
		"language":        "zh-CN",
		"external_source": "imdb_id",
	}
	resp, err := tmdbClient.GetFindByID(imdbID, options)
	if err != nil {

		return 0, errors.Wrapf(err, "TMDB findbyid[%s] failed", imdbID)
	}
	if len(resp.MovieResults) > 0 {
		return int(resp.MovieResults[0].ID), nil
	}
	return 0, nil
}

func FindMediaByIMDBID(imdbID string) (*TMDBFindResult, error) {
	options := map[string]string{
		"language":        "zh-CN",
		"external_source": "imdb_id",
	}
	resp, err := tmdbClient.GetFindByID(imdbID, options)
	if err != nil {
		return nil, errors.Wrapf(err, "TMDB findbyid[%s] failed", imdbID)
	}
	result := &TMDBFindResult{}
	if len(resp.MovieResults) > 0 {
		result.Media, err = GetMovieById(int(resp.MovieResults[0].ID))
	} else if len(resp.PersonResults) > 0 {
		result.Person, err = getPersonDetails(int(resp.PersonResults[0].ID))
	}
	return result, nil
}

func GetMovieById(movieId int) (*model.Media, error) {
	cacheKey := fmt.Sprintf("movie_%d", movieId)
	movie := &model.Media{}
	var err error
	if global.WF != nil && !global.WF.Cache.Expired(cacheKey, time.Hour*3) {
		if data, err := global.WF.Cache.Load(cacheKey); err == nil {
			if err := json.Unmarshal(data, movie); err == nil {
				return movie, nil
			}
		}
	}
	movie, err = getMovieByIdFromRemote(movieId)
	if err != nil {
		return nil, err
	}
	if global.WF != nil {
		data, _ := json.Marshal(movie)
		err = global.WF.Cache.Store(cacheKey, data)
		if err != nil {
			log.Printf("ERROR: store cache failed[%s]: %s", cacheKey, err)
		}
	}
	return movie, nil
}

func getMovieByIdFromRemote(movieId int) (*model.Media, error) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	var zhMovie *tmdb.MovieDetails
	var enMovie *tmdb.MovieDetails
	var zhErr, enErr error
	go func() {
		options := map[string]string{
			"language":           "zh-CN",
			"append_to_response": "credits,images",
		}
		zhMovie, zhErr = tmdbClient.GetMovieDetails(movieId, options)
		wg.Done()
	}()
	go func() {
		options := map[string]string{
			"language":           "en-US",
			"append_to_response": "credits,images",
		}
		enMovie, enErr = tmdbClient.GetMovieDetails(movieId, options)
		wg.Done()
	}()
	wg.Wait()
	if zhErr != nil {
		return nil, zhErr
	}
	if enErr == nil {
		MergeMovie(enMovie, zhMovie)
	}
	var media = &model.Media{
		Title:             formatTitle(zhMovie),
		Posters:           formatImagePath(zhMovie.PosterPath),
		TMDBID:            int(zhMovie.ID),
		DoubanID:          "",
		IMDBID:            zhMovie.IMDbID,
		Genres:            formatGenres(zhMovie.Genres),
		ProductionCountry: formatCountry(zhMovie),
		ReleaseDate:       formatReleaseDate(zhMovie.ReleaseDate),
		RuntimeMin:        zhMovie.Runtime,
		RatingCount:       int(zhMovie.VoteCount),
		Score:             int(zhMovie.VoteAverage * 10),
		Tagline:           zhMovie.Tagline,
		Overview:          zhMovie.Overview,
		Alias:             nil,
		Season:            0,
		EpisodeCount:      0,
		EpisodeDuration:   0,
		Language:          global.LanguageMap[zhMovie.OriginalLanguage],
		Director:          nil,
		Author:            nil,
		Actor:             nil,
	}
	for _, actor := range zhMovie.MovieCreditsAppend.Credits.MovieCredits.Cast {
		media.Actor = append(media.Actor, &model.Person{Name: actor.Name, TMDBID: int(actor.ID)})
	}

	for _, crew := range zhMovie.MovieCreditsAppend.Credits.MovieCredits.Crew {
		if crew.Job == "Writer" {
			media.Author = append(media.Author, &model.Person{Name: crew.Name, TMDBID: int(crew.ID)})
		} else if crew.Job == "Director" {
			media.Director = append(media.Director, &model.Person{Name: crew.Name, TMDBID: int(crew.ID)})
		}
	}
	return media, nil
}

func MergeMovie(from, to *tmdb.MovieDetails) {
	if from == nil || to == nil {
		return
	}
	if to.Title == "" {
		to.Title = from.Title
	}
	if to.Overview == "" {
		to.Overview = from.Overview
	}
	if to.Tagline == "" {
		to.Tagline = from.Tagline
	}
	if to.ReleaseDate == "" {
		to.ReleaseDate = from.ReleaseDate
	}
}

func FetchSeriesCast(media *model.Media) []error {
	wg := &sync.WaitGroup{}
	wg.Add(len(media.Actor))
	actorMap := &sync.Map{}
	var errs []error
	for _, s := range media.Actor {
		go func(id int) {
			personDetail, err := getPersonDetails(id)
			if err != nil {
				e := errors.Wrapf(err, "Error: GetPersonDetails[%d] failed", id)
				errs = append(errs, e)
				log.Print(e.Error())
			}
			actorMap.Store(personDetail.TMDBID, personDetail)
			wg.Done()
		}(s.TMDBID)
	}

	directorMap := &sync.Map{}
	wg.Add(len(media.Director))
	for _, s := range media.Director {
		go func(id int) {
			personDetail, err := getPersonDetails(id)
			if err != nil {
				e := errors.Wrapf(err, "Error: GetPersonDetails[%d] failed", id)
				errs = append(errs, e)
				log.Print(e.Error())
			}
			directorMap.Store(personDetail.TMDBID, personDetail)
			wg.Done()
		}(s.TMDBID)
		break
	}

	authorMap := &sync.Map{}
	wg.Add(len(media.Author))
	for _, s := range media.Author {
		go func(id int) {
			personDetail, err := getPersonDetails(id)
			if err != nil {
				e := errors.Wrapf(err, "Error: GetPersonDetails[%d] failed", id)
				errs = append(errs, e)
				log.Print(e.Error())
			}
			authorMap.Store(personDetail.TMDBID, personDetail)
			wg.Done()
		}(s.TMDBID)
		break
	}
	wg.Wait()
	actorPersons := make([]*model.Person, 0)
	for _, person := range media.Actor {
		if personDetail, ok := actorMap.Load(person.TMDBID); ok {
			actorPersons = append(actorPersons, personDetail.(*model.Person))
		}
	}
	authorPersons := make([]*model.Person, 0)
	for _, person := range media.Author {
		if personDetail, ok := authorMap.Load(person.TMDBID); ok {
			authorPersons = append(authorPersons, personDetail.(*model.Person))
		}
	}
	directorPersons := make([]*model.Person, 0)
	for _, person := range media.Director {
		if personDetail, ok := directorMap.Load(person.TMDBID); ok {
			directorPersons = append(directorPersons, personDetail.(*model.Person))
		}
	}
	media.Actor = actorPersons
	media.Author = authorPersons
	media.Director = directorPersons
	return errs
}

func getPersonDetails(personId int) (*model.Person, error) {
	cacheKey := fmt.Sprintf("person_%d", personId)
	tmdbPerson := &model.Person{}
	var err error
	if global.WF != nil && !global.WF.Cache.Expired(cacheKey, time.Hour*3) {
		if data, err := global.WF.Cache.Load(cacheKey); err == nil {
			if err := json.Unmarshal(data, tmdbPerson); err == nil {
				return tmdbPerson, nil
			}
		}
	}
	tmdbPerson, err = getPersonDetailsFromRemote(personId)
	if err != nil {
		return nil, err
	}

	if global.WF != nil {
		data, _ := json.Marshal(tmdbPerson)
		err = global.WF.Cache.Store(cacheKey, data)
		if err != nil {
			log.Printf("ERROR: store cache failed[%s]: %s", cacheKey, err)
		}
	}
	return tmdbPerson, nil
}

func getPersonDetailsFromRemote(id int) (*model.Person, error) {
	options := map[string]string{
		"language": "zh-CN",
	}
	personDetail, err := tmdbClient.GetPersonDetails(id, options)
	if err != nil {
		return nil, err
	}
	person := &model.Person{
		Name:         personDetail.Name,
		Profiles:     formatImagePath(personDetail.ProfilePath),
		Gender:       formatGender(personDetail.Gender),
		Birthday:     formatDate(personDetail.Birthday),
		PlaceOfBirth: personDetail.PlaceOfBirth,
		Deathday:     formatDate(personDetail.Deathday),
		Alias:        personDetail.AlsoKnownAs,
		Biography:    personDetail.Biography,
		Jobs:         []string{formatDepartment(personDetail.KnownForDepartment)},
		TMDBID:       int(personDetail.ID),
		DoubanID:     "",
		IMDBID:       personDetail.IMDbID,
	}
	return person, err
}
