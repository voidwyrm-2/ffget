package api

import (
	"fmt"
	"strings"

	"github.com/pejman-hkh/gdp/gdp"
)

type FFInfo struct {
	Url, Name, Author, Summary string
	FicCating                  struct {
		Rating                                                               string
		ArchiveWarning, Categories, Fandoms, Relationships, Characters, Tags []string
	}
	Language  string
	Stats     struct{ Published, Status, Words, Chapters, Comments, Kudos, Bookmarks, Hits string }
	Downloads struct{ Azw3, Epub, Mobi, Pdf, Html string }
}

func New(url string) (FFInfo, error) {
	url = CleanLink(url)

	html, err := getHTML(url)
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

func (info FFInfo) DownloadAzw3() ([]byte, error) {
	return getHTML(info.Downloads.Azw3)
}

func (info FFInfo) DownloadEpub() ([]byte, error) {
	return getHTML(info.Downloads.Epub)
}

func (info FFInfo) DownloadMobi() ([]byte, error) {
	return getHTML(info.Downloads.Mobi)
}

func (info FFInfo) DownloadPdf() ([]byte, error) {
	return getHTML(info.Downloads.Pdf)
}

func (info FFInfo) DownloadHtml() ([]byte, error) {
	return getHTML(info.Downloads.Html)
}

func (info FFInfo) FormatStats() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		info.Stats.Published,
		info.Stats.Status,
		info.Stats.Words,
		info.Stats.Chapters,
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
		strings.Split(info.Url, "?")[0],
		info.Name,
		info.Author,
		strings.Join(info.FicCating.Fandoms, "', '"),
		info.FicCating.Rating,
		strings.Join(info.FicCating.Categories, ", "),
		strings.Join(info.FicCating.Tags, "',\t\t\n'"),
		info.FormatStats(),
		info.Summary,
	)
}
