package douban

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sipt/data2notion/notice"

	"github.com/gocolly/colly"
	"github.com/sipt/data2notion/model"
)

var jar *cookiejar.Jar
var UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15"
var cacheDir func(collector *colly.Collector)

func init() {
	jar, _ = cookiejar.New(nil)
	u, _ := url.Parse("https://movie.douban.com")
	jar.SetCookies(u, []*http.Cookie{{Domain: "m.douban.com", Path: "/movie/subject/30314848", Secure: false, Name: "talionnav_show_app", Value: "0"},
		{Domain: "m.douban.com", Path: "/movie/subject/1293181", Secure: false, Name: "talionnav_show_app", Value: "0"},
		{Domain: "m.douban.com", Path: "/movie/subject/1866471", Secure: false, Name: "talionnav_show_app", Value: "0"},
		{Domain: "m.douban.com", Path: "/movie", Secure: false, Name: "talionnav_show_app", Value: "0"},
		{Domain: "m.douban.com", Path: "/pwa", Secure: false, Name: "talionnav_show_app", Value: "0"},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "apiKey", Value: ""},
		{Domain: ".book.douban.com", Path: "/", Secure: false, Name: "__utmc", Value: "81379588"},
		{Domain: ".book.douban.com", Path: "/", Secure: false, Name: "_vwo_uuid_v2", Value: "D39E045B87B8A7C3F842BBE79E0030ED5|aa1a29b5b858569e7671e534bd4cadcd"},
		{Domain: "book.douban.com", Path: "/", Secure: false, Name: "_pk_ref.100001.3ac3", Value: "%5B%22%22%2C%22%22%2C1652885503%2C%22https%3A%2F%2Fmovie.douban.com%2Fsubject_search%3Fsearch_text%3D%25E6%259E%25B6%25E6%259E%2584%25E5%258D%25B3%25E6%259C%25AA%25E6%259D%25A5%22%5D"},
		{Domain: ".book.douban.com", Path: "/", Secure: false, Name: "__utma", Value: "81379588.1068776736.1537271685.1652628323.1652885503.3"},
		{Domain: ".book.douban.com", Path: "/", Secure: false, Name: "__utmz", Value: "81379588.1652885503.3.3.utmcsr=movie.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/subject_search"},
		{Domain: "book.douban.com", Path: "/", Secure: false, Name: "_pk_id.100001.3ac3", Value: "cc03918c51ea0f39.1650212653.3.1652885520.1652628333."},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "_pk_ref.100001.2fad", Value: "%5B%22%22%2C%22%22%2C1653216284%2C%22https%3A%2F%2Fmovie.douban.com%2Fphotos%2Fphoto%2F990592243%2F%22%5D"},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "_pk_id.100001.2fad", Value: "682ae53b414bcef5.1653216284.1.1653216284.1653216284."},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "vtoken", Value: "undefined"},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "last_login_way", Value: "phone"},
		{Domain: ".m.douban.com", Path: "/", Secure: true, Name: "frodotk", Value: "9c6b3ac347d80143d541b049d78803c9"},
		{Domain: ".m.douban.com", Path: "/", Secure: false, Name: "talionusr", Value: "eyJpZCI6ICIxNTM2MzQzODAiLCAibmFtZSI6ICJzaXB0In0="},
		{Domain: ".m.douban.com", Path: "/", Secure: false, Name: "Hm_lvt_6d4a8cfea88fa457c3127e14fb5fabc2", Value: "1651248676,1653227524"},
		{Domain: ".m.douban.com", Path: "/", Secure: false, Name: "Hm_lpvt_6d4a8cfea88fa457c3127e14fb5fabc2", Value: "1653410753"},
		{Domain: "search.douban.com", Path: "/", Secure: false, Name: "_pk_ref.100001.2939", Value: "%5B%22%22%2C%22%22%2C1653556655%2C%22https%3A%2F%2Fmovie.douban.com%2Fcelebrity%2F1274477%2F%22%5D"},
		{Domain: "search.douban.com", Path: "/", Secure: false, Name: "_pk_id.100001.2939", Value: "d7d4dbf8bc337830.1637258076.50.1653556720.1653495263."},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "ll", Value: "108296"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "bid", Value: "u1EDyQA0zSg"},
		{Domain: "movie.douban.com", Path: "/", Secure: false, Name: "_pk_ses.100001.4cf6", Value: "*"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "ap_v", Value: "0,6.0"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "__utma", Value: "30149280.596376423.1653580443.1653580443.1653580443.1"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "__utmc", Value: "30149280"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "__utmz", Value: "30149280.1653580443.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none)"},
		{Domain: ".movie.douban.com", Path: "/", Secure: false, Name: "__utma", Value: "223695111.1888166266.1653580443.1653580443.1653580443.1"},
		{Domain: ".movie.douban.com", Path: "/", Secure: false, Name: "__utmb", Value: "223695111.0.10.1653580443"},
		{Domain: ".movie.douban.com", Path: "/", Secure: false, Name: "__utmc", Value: "223695111"},
		{Domain: ".movie.douban.com", Path: "/", Secure: false, Name: "__utmz", Value: "223695111.1653580443.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none)"},
		{Domain: "www.douban.com", Path: "/", Secure: false, Name: "_pk_ses.100001.8cb4", Value: "*"},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "login_start_time", Value: "1653580782659"},
		{Domain: "accounts.douban.com", Path: "/", Secure: false, Name: "user_data", Value: "{%22area_code%22:%22+86%22%2C%22number%22:%2218117400405%22%2C%22code%22:%222823%22}"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "dbcl2", Value: "153634380:AhVe/pTw468"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "ck", Value: "4sUW"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "push_noty_num", Value: "0"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "push_doumail_num", Value: "0"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "__utmv", Value: "30149280.15363"},
		{Domain: ".douban.com", Path: "/", Secure: false, Name: "__utmb", Value: "30149280.6.10.1653580443"},
		{Domain: "www.douban.com", Path: "/", Secure: false, Name: "_pk_id.100001.8cb4", Value: "2d8988a824dfeda5.1653580724.1.1653580857.1653580724."},
		{Domain: "movie.douban.com", Path: "/", Secure: false, Name: "_pk_id.100001.4cf6", Value: "3dccabd9e6266d3b.1653580443.1.1653581321.1653580443."}})
	cacheDir = colly.CacheDir(os.TempDir() + "/data2notion/cache")
}

