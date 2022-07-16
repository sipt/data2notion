package notion

import (
	"os"

	"github.com/jomei/notionapi"
)

var notionClient *notionapi.Client

var movieDatabaseId notionapi.DatabaseID
var personDatabaseId notionapi.DatabaseID
var secret string

func init() {
	movieDatabaseId = notionapi.DatabaseID(os.Getenv("MOVIE_DBID"))
	personDatabaseId = notionapi.DatabaseID(os.Getenv("PERSON_DBID"))
	secret = os.Getenv("NOTION_SECRET")
	notionClient = notionapi.NewClient(notionapi.Token(secret))
}
