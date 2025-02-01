package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/pejman-hkh/gdp/gdp"
)

func writeFile(filename string, data string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func assert[T any](v T, _ error) T {
	return v
}

func getHTML(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	} else if string(content) == "404: Not Found" {
		return "", errors.New("404: Not Found")
	}

	return string(content), nil
}

func cleanFFLink(url string) string {
	url = strings.TrimSpace(url)
	if i := strings.Index(url, "/chapter/"); i != -1 {
		url = url[:i]
	} else if i := strings.Index(url, "#"); i != -1 {
		url = url[:i]
	} else if i = strings.Index(url, "?"); i != -1 {
		url = url[:i]
	} else if r, _ := regexp.Compile(`[0-9]*`); string(r.Find([]byte(url))) == url {
		url = "https://archiveofourown.org/works/" + url
	}

	url = strings.TrimSpace(url)
	url += "?hide_banner=true&amp;view_adult=true&amp;=true&amp;view_full_work=true"

	return url
}

type FFInfo struct {
	name, author, summary string
	ficCating             struct {
		rating                                                               string
		archiveWarning, categories, fandoms, relationships, characters, tags []string
	}
	language  string
	stats     struct{ published, status, words, chapters, comments, kudos, bookmarks, hits string }
	downloads struct{ azw3, epub, mobi, pdf, html string }
}

func getDownloads(ff gdp.Tag, info *FFInfo) error {
	var (
		dmenu, download *gdp.Tag = &gdp.Tag{}, &gdp.Tag{}
		links                    = map[string]string{}
		ulCount                  = 0
	)
	ff.Find("body").Eq(0).GetElementById("outer").GetElementById("inner").GetElementById("main").GetChildren().Each(func(_ int, t *gdp.Tag) {
		if t.Name == "ul" {
			if ulCount == 1 {
				dmenu = t
			}
			ulCount++
		}
	})

	dmenu.GetChildren().Each(func(_ int, t *gdp.Tag) {
		if t.Name == "li" && t.HasClass("download") {
			download = t
		}
	})

	dsecondary := download.Find("ul").Eq(0)

	dsecondary.GetChildren().Each(func(_ int, t *gdp.Tag) {
		if t.Name == "li" {
			a := t.Find("a").Eq(0)
			links[a.Children[0].Content] = "https://archiveofourown.org" + a.Attr("href")
		}
	})

	info.downloads.azw3 = links["AZW3"]
	info.downloads.epub = links["EPUB"]
	info.downloads.mobi = links["MOBI"]
	info.downloads.pdf = links["PDF"]
	info.downloads.html = links["HTML"]

	return nil
}

func getStats(ff gdp.Tag, info *FFInfo) error {
	var (
		wrapper   = &gdp.Tag{}
		stats     = map[string]string{}
		statLists = map[string][]string{}
	)

	_, _ = stats, statLists

	ff.Find("body").Eq(0).GetElementById("outer").GetElementById("inner").GetElementById("main").GetChildren().Each(func(i int, t *gdp.Tag) {
		if t.Name == "div" && t.HasClass("wrapper") {
			wrapper = t
		}
	})

	wrapper.Find("dl").Eq(0).GetChildren().Each(func(_ int, t *gdp.Tag) {
		if t.Name == "dd" {
			if t.HasClass("tags") {
				kind := strings.Split(t.Attr("class"), " ")[0]
				switch kind {
				case "rating":
					stats[kind] = t.Find("ul").Eq(0).Find("li").Eq(0).Find("a").Eq(0).Children[0].Content
				case "warning":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], ftag)
						}
					})
				case "category":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], ftag)
						}
					})
				case "fandom":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], ftag)
						}
					})
				case "relationship":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], ftag)
						}
					})
				case "character":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], ftag)
						}
					})
				case "freeform":
					statLists[kind] = []string{}
					t.Find("ul").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
						if ftag := strings.TrimSpace(subt.Find("a").Eq(0).Html()); ftag != "" {
							statLists[kind] = append(statLists[kind], func(s string, old, new []string) string {
								for i := range old {
									s = strings.ReplaceAll(s, old[i], new[i])
								}
								return s
							}(ftag, []string{"&#39;"}, []string{"'"}))
						}
					})
				default:
					panic(fmt.Sprintf("invalid kind '%s'", kind))
				}
			} else if t.HasClass("language") {
				stats["language"] = t.Children[0].Content
			} else if t.HasClass("stats") {
				lastdt := ""
				t.Find("dl").Eq(0).GetChildren().Each(func(_ int, subt *gdp.Tag) {
					if subt.Name == "dt" {
						lastdt = subt.Html()
					} else if subt.Name == "dd" {
						if subt.HasClass("bookmarks") {
							stats[subt.Attr("class")] = lastdt + " " + subt.Find("a").Eq(0).Html()
						} else {
							stats[subt.Attr("class")] = lastdt + " " + subt.Html()
						}
					}
				})
			}
		}
	})

	info.ficCating.rating = stats["rating"]
	info.ficCating.archiveWarning = statLists["warning"]
	info.ficCating.categories = statLists["category"]
	info.ficCating.fandoms = statLists["fandom"]
	info.ficCating.relationships = statLists["relationships"]
	info.ficCating.characters = statLists["character"]
	info.ficCating.tags = statLists["freeform"]

	info.language = stats["language"]

	info.stats.published = stats["published"]
	info.stats.status = stats["status"]
	info.stats.words = stats["words"]
	info.stats.chapters = stats["chapters"]
	info.stats.comments = stats["comments"]
	info.stats.kudos = stats["kudos"]
	info.stats.bookmarks = stats["bookmarks"]
	info.stats.hits = stats["hits"]

	return nil
}

func getFFInfo(ffhtml string) (FFInfo, error) {
	info := FFInfo{}

	ff := gdp.Default(ffhtml)

	{
		preface := ff.Find("body").Eq(0).GetElementById("outer").GetElementById("inner").GetElementById("main").GetElementById("workskin").Find("div").Eq(0)
		info.name = strings.TrimSpace(preface.Find("h2").Eq(0).Html())
		info.author = strings.TrimSpace(preface.Find("h3").Eq(0).Find("a").Eq(0).Html())
		sumpara := []string{}
		preface.Find("div").Eq(0).Find("blockquote").Eq(0).GetChildren().Each(func(_ int, t *gdp.Tag) {
			sumpara = append(sumpara, t.Html())
		})
		info.summary = strings.TrimSpace(strings.Join(sumpara, "\n\n"))
	}

	if err := getDownloads(ff, &info); err != nil {
		return FFInfo{}, err
	}

	if err := getStats(ff, &info); err != nil {
		return FFInfo{}, err
	}

	return info, nil
}
