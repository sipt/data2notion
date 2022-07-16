package notion

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jomei/notionapi"
	"github.com/sipt/data2notion/model"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func formatTitle(movie *tmdb.MovieDetails) string {
	year := ""
	items := strings.Split(movie.ReleaseDate, "-")
	if len(items) > 0 {
		year = items[0]
	} else {
		year = strconv.Itoa(time.Now().Year())
	}
	if movie.Title != movie.OriginalTitle {
		return fmt.Sprintf("%s | %s (%s)", movie.Title, movie.OriginalTitle, year)
	} else {
		return fmt.Sprintf("%s (%s)", movie.Title, year)
	}
}

func formatTags(tags []string) []notionapi.Option {
	options := make([]notionapi.Option, len(tags))
	for i, tag := range tags {
		options[i] = notionapi.Option{
			Name: tag,
		}
	}
	return options
}

func formatRunTime(value int) string {
	duration := ""
	if value/60 > 0 {
		duration = fmt.Sprintf("%d小时", value/60)
	}
	if value%60 > 0 {
		duration += fmt.Sprintf("%d分钟", value%60)
	}
	return duration
}

func formatReleaseDate(value *model.Date) *DateObject {
	if value == nil {
		return nil
	}
	d := Date(time.Date(value.Year, time.Month(value.Month), value.Day, 0, 0, 0, 0, time.Local))
	return &DateObject{Start: &d}
}

func makeImagePathProperty(paths []string) FilesProperty {
	if len(paths) == 0 {
		return FilesProperty{Type: "files",
			Files: []File{
				{
					Type: "external",
					Name: "Poster",
				},
			}}
	}
	file := FilesProperty{
		Type: "files",
	}
	for _, path := range paths {

		file.Files = append(file.Files, File{
			Type:     "external",
			Name:     "Poster",
			External: &External{Url: path},
		})
	}
	return file
}

func makeTMDBImagePath(path string) string {
	return "https://image.tmdb.org/t/p/original" + path
}

func makeDoubanMediaLink(id string) string {
	return fmt.Sprintf("https://movie.douban.com/subject/%s/", id)
}
func makeTMDBMovieLink(id int) string {
	return fmt.Sprintf("https://www.themoviedb.org/movie/%d", id)
}
func makeIMDBMovieLink(id string) string {
	return "https://www.imdb.com/title/" + id
}

func makeDoubanPersonLink(id string) string {
	return fmt.Sprintf("https://movie.douban.com/celebrity/%s/", id)
}
func makeTMDBPersonLink(id int) string {
	return fmt.Sprintf("https://www.themoviedb.org/person/%d", id)
}
func makeIMDBPersonLink(id string) string {
	return "https://www.imdb.com/name/" + id
}
func formatCountry(movie *tmdb.MovieDetails) string {
	if len(movie.ProductionCountries) > 0 {
		code := movie.ProductionCountries[0].Iso3166_1
		return fmt.Sprintf("%s (%s)", CountryMap[code], code)
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
	case "Director":
		return "导演"
	}
	return "未知"
}
