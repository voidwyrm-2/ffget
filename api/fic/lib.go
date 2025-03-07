package fic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pejman-hkh/gdp/gdp"
)

func download(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return []byte{}, err
	} else if string(content) == "404: Not Found" {
		return []byte{}, errors.New("404: Not Found")
	}

	return content, nil
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

	info.Downloads.Azw3 = links["AZW3"]
	info.Downloads.Epub = links["EPUB"]
	info.Downloads.Mobi = links["MOBI"]
	info.Downloads.Pdf = links["PDF"]
	info.Downloads.Html = links["HTML"]

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

	info.Categorization.Rating = stats["rating"]
	info.Categorization.ArchiveWarning = statLists["warning"]
	info.Categorization.Categories = statLists["category"]
	info.Categorization.Fandoms = statLists["fandom"]
	info.Categorization.Relationships = statLists["relationships"]
	info.Categorization.Characters = statLists["character"]
	info.Categorization.Tags = statLists["freeform"]

	info.Language = stats["language"]

	info.Stats.Published = stats["published"]
	info.Stats.Status = stats["status"]

	statConv := func(stat string) (int, error) {
		s := strings.Split(stat, ":")

		n, err := strconv.Atoi(strings.ReplaceAll(strings.TrimSpace(s[len(s)-1]), ",", ""))
		if err != nil {
			if err.Error() == "strconv.Atoi: parsing \"\": invalid syntax" {
				return 0, nil
			} else if err.Error() == "strconv.Atoi: parsing \"?\": invalid syntax" {
				return -1, nil
			}
		}

		return n, err
	}

	var err error

	info.Stats.Words, err = statConv(stats["words"])
	if err != nil {
		return err
	}

	{
		chapters := strings.Split(stats["chapters"], "/")

		info.Stats.Chapters.Current, err = statConv(chapters[0])
		if err != nil {
			return err
		}

		info.Stats.Chapters.Current, err = statConv(chapters[1])
		if err != nil {
			return err
		}
	}

	info.Stats.Comments, err = statConv(stats["comments"])
	if err != nil {
		return err
	}

	info.Stats.Kudos, err = statConv(stats["kudos"])
	if err != nil {
		return err
	}

	info.Stats.Bookmarks, err = statConv(stats["bookmarks"])
	if err != nil {
		return err
	}

	info.Stats.Hits, err = statConv(stats["hits"])
	if err != nil {
		return err
	}

	return nil
}