func DumpMedia(subjectID string) (*model.Media, error) {
	media, err := dumpMedia(subjectID)
	if err != nil {
		return nil, err
	}
	return media, nil
}

func DumpPersons(media *model.Media, notifier notice.INotifier) {
	{
		var director []*model.Person
		directorThread := notifier.MakeThread("Dump Director ")
		directorThread.Update(directorThread.Title()+" [0/%d]", len(media.Director))
		for i, c := range media.Director {
			detail, err := DumpCelebrity(c.DoubanID)
			if err != nil {
				notifier.SendError(fmt.Errorf("dump celebrity[%s] failed: %s", c.DoubanID, err.Error()))
			} else {
				director = append(director, detail)
			}
			directorThread.Update(directorThread.Title()+" [%d/%d]", i+1, len(media.Director))
		}
		media.Director = director
	}

	{
		var author []*model.Person
		authorThread := notifier.MakeThread("Dump Author ")
		authorThread.Update(authorThread.Title()+" s[0/%d]", len(media.Director))
		for i, c := range media.Author {
			detail, err := DumpCelebrity(c.DoubanID)
			if err != nil {
				notifier.SendError(fmt.Errorf("dump celebrity[%s] failed: %s", c.DoubanID, err.Error()))
			} else {
				author = append(author, detail)
			}
			authorThread.Update(authorThread.Title()+" [%d/%d]", i+1, len(media.Author))
		}
		media.Author = author
	}
	{
		var actor []*model.Person
		actorThread := notifier.MakeThread("Dump Actor ")
		actorThread.Update(actorThread.Title()+" [0/%d]", len(media.Director))
		for i, c := range media.Actor {
			detail, err := DumpCelebrity(c.DoubanID)
			if err != nil {
				notifier.SendError(fmt.Errorf("dump celebrity[%s] failed: %s", c.DoubanID, err.Error()))
			} else {
				actor = append(actor, detail)
			}
			actorThread.Update(actorThread.Title()+" [%d/%d]", i+1, len(media.Actor))
		}
		media.Actor = actor
	}
}

