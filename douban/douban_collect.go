package douban

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const DefaultCollectLink = "https://movie.douban.com/people/153634380/collect?start=%d&sort=time&rating=all&filter=all&mode=list"

func DumpCollect(before, current int) (subjectIds []string, total int, err error) {
	currentLink := fmt.Sprintf(DefaultCollectLink, current)
	var beforeLink string
	if before > -1 {
		beforeLink = fmt.Sprintf(DefaultCollectLink, current)
	}
	c := colly.NewCollector()
	c.SetCookieJar(jar)
	cacheDir(c)
	c.UserAgent = UserAgent

	c.OnHTML("div.title>a", func(e *colly.HTMLElement) {
		subjectIds = append(subjectIds, strings.Split(e.Attr("href"), "/")[4])
	})

	c.OnHTML("span.subject-num", func(e *colly.HTMLElement) {
		cells := strings.Split(e.Text, "/")
		if len(cells) == 2 {
			total, _ = strconv.Atoi(strings.TrimSpace(cells[1]))
		}
	})

	c.OnRequest(func(request *colly.Request) {
		if len(beforeLink) > 0 {
			request.Headers.Set("Referer", beforeLink)
		}
	})

	err = c.Visit(currentLink)
	if err != nil {
		return
	}
	return
}
