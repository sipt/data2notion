package notion

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/sipt/data2notion/model"

	"github.com/jomei/notionapi"
	"github.com/sipt/data2notion/global"
)

func SavePersons2Notion(persons []*model.Person) ([]notionapi.PageID, []error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(persons))
	pageMap := new(sync.Map)
	errs := make([]error, 0)
	for _, person := range persons {
		go func(person *model.Person) {
			defer func() {
				if e := recover(); e != nil {
					e := errors.Wrapf(e.(error), "Error: save person[%s] 2 notion failed", person.IMDBID)
					log.Printf(e.Error())
					errs = append(errs, e)
				}
			}()
			pageID, err := SavePerson2Notion(person)
			if err != nil {
				e := errors.Wrapf(err, "Error: save person[%s] 2 notion failed", person.IMDBID)
				log.Printf(e.Error())
				errs = append(errs, e)
			}
			pageMap.Store(person.IMDBID, pageID)
			wg.Done()
		}(person)
	}
	wg.Wait()
	pages := make([]notionapi.PageID, 0, len(persons))
	for _, person := range persons {
		if pageID, ok := pageMap.Load(person.IMDBID); ok {
			pages = append(pages, pageID.(notionapi.PageID))
		}
	}
	return pages, errs
}

func SavePerson2Notion(person *model.Person) (notionapi.PageID, error) {

	// create person to notion
	ctx := context.Background()
	k, v := "", ""
	if len(person.IMDBID) > 0 {
		k = "IMDB ID"
		v = person.IMDBID
	} else if person.TMDBID > 0 {
		k = "TMDB ID"
		v = fmt.Sprintf("%d", person.TMDBID)
	} else {
		k = "Douban ID"
		v = fmt.Sprintf("%s", person.DoubanID)
	}
	queryResp, err := notionClient.Database.Query(ctx, personDatabaseId, &notionapi.DatabaseQueryRequest{
		PropertyFilter: &notionapi.PropertyFilter{
			Property: k,
			RichText: &notionapi.TextFilterCondition{
				Equals: v,
			},
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "save person[Douban:%s, IMDBID:%s] filed", person.DoubanID, person.IMDBID)
	}
	if len(queryResp.Results) == 0 {
		return CreatePersonPageToDB(ctx, person)
	}
	if global.PersonUpdate {
		forUpdate := &queryResp.Results[0]
		return notionapi.PageID(forUpdate.ID), UpdatePersonPageToDB(ctx, person, forUpdate)
	}
	return notionapi.PageID(queryResp.Results[0].ID), nil
}

func UpdatePersonPageToDB(ctx context.Context, person *model.Person, exist *notionapi.Page) error {
	req := &notionapi.PageUpdateRequest{
		Properties: makeProperties(person),
	}
	_, err := notionClient.Page.Update(ctx, notionapi.PageID(exist.ID), req)
	return err
}

func CreatePersonPageToDB(ctx context.Context, person *model.Person) (notionapi.PageID, error) {
	req := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       "database_id",
			DatabaseID: personDatabaseId,
		},
		Properties: makeProperties(person),
	}
	if len(person.Profiles) > 0 {
		req.Icon = &notionapi.Icon{
			Type: "external",
			External: &notionapi.FileObject{
				URL: person.Profiles[0],
			},
		}
		//log.Printf("DEBUG: profile [%s->%s]", person.Name, person.Profiles[0])
	}
	page, err := notionClient.Page.Create(ctx, req)
	if err != nil {
		return "", err
	}
	return notionapi.PageID(page.ID), nil
}

func makeProperties(person *model.Person) notionapi.Properties {
	biography := []rune(person.Biography)
	if len(biography) > 2000 {
		biography = biography[:2000]
	}
	properties := notionapi.Properties{
		"Place Of Birth": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: person.PlaceOfBirth},
			}},
		},
		"IMDB ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: person.IMDBID},
			}},
		},
		"TMDB ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: fmt.Sprintf("%d", person.TMDBID)},
			}},
		},
		"Douban ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: person.DoubanID},
			}},
		},
		"Biography": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: string(biography)},
			}},
		},
		"Known For Department": notionapi.MultiSelectProperty{
			Type:        "multi_select",
			MultiSelect: formatTags(person.Jobs),
		},
		"Name": notionapi.TitleProperty{
			Type: "title",
			Title: []notionapi.RichText{
				{
					Type: "text",
					Text: notionapi.Text{
						Content: person.Name,
					},
				},
			},
		},
	}
	if len(person.Profiles) > 0 {
		properties["Profiles"] = makeImagePathProperty(person.Profiles)
	}
	if person.Deathday != nil {
		properties["Deathday"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: person.Deathday.String()},
			}},
		}
	}
	if person.Birthday != nil {
		properties["Birthday"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: person.Birthday.String()},
			}},
		}
	}
	if len(person.Gender) > 0 {
		properties["Gender"] = notionapi.SelectProperty{
			Type: "select",
			Select: notionapi.Option{
				Name: person.Gender,
			},
		}
	}
	if len(person.Alias) > 0 {
		properties["Alias"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: strings.Join(person.Alias, ", ")},
			}},
		}
	}

	// external link
	var texts []notionapi.RichText
	if len(person.DoubanID) > 0 {
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "豆瓣", Link: &notionapi.Link{Url: makeDoubanPersonLink(person.DoubanID)}},
		})
	}
	if person.TMDBID > 0 {
		if len(texts) > 0 {
			texts = append(texts, notionapi.RichText{
				Text: notionapi.Text{Content: " | "},
			})
		}
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "TMDB", Link: &notionapi.Link{Url: makeTMDBPersonLink(person.TMDBID)}},
		})
	}
	if len(person.IMDBID) > 0 {
		if len(texts) > 0 {
			texts = append(texts, notionapi.RichText{
				Text: notionapi.Text{Content: " | "},
			})
		}
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "IMDB", Link: &notionapi.Link{Url: makeIMDBPersonLink(person.IMDBID)}},
		})
	}

	properties["External Links"] = notionapi.RichTextProperty{
		Type:     "rich_text",
		RichText: texts,
	}

	return properties
}
