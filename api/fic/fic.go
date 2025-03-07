package fic

import (
	"fmt"
	"strings"

	"github.com/pejman-hkh/gdp/gdp"
	"github.com/voidwyrm-2/ffget/api"
)

type FFInfo struct {
	Url, Name, Author, Summary string
	Categorization             struct {
		Rating                                                               string
		ArchiveWarning, Categories, Fandoms, Relationships, Characters, Tags []string
	}
	Language string
	Stats    struct {
		Chapters struct {
			Current, Planned int
		}
		Published, Status                       string
		Words, Comments, Kudos, Bookmarks, Hits int
	}
	Downloads struct{ Azw3, Epub, Mobi, Pdf, Html string }
}

func New(url string) (FFInfo, error) {
	url = api.CleanLink(url)

	html, err := download(url)
	if err != nil {
		return FFInfo{}, err
	}

	return NewFromHTML(url, string(html))
}

func NewFromHTML(url, html string) (FFInfo, error) {
	info := FFInfo{Url: url}

	ff := gdp.Default(html)

	{
		preface := ff.Find("body").Eq(0).GetElementById("outer").GetElementById("inner").GetElementById("main").GetElementById("workskin").Find("div").Eq(0)
		info.Name = strings.TrimSpace(preface.Find("h2").Eq(0).Html())
		info.Author = strings.TrimSpace(preface.Find("h3").Eq(0).Find("a").Eq(0).Html())
		sumpara := []string{}
		preface.Find("div").Eq(0).Find("blockquote").Eq(0).GetChildren().Each(func(_ int, t *gdp.Tag) {
			sumpara = append(sumpara, t.Html())
		})
		info.Summary = strings.TrimSpace(strings.Join(sumpara, "\n\n"))
	}

	if err := getDownloads(ff, &info); err != nil {
		return FFInfo{}, err
	}

	if err := getStats(ff, &info); err != nil {
		return FFInfo{}, err
	}

	return info, nil
}

/*
Returns the name of the fanfiction cleaned for use in file systems
*/
func (info FFInfo) NameForFS() string {
	name := ""

	for _, c := range info.Name {
		if c < 0 || c > 255 || c == '\t' || c == '\n' || c == '/' || c == '\\' || c == ';' || c == ':' {
			name += "_"
		} else {
			name += string(c)
		}
	}

	return name
}

func (info FFInfo) DownloadAzw3() ([]byte, error) {
	return download(info.Downloads.Azw3)
}

func (info FFInfo) DownloadEpub() ([]byte, error) {
	return download(info.Downloads.Epub)
}

func (info FFInfo) DownloadMobi() ([]byte, error) {
	return download(info.Downloads.Mobi)
}

func (info FFInfo) DownloadPdf() ([]byte, error) {
	return download(info.Downloads.Pdf)
}

func (info FFInfo) DownloadHtml() ([]byte, error) {
	return download(info.Downloads.Html)
}

func (info FFInfo) FormatStats() string {
	return fmt.Sprintf("Published: %s\nStatus: %s\nWords: %d\nChapters: %d/%d\nComments: %d\nKudos: %d\nBookmarks: %d\nHits: %d",
		info.Stats.Published,
		info.Stats.Status,
		info.Stats.Words,
		info.Stats.Chapters.Current,
		info.Stats.Chapters.Planned,
		info.Stats.Comments,
		info.Stats.Kudos,
		info.Stats.Bookmarks,
		info.Stats.Hits,
	)
}

func (info FFInfo) String() string {
	return fmt.Sprintf(`from link: %s
		
==fic info==
name: %s
author: %s
fandoms: '%s'

rating: %s
categories: %s
============

==tags==
'%s'
========

==stats==
%s
=========

==description==
%s
===============`+"\n",
		info.Url,
		info.Name,
		info.Author,
		strings.Join(info.Categorization.Fandoms, "', '"),
		info.Categorization.Rating,
		strings.Join(info.Categorization.Categories, ", "),
		strings.Join(info.Categorization.Tags, "',\t\t\n'"),
		info.FormatStats(),
		info.Summary,
	)
}
