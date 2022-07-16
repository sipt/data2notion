package model

import "fmt"

type Media struct {
	Title             string
	Posters           []string
	TMDBID            int
	DoubanID          string
	IMDBID            string
	Genres            []string
	ProductionCountry string
	ReleaseDate       *Date
	RuntimeMin        int // 分钟
	RatingCount       int
	Score             int
	Tagline           string
	Overview          string
	Alias             []string
	Season            int
	EpisodeCount      int
	EpisodeDuration   int
	Language          string
	Director          []*Person
	Author            []*Person
	Actor             []*Person

	Status string
	Tags   string
}

type Date struct {
	Year  int
	Month int
	Day   int
}

func (d *Date) String() string {
	str := ""
	if d == nil {
		return str
	}
	if d.Year > 0 {
		str += fmt.Sprintf("%d年", d.Year)
	}
	if d.Month > 0 {
		str += fmt.Sprintf("%d月", d.Month)
	}
	if d.Day > 0 {
		str += fmt.Sprintf("%d日", d.Day)
	}
	return str
}

type Person struct {
	Name         string
	Profiles     []string
	Gender       string
	Birthday     *Date
	PlaceOfBirth string
	Deathday     *Date
	Alias        []string
	Biography    string
	Jobs         []string
	TMDBID       int
	DoubanID     string
	IMDBID       string
}