type DoubanMovie struct {
	Context       string   `json:"@context"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Image         string   `json:"image"`
	DatePublished string   `json:"datePublished"`
	Genre         []string `json:"genre"`
	Duration      string   `json:"duration"`
	Description   string   `json:"description"`
	Type          string   `json:"@type"`
	Director      []struct {
		Type string `json:"@type"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"director"`
	Author []struct {
		Type string `json:"@type"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"author"`
	Actor []struct {
		Type string `json:"@type"`
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"actor"`
	AggregateRating struct {
		Type        string `json:"@type"`
		RatingCount string `json:"ratingCount"`
		BestRating  string `json:"bestRating"`
		WorstRating string `json:"worstRating"`
		RatingValue string `json:"ratingValue"`
	} `json:"aggregateRating"`
}

func dumpMedia(subjectID string) (*model.Media, error) {
	movie := new(DoubanMovie)
	media := new(model.Media)
	c := colly.NewCollector()
	cacheDir(c)
	c.SetCookieJar(jar)
	c.UserAgent = UserAgent

	// Find and visit all links
	var decodeJsonError error
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		err := json.Unmarshal([]byte(strings.ReplaceAll(e.Text, "\n", "")), movie)
		if err != nil {
			decodeJsonError = errors.New(e.Text + "\n" + err.Error())
		}
	})
	c.OnHTML("#content>h1", func(e *colly.HTMLElement) {
		media.Title = e.ChildText("span")
	})
	c.OnHTML("span.all.hidden", func(e *colly.HTMLElement) {
		media.Overview = strings.TrimSpace(e.Text)
	})
	c.OnHTML("#link-report>span[property='v:summary']", func(e *colly.HTMLElement) {
		media.Overview = strings.TrimSpace(e.Text)
	})
	c.OnHTML(fmt.Sprintf("#season>option[value='%s']", subjectID), func(e *colly.HTMLElement) {
		media.Season, _ = strconv.Atoi(strings.TrimSpace(e.Text))
	})
	c.OnHTML("#info", func(e *colly.HTMLElement) {
		rows := strings.Split(e.Text, "\n")
		for _, row := range rows {
			row = strings.TrimSpace(row)
			if len(row) == 0 {
				continue
			}
			cells := strings.Split(row, ":")
			if len(cells) < 2 {
				continue
			}
			switch cells[0] {
			case "制片国家/地区":
				media.ProductionCountry = strings.TrimSpace(cells[1])
			case "语言":
				media.Language = strings.TrimSpace(cells[1])
			case "又名":
				media.Alias = strings.Split(strings.TrimSpace(cells[1]), " / ")
			case "集数":
				media.EpisodeCount, _ = strconv.Atoi(strings.TrimSpace(cells[1]))
			case "单集片长":
				duration := strings.TrimSuffix(strings.TrimSpace(cells[1]), "分钟")
				media.EpisodeDuration, _ = strconv.Atoi(duration)
			case "IMDb":
				media.IMDBID = strings.TrimSpace(cells[1])
			}
		}
	})

	err := c.Visit(fmt.Sprintf("https://movie.douban.com/subject/%s/", subjectID))
	if err != nil {
		return nil, err
	}
	if decodeJsonError != nil {
		return nil, decodeJsonError
	}
	media.Posters = []string{movie.Image}
	media.DoubanID = subjectID
	media.Genres = movie.Genre
	media.ReleaseDate = formatDate(movie.DatePublished)
	media.RuntimeMin = decodeDuration(movie.Duration)
	media.RatingCount, _ = strconv.Atoi(movie.AggregateRating.RatingCount)
	score, _ := strconv.ParseFloat(movie.AggregateRating.RatingValue, 64)
	if score > 0 {
		media.Score = int(score * 10)
	}
	for _, c := range movie.Director {
		cells := strings.Split(c.URL, "/")
		if len(cells) == 4 {
			media.Director = append(media.Director, &model.Person{
				Name:     c.Name,
				DoubanID: cells[2],
			})
		}
	}
	for _, c := range movie.Author {
		cells := strings.Split(c.URL, "/")
		if len(cells) == 4 {
			media.Author = append(media.Author, &model.Person{
				Name:     c.Name,
				DoubanID: cells[2],
			})
		}
	}
	for _, c := range movie.Actor {
		cells := strings.Split(c.URL, "/")
		if len(cells) == 4 {
			media.Actor = append(media.Actor, &model.Person{
				Name:     c.Name,
				DoubanID: cells[2],
			})
		}
	}
	return media, err
}

