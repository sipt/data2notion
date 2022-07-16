package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jomei/notionapi"
	"github.com/pkg/errors"
	"github.com/sipt/data2notion/douban"
	"github.com/sipt/data2notion/global"
	"github.com/sipt/data2notion/model"
	"github.com/sipt/data2notion/notice"
	"github.com/sipt/data2notion/notion"
	"github.com/sipt/data2notion/tmdb"
	"github.com/spf13/cobra"

	aw "github.com/deanishe/awgo"
)

var personUpdate = false
var existSkip = true

var status string
var tags string

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	rootCmd.PersistentFlags().StringVarP(&status, "status", "s", "", "status of movie")
	rootCmd.PersistentFlags().StringVarP(&tags, "tags", "t", "", "tags of movie")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "server",
		Run: func(cmd *cobra.Command, args []string) {
			notice.ReceiveMessage(func(message string) {

			})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "s4n",
		Short: "search items in Notion",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 || len(args[0]) == 0 {
				global.WF.NewWarningItem("Please Input keyword", "").Icon(&aw.Icon{
					Value: "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
				})
				return
			}
			key := args[0]
			resp, err := notion.Search(key)
			if err != nil {
				global.WF.NewWarningItem("Search keyword failed", err.Error()).Icon(&aw.Icon{
					Value: "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
				})
				return
			}
			if len(resp.Results) == 0 {
				global.WF.NewWarningItem("Not found.", "").Icon(&aw.Icon{
					Value: "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
				})
				return
			}
			for _, page := range resp.Results {

				global.WF.NewItem(page.Properties["Title"].(*notionapi.TitleProperty).Title[0].Text.Content).
					Subtitle(fmt.Sprintf("[%s | %.1f] Download At: %s",
						page.Properties["Status"].(*notionapi.SelectProperty).Select.Name,
						page.Properties["Score"].(*notionapi.NumberProperty).Number/10,
						page.CreatedTime.Format("2006-01-02"),
					)).Arg(strings.ReplaceAll(page.URL, "https", "notion")).Icon(
					&aw.Icon{"/Applications/Notion.app", aw.IconTypeFileIcon}).Valid(true)
			}
			global.WF.SendFeedback()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "t2n",
		Short: "sync TMDB to Notion",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Print("[ERROR] Please input movieId")
				return
			}
			in := args[0]
			if len(in) == 0 {
				fmt.Print("[ERROR] Invalid movieId: ", in)
				return
			}
			ids := strings.Split(in, ",")
			for i, id := range ids {
				thread := notice.NewSlackThread(fmt.Sprintf("[%d/%d] Dump Media [%s] ", i+1, len(ids), id))
				thread.Send("Dump Media [%s]", id)
				dumpMediaWithDouban(TMDB, id, thread)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "d2n",
		Short: "sync TMDB to Notion",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Print("[ERROR] Please input movieId")
				return
			}
			in := args[0]
			if len(in) == 0 {
				fmt.Print("[ERROR] Invalid movieId: ", in)
				return
			}
			ids := strings.Split(in, ",")
			for i, id := range ids {
				thread := notice.NewSlackThread(fmt.Sprintf("[%d/%d] Dump Media [%s] ", i+1, len(ids), id))
				thread.Send("Dump Media [%s]", id)
				dumpMediaWithDouban(DOUBAN, id, thread)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "dump",
		Short: "sync douban collect to Notion",
		Run: func(cmd *cobra.Command, args []string) {
			before, current := -1, 600
			total := current + 1
			for current < total {
				var (
					ids []string
					err error
				)
				pageNotice := notice.NewSlackThread(fmt.Sprintf("Dump Collect [Cursor:%d] ", current))
				ids, total, err = douban.DumpCollect(before, current)
				if err != nil {
					pageNotice.Append("Error!")
					pageNotice.SendError(err)
					return
				}
				pageNotice.Append("Success!")
				for i, id := range ids {
					thread := notice.NewSlackThread(fmt.Sprintf("[%d/%d] Dump Media [%s] ", current+i+1, total, id))
					resp, err := notion.IsExist(&model.Media{
						DoubanID: id,
					})
					if err == nil && len(resp.Results) > 0 {
						text := ""
						title, ok := resp.Results[0].Properties["Title"].(*notionapi.TitleProperty)
						if ok && len(title.Title) > 0 {
							text = title.Title[0].Text.Content
						}
						text += " Skip!"
						thread.Append(text)
						continue
					}
					thread.Send("Dump Media [%s]", id)
					dumpMediaWithDouban(DOUBAN, id, thread)
				}
				before, current = current, current+30
			}
		},
	})
}

var rootCmd = &cobra.Command{
	Use:   "tmdb2notion",
	Short: "tmdb to nation",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}

const DOUBAN = "DOUBAN"
const TMDB = "TMDB"

