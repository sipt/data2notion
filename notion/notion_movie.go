package notion

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipt/data2notion/model"

	"github.com/jomei/notionapi"
)

func IsExist(media *model.Media) (*notionapi.DatabaseQueryResponse, error) {
	ctx := context.Background()
	k, v := "", ""
	if len(media.IMDBID) > 0 {
		k = "IMDB ID"
		v = media.IMDBID
	} else if media.TMDBID > 0 {
		k = "TMDB ID"
		v = fmt.Sprintf("%d", media.TMDBID)
	} else {
		k = "Douban ID"
		v = fmt.Sprintf("%s", media.DoubanID)
	}
	return notionClient.Database.Query(ctx, movieDatabaseId, &notionapi.DatabaseQueryRequest{
		PropertyFilter: &notionapi.PropertyFilter{
			Property: k,
			RichText: &notionapi.TextFilterCondition{
				Equals: v,
			},
		},
	})
}

func Search(keyword string) (*notionapi.DatabaseQueryResponse, error) {
	resp, err := notionClient.Database.Query(context.Background(), movieDatabaseId, &notionapi.DatabaseQueryRequest{
		CompoundFilter: &notionapi.CompoundFilter{
			notionapi.FilterOperatorOR: []notionapi.PropertyFilter{
				{
					Property: "Title",
					RichText: &notionapi.TextFilterCondition{
						Contains: keyword,
					},
				}, {
					Property: "IMDB ID",
					RichText: &notionapi.TextFilterCondition{
						Contains: keyword,
					},
				}, {
					Property: "Douban ID",
					RichText: &notionapi.TextFilterCondition{
						Contains: keyword,
					},
				}, {
					Property: "TMDB ID",
					RichText: &notionapi.TextFilterCondition{
						Contains: keyword,
					},
				},
			},
		},
		Sorts: nil,
	})
	return resp, err
}

func SaveAuthors2Notion(moviePageID notionapi.PageID, authorPageIDs []notionapi.PageID) error {
	if len(authorPageIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	relations := make([]notionapi.Relation, len(authorPageIDs))
	for i, id := range authorPageIDs {
		relations[i] = notionapi.Relation{ID: id}
	}
	_, err := notionClient.Page.Update(ctx, moviePageID, &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Authors": notionapi.RelationProperty{
				Type:     "relation",
				Relation: relations,
			},
		},
	})
	return err
}

func SaveActors2Notion(moviePageID notionapi.PageID, actorPageIDs []notionapi.PageID) error {
	if len(actorPageIDs) > 100 {
		actorPageIDs = actorPageIDs[:100]
	}
	if len(actorPageIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	relations := make([]notionapi.Relation, len(actorPageIDs))
	for i, id := range actorPageIDs {
		relations[i] = notionapi.Relation{ID: id}
	}
	_, err := notionClient.Page.Update(ctx, moviePageID, &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Actors": notionapi.RelationProperty{
				Type:     "relation",
				Relation: relations,
			},
		},
	})
	return err
}

func SaveDirectors2Notion(moviePageID notionapi.PageID, directorPageIDs []notionapi.PageID) error {
	if len(directorPageIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	var relations []notionapi.Relation
	for _, pageID := range directorPageIDs {
		relations = append(relations, notionapi.Relation{ID: pageID})
	}
	_, err := notionClient.Page.Update(ctx, moviePageID, &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Director": notionapi.RelationProperty{
				Type:     "relation",
				Relation: relations,
			},
		},
	})
	return err
}

func SaveMedia2Notion(media *model.Media) (notionapi.PageID, error) {
	// create media to notion
	ctx := context.Background()
	queryResp, err := IsExist(media)
	if err != nil {
		return "", err
	}
	if len(queryResp.Results) == 0 {
		return CreateMoviePageToDB(ctx, media)
	}
	forUpdate := &queryResp.Results[0]
	return notionapi.PageID(forUpdate.ID), UpdateMoviePageToDB(ctx, media, forUpdate)
}

func UpdateMoviePageToDB(ctx context.Context, media *model.Media, exist *notionapi.Page) error {
	req := &notionapi.PageUpdateRequest{
		Properties: makeMovieProperties(media),
	}
	_, err := notionClient.Page.Update(ctx, notionapi.PageID(exist.ID), req)
	return err
}

func CreateMoviePageToDB(ctx context.Context, media *model.Media) (notionapi.PageID, error) {
	req := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       "database_id",
			DatabaseID: movieDatabaseId,
		},
		Properties: makeMovieProperties(media),
	}
	page, err := notionClient.Page.Create(ctx, req)
	if err != nil {
		return "", err
	}
	return notionapi.PageID(page.ID), err
}