func formatDate(dateStr string) *model.Date {
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

func decodeDuration(duration string) int {
	duration = strings.TrimSpace(duration)
	duration = strings.TrimPrefix(duration, "PT")
	duration = strings.TrimSuffix(duration, "M")
	cells := strings.Split(duration, "H")
	d := 0
	if len(cells) == 2 {
		h, _ := strconv.Atoi(cells[0])
		m, _ := strconv.Atoi(cells[1])
		d = h*60 + m
	}
	return d
}

type Celebrity struct {
	Name        string `json:"name"`
	ProfilePath string
	Gender      string `json:"gender"`
	Birthday    *model.Date
	BirthPlace  string
	Deathday    *model.Date
	Jobs        []string
	Alias       []string
	Biography   string
	IMDBID      string
}

func DumpCelebrity(celebrityID string) (*model.Person, error) {
	celebrity := new(model.Person)
	c := colly.NewCollector()
	cacheDir(c)
	c.SetCookieJar(jar)
	c.UserAgent = UserAgent

	// Find and visit all links
	c.OnHTML("#content>h1", func(e *colly.HTMLElement) {
		celebrity.Name = e.Text
	})
	c.OnHTML("#headline img", func(e *colly.HTMLElement) {
		celebrity.Profiles = []string{e.Attr("src")}
	})
	c.OnHTML("#intro>.bd", func(e *colly.HTMLElement) {
		celebrity.Biography = strings.TrimSpace(e.Text)
	})
	c.OnHTML("span.all.hidden", func(e *colly.HTMLElement) {
		celebrity.Biography = strings.TrimSpace(e.Text)
	})
	var zhAlias, enAlias []string
	c.OnHTML("#headline div.info>ul>li", func(e *colly.HTMLElement) {
		cells := strings.Split(strings.ReplaceAll(strings.ReplaceAll(e.Text, "\n", ""), " ", ""), ":")
		if len(cells) < 2 {
			return
		}
		switch cells[0] {
		case "性别":
			celebrity.Gender = cells[1]
		case "出生日期":
			celebrity.Birthday = parseDate(cells[1])
		case "生卒日期":
			dates := strings.Split(strings.ReplaceAll(cells[1], "+", ""), "至")
			if len(dates) >= 2 {
				celebrity.Birthday = parseDate(dates[0])
				celebrity.Deathday = parseDate(dates[1])
			}
		case "出生地":
			celebrity.PlaceOfBirth = cells[1]
		case "职业":
			celebrity.Jobs = strings.Split(cells[1], "/")
		case "更多外文名":
			enAlias = strings.Split(cells[1], "/")
		case "更多中文名":
			zhAlias = strings.Split(cells[1], "/")
		case "imdb编号":
			celebrity.IMDBID = cells[1]
		}
	})

	err := c.Visit(fmt.Sprintf("https://movie.douban.com/celebrity/%s/", celebrityID))
	if err != nil {
		return nil, err
	}
	celebrity.Alias = append(celebrity.Alias, zhAlias...)
	celebrity.Alias = append(celebrity.Alias, enAlias...)
	celebrity.DoubanID = celebrityID
	return celebrity, err
}

func parseDate(str string) *model.Date {
	rex, _ := regexp.Compile(`(\d{4})年((\d{2})月)*((\d{2})日)*`)
	results := rex.FindStringSubmatch(str)
	date := new(model.Date)
	if len(results) >= 2 {
		date.Year, _ = strconv.Atoi(results[1])
	}
	if len(results) >= 4 {
		date.Month, _ = strconv.Atoi(results[3])
	}
	if len(results) >= 6 {
		date.Day, _ = strconv.Atoi(results[5])
	}
	return date
}