func dumpMediaWithDouban(mainChanel string, id string, thread notice.INotifier) {
	defer func() {
		if e := recover(); e != nil {
			thread.SendError(e.(error))
		}
		thread.Finish(" Success!", " Failed!")
	}()
	if mainChanel == DOUBAN && strings.HasPrefix(id, "tt") {
		resp, err := douban.SearchFromDouban(id)
		if err != nil {
			thread.SendError(errors.Errorf("ERROR: Search Media by IMDBID Failed [%s:%s]", id, err.Error()))
			return
		}
		if len(resp.Payload.Items) == 0 {
			mainChanel = TMDB
			thread.SendError(errors.Errorf("ERROR: [DOUBAN:%s] Not found, switch to TMDB", id))
		} else {
			id = fmt.Sprintf("%d", resp.Payload.Items[0].ID)
		}
	}
	var media *model.Media
	var err error

	if mainChanel == DOUBAN {
		media, err = douban.DumpMedia(id)
		if err != nil {
			thread.SendError(errors.Errorf("ERROR: Dump Media Failed [%s:%s]", id, err.Error()))
			return
		}
		media.TMDBID, _ = tmdb.FindMediaIDByIMDBID(media.IMDBID)
	} else {
		result, err := tmdb.FindMediaByIMDBID(id)
		if err != nil {
			thread.SendError(errors.Errorf("ERROR: Dump Media Failed [%s:%s]", id, err.Error()))
			return
		}
		if result.Media == nil {
			thread.SendError(errors.Errorf("[ERROR] [IMDBID:%s] Not found in TMDB", id))
			return
		}
		media = result.Media
	}
	thread.Append(media.Title)
	thread.Send("Dump Media [%s] in [%s] Success!", media.Title, mainChanel)

	media.Status = status
	media.Tags = tags
	mediaPageID, err := notion.SaveMedia2Notion(media)
	if err != nil {
		thread.SendError(errors.Errorf("ERROR: Save Media [%s] to Notion Fialed: %s", media.Title, err.Error()))
		return
	}
	thread.Send("Save media to Notion [%s] success !", mediaPageID)
	// fetch cast & crew
	if mainChanel == DOUBAN {
		douban.DumpPersons(media, thread)
		thread.Send("Dump Persons Success [Director:%d] [Author:%d] [Actor:%d]!", len(media.Director), len(media.Author), len(media.Actor))
		if len(media.Author) == 0 || len(media.Director) == 0 || len(media.Actor) == 0 {
			result, err := tmdb.FindMediaByIMDBID(media.IMDBID)
			if err != nil {
				thread.SendError(errors.Wrapf(err, "ERROR: Find media [%d] on TMDB Failed", result.Media.TMDBID))
			} else {
				if len(media.Actor) > 0 {
					result.Media.Actor = nil
				}
				if len(media.Author) > 0 {
					result.Media.Author = nil
				}
				if len(media.Director) > 0 {
					result.Media.Director = nil
				}
				thread.Send(fmt.Sprintf("[Feedback] Find media [%d] on TMDB [Director:%d] [Author:%d] [Actor:%d]!",
					result.Media.TMDBID, len(result.Media.Director), len(result.Media.Author), len(result.Media.Actor)))
				errs := tmdb.FetchSeriesCast(result.Media)
				for _, err2 := range errs {
					thread.SendError(errors.Wrapf(err2, "ERROR: FetchSeriesCast [%d] on TMDB Failed", result.Media.TMDBID))
				}
				thread.Send("[Feedback] FetchSeriesCast [%d] on TMDB", result.Media.TMDBID)
				if len(result.Media.Actor) > 0 {
					media.Author = result.Media.Actor
				}
				if len(result.Media.Author) > 0 {
					media.Author = result.Media.Author
				}
				if len(result.Media.Director) > 0 {
					media.Director = result.Media.Director
				}
				media.TMDBID = result.Media.TMDBID
			}
		}
	} else {
		errs := tmdb.FetchSeriesCast(media)
		for _, err2 := range errs {
			thread.SendError(errors.Wrapf(err2, "ERROR: FetchSeriesCast [%d] on TMDB Failed", media.TMDBID))
		}
		thread.Send("Dump Persons Success [Director:%d] [Author:%d] [Actor:%d]!", len(media.Director), len(media.Author), len(media.Actor))
	}

	directorPageIds, errs := notion.SavePersons2Notion(media.Director)
	for _, err := range errs {
		if err != nil {
			thread.SendError(err)
		}
	}
	thread.Send("Save Persons [Director:%d] to Notion success !", len(media.Director))
	authorPageIds, errs := notion.SavePersons2Notion(media.Author)
	for _, err := range errs {
		if err != nil {
			thread.SendError(err)
		}
	}
	thread.Send("Save Persons [Author:%d] to Notion success !", len(media.Author))
	actorPageIds, errs := notion.SavePersons2Notion(media.Actor)
	for _, err := range errs {
		if err != nil {
			thread.SendError(err)
		}
	}
	thread.Send("Save Persons [Actor:%d] to Notion success !", len(media.Actor))
	// director
	err = notion.SaveDirectors2Notion(mediaPageID, directorPageIds)
	if err != nil {
		thread.SendError(errors.Errorf("ERROR: Save Persons' Relation [Director:%d] to Director Fialed: %s", len(media.Author), err.Error()))
	}
	thread.Send("Save Persons' Relation [Director:%d] to Director success !", len(media.Director))
	// author
	err = notion.SaveAuthors2Notion(mediaPageID, authorPageIds)
	if err != nil {
		thread.SendError(errors.Errorf("ERROR: Save Persons' Relation [Author:%d] to Notion Fialed: %s", len(media.Author), err.Error()))
	}
	thread.Send("Save Persons' Relation [Author:%d] to Notion success !", len(media.Author))
	// actor
	err = notion.SaveActors2Notion(mediaPageID, actorPageIds)
	if err != nil {
		thread.SendError(errors.Errorf("ERROR: Save Persons' Relation [Actor:%d] to Notion Fialed: %s", len(media.Actor), err.Error()))
	}
	thread.Send("Save Persons' Relation [Actor:%d] to Notion success !", len(media.Actor))
	fmt.Print("Success Dump: ", media.Title)
	thread.Send("Dump Finished !")
}
