package tmdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sipt/data2notion/global"

	"github.com/sipt/data2notion/model"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func formatMovie(movie *tmdb.MovieDetails) *model.Media {
	return &model.Media{
		Title:             formatTitle(movie),
		Posters:           formatPosters(movie),
		TMDBID:            int(movie.ID),
		DoubanID:          "",
		IMDBID:            movie.IMDbID,
		Genres:            formatGenres(movie.Genres),
		ProductionCountry: formatCountry(movie),
		ReleaseDate:       formatReleaseDate(movie.ReleaseDate),
		RuntimeMin:        movie.Runtime,
		Score:             int(movie.VoteAverage * 10),
		RatingCount:       int(movie.VoteCount),
		Tagline:           movie.Tagline,
		Overview:          movie.Overview,
		Alias:             nil,
	}
}

func formatTitle(movie *tmdb.MovieDetails) string {
	year := ""
	items := strings.Split(movie.ReleaseDate, "-")
	if len(items) > 0 {
		year = items[0]
	} else {
		year = strconv.Itoa(time.Now().Year())
	}
	if movie.Title != movie.OriginalTitle {
		return fmt.Sprintf("%s %s (%s)", movie.Title, movie.OriginalTitle, year)
	} else {
		return fmt.Sprintf("%s (%s)", movie.Title, year)
	}
}

func formatGenres(genres []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) []string {
	options := make([]string, len(genres))
	for i, genre := range genres {
		options[i] = genre.Name
	}
	return options
}

func formatRunTime(value int) string {
	return fmt.Sprintf("%dh%dmin", value/60, value%60)
}

func formatReleaseDate(value string) *model.Date {
	if value == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		t = time.Time{}
	}
	return &model.Date{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}
}

func formatPosters(movie *tmdb.MovieDetails) []string {
	return []string{"https://image.tmdb.org/t/p/original" + movie.PosterPath}
}

func formatImagePath(path string) []string {
	if len(path) == 0 {
		return nil
	}
	return []string{"https://image.tmdb.org/t/p/original" + path}
}

func formatDate(dateStr string) *model.Date {
	if dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	return &model.Date{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}
}

func makeTMDBMovieLink(id int64) string {
	return fmt.Sprintf("https://www.themoviedb.org/movie/%d", id)
}
func makeIMDBMovieLink(id string) string {
	return "https://www.imdb.com/title/" + id
}

func formatCountry(movie *tmdb.MovieDetails) string {
	if len(movie.ProductionCountries) > 0 {
		code := movie.ProductionCountries[0].Iso3166_1
		return fmt.Sprintf("%s (%s)", global.CountryMap[code], code)
	}
	return ""
}

func formatGender(gender int) string {
	if gender == 2 {
		return "男"
	}
	return "女"
}

func formatDepartment(department string) string {
	switch department {
	case "Acting":
		return "演员"
	case "Directing":
		return "导演"
	case "Writing":
		return "编剧"
	case "Producer":
		return "制片人"
	}
	return "未知"
}
