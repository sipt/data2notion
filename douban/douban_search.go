package douban

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/sipt/data2notion/global"

	v8 "rogchap.com/v8go"
)

func SearchFromDouban(keyword string) (*DoubanSearchResponse, error) {
	cacheKey := fmt.Sprintf("douban_search_%s", keyword)
	result := &DoubanSearchResponse{}
	var err error
	if !global.WF.Cache.Expired(cacheKey, time.Hour*3) {
		if data, err := global.WF.Cache.Load(cacheKey); err == nil {
			if err := json.Unmarshal(data, result); err == nil {
				return result, nil
			}
		}
	}
	result, err = searchFromDoubanFromRemote(keyword)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result)
	err = global.WF.Cache.Store(cacheKey, data)
	if err != nil {
		log.Printf("ERROR: store cache failed[%s]: %s", cacheKey, err)
	}
	return result, nil
}

func searchFromDoubanFromRemote(keyword string) (*DoubanSearchResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://search.douban.com/movie/subject_search?search_text=%s", keyword))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rex, err := regexp.Compile(`[ ]*window.__DATA__[ ]*=[ ]*"(.*)"`)
	if err != nil {
		return nil, err
	}
	list := rex.FindSubmatch(html)

	ctx := v8.NewContext()
	val, err := ctx.RunScript(decodeJSSource+fmt.Sprintf(`de("%s")`, string(list[1])), "decode.go") // return a value in JavaScript back to Go
	if err != nil {
		return nil, err
	}
	data, err := val.MarshalJSON()
	if err != nil {
		return nil, err
	}

	searchResp := &DoubanSearchResponse{}
	err = json.Unmarshal(data, searchResp)
	if err != nil {
		return nil, err
	}
	return searchResp, err
}

type DoubanSearchResponse struct {
	Type    string `json:"type"`
	Payload struct {
		ErrorInfo string       `json:"error_info"`
		Text      string       `json:"text"`
		Start     int          `json:"start"`
		Total     int          `json:"total"`
		Items     []DoubanItem `json:"items"`
		Count     int          `json:"count"`
		Report    struct {
			Qtype string `json:"qtype"`
			Tags  string `json:"tags"`
		} `json:"report"`
	} `json:"payload"`
}
type DoubanItem struct {
	Rating struct {
		StarCount  float64 `json:"star_count"`
		Value      float32 `json:"value"`
		Count      int64   `json:"count"`
		RatingInfo string  `json:"rating_info"`
	} `json:"rating"`
	Title        string        `json:"title"`
	URL          string        `json:"url"`
	MoreURL      string        `json:"more_url"`
	Abstract     string        `json:"abstract"`
	Labels       []interface{} `json:"labels"`
	CoverURL     string        `json:"cover_url"`
	LabelActions []interface{} `json:"label_actions"`
	Abstract2    string        `json:"abstract_2"`
	Interest     interface{}   `json:"interest"`
	ExtraActions []interface{} `json:"extra_actions"`
	ID           int           `json:"id"`
	TplName      string        `json:"tpl_name"`
	Topics       []interface{} `json:"topics"`
}