func makeMovieProperties(media *model.Media) notionapi.Properties {
	properties := notionapi.Properties{
		"IMDB ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: media.IMDBID},
			}},
		},
		"TMDB ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: fmt.Sprintf("%d", media.TMDBID)},
			}},
		},
		"Douban ID": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: media.DoubanID},
			}},
		},
		"Score": notionapi.NumberProperty{
			Type:   "number",
			Number: float64(media.Score),
		},
		"Rating Count": notionapi.NumberProperty{
			Type:   "number",
			Number: float64(media.RatingCount),
		},
		"Overview": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: strings.TrimSpace(media.Overview)},
			}},
		},
		"Posters": makeImagePathProperty(media.Posters),
		"Release Date": DateProperty{
			Type: "date",
			Date: formatReleaseDate(media.ReleaseDate),
		},
		"Production Country": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: media.ProductionCountry},
			}},
		},
		"Tag Line": notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: strings.TrimSpace(media.Tagline)},
			}},
		},
		"Genres": notionapi.MultiSelectProperty{
			Type:        "multi_select",
			MultiSelect: formatTags(media.Genres),
		},
		"Title": notionapi.TitleProperty{
			Type: "title",
			Title: []notionapi.RichText{
				{
					Type: "text",
					Text: notionapi.Text{
						Content: media.Title,
					},
				},
			},
		},
	}
	if len(media.Status) > 0 {
		properties["Status"] = notionapi.SelectProperty{
			Type: "select",
			Select: notionapi.Option{
				Name: media.Status,
			},
		}
	}
	if len(media.Tags) > 0 {
		tags := strings.Split(media.Tags, ",")
		properties["Tags"] = notionapi.MultiSelectProperty{
			Type:        "multi_select",
			MultiSelect: formatTags(tags),
		}
	}
	if media.Season > 0 {
		mediaType := "TV Season"
		for _, genre := range media.Genres {
			if genre == "真人秀" {
				mediaType = "TV Show"
				break
			}
		}
		properties["Media Type"] = notionapi.SelectProperty{
			Type: "select",
			Select: notionapi.Option{
				Name: mediaType,
			},
		}
		properties["Season"] = notionapi.NumberProperty{
			Type:   "number",
			Number: float64(media.Season),
		}
		properties["Episode Count"] = notionapi.NumberProperty{
			Type:   "number",
			Number: float64(media.EpisodeCount),
		}
		preview := fmt.Sprintf("第%d季 / 共%d集", media.Season, media.EpisodeCount)
		if media.RuntimeMin > 0 {
			preview += " / 每集" + formatRunTime(media.RuntimeMin)
		}
		properties["Preview"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: fmt.Sprintf("%s (%s | %s)", preview, media.ProductionCountry, media.Language)},
			}},
		}
	} else {
		properties["Media Type"] = notionapi.SelectProperty{
			Type: "select",
			Select: notionapi.Option{
				Name: "Movie",
			},
		}
		properties["Preview"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: fmt.Sprintf("%s (%s | %s)", formatRunTime(media.RuntimeMin), media.ProductionCountry, media.Language)},
			}},
		}
	}
	if len(media.Alias) > 0 {
		properties["Alias"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: strings.Join(media.Alias, ", ")},
			}},
		}
	}
	if len(media.Language) > 0 {
		properties["Language"] = notionapi.RichTextProperty{
			Type: "rich_text",
			RichText: []notionapi.RichText{{
				Text: notionapi.Text{Content: media.Language},
			}},
		}
	}
	if media.RuntimeMin > 0 {
		properties["Duration"] = notionapi.NumberProperty{
			Type:   "number",
			Number: float64(media.RuntimeMin),
		}
	}

	// external link
	var texts []notionapi.RichText
	if len(media.DoubanID) > 0 {
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "豆瓣", Link: &notionapi.Link{Url: makeDoubanMediaLink(media.DoubanID)}},
		})
	}
	if media.TMDBID > 0 {
		if len(texts) > 0 {
			texts = append(texts, notionapi.RichText{
				Text: notionapi.Text{Content: " | "},
			})
		}
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "TMDB", Link: &notionapi.Link{Url: makeTMDBMovieLink(media.TMDBID)}},
		})
	}
	if len(media.IMDBID) > 0 {
		if len(texts) > 0 {
			texts = append(texts, notionapi.RichText{
				Text: notionapi.Text{Content: " | "},
			})
		}
		texts = append(texts, notionapi.RichText{
			Text: notionapi.Text{Content: "IMDB", Link: &notionapi.Link{Url: makeIMDBMovieLink(media.IMDBID)}},
		})
	}

	properties["External Links"] = notionapi.RichTextProperty{
		Type:     "rich_text",
		RichText: texts,
	}
	return properties
}
