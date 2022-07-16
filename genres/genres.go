package main

import (
	"context"
	"math/rand"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/jomei/notionapi"
)

var tmdbClient *tmdb.Client
var notionClient *notionapi.Client

const databaseId = "5664ae118c624f7cb804ee41727f9313"

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	tmdbClient, err = tmdb.Init("0962fd5de57114ff511916e25e760d75")
	if err != nil {
		panic(err)
	}
	notionClient = notionapi.NewClient("secret_08No8Lm8gWwYPdrWpIqFgzikxSW6VfWXqVwFEymwrlg")
}

func main() {
	genres, err := tmdbClient.GetGenreMovieList(map[string]string{"language": "zh-CN"})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	options := make([]notionapi.Option, len(genres.Genres))
	for i, genre := range genres.Genres {
		options[i] = notionapi.Option{
			Name:  genre.Name,
			Color: randomColor(),
		}
	}
	_, err = notionClient.Database.Update(ctx, databaseId, &notionapi.DatabaseUpdateRequest{
		Properties: map[string]notionapi.PropertyConfig{
			"Genres": &notionapi.MultiSelectPropertyConfig{
				Type:        "multi_select",
				MultiSelect: notionapi.Select{Options: options},
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

func randomColor() notionapi.Color {
	colors := []notionapi.Color{notionapi.ColorDefault,
		notionapi.ColorGray,
		notionapi.ColorBrown,
		notionapi.ColorOrange,
		notionapi.ColorYellow,
		notionapi.ColorGreen,
		notionapi.ColorBlue,
		notionapi.ColorPurple,
		notionapi.ColorPink,
		notionapi.ColorRed}
	return colors[rand.Int()%len(colors)]
}
