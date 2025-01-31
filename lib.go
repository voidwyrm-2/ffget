package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	titleRaw                                                         *string
	name, author, fandom, rating, category, tags, stats, description string
	downloads                                                        struct{ azw3, epub, mobi, pdf, html string }
}

func getFFInfo(ffhtml string) (FFInfo, error) {
	info := FFInfo{}
	info.titleRaw = nil

	// title, auther, fandom
	{
		titleReg, err := regexp.Compile(`<meta name="viewport" content="width=device-width, initial-scale=[0-9.]*"/>\n *<title>`)
		if err != nil {
			return FFInfo{}, err
		}

		titlem := titleReg.FindIndex([]byte(ffhtml))
		if titlem == nil {
			// return FFInfo{}, errors.New("could not find fanfiction title")
			info.name, info.author, info.fandom = "[NOT FOUND]", "[NOT FOUND]", "[NOT FOUND]"
		} else {
			titleRaw := strings.TrimSpace(strings.Split(ffhtml[titlem[1]:], "</title>")[0])
			titlespl := strings.Split(titleRaw, " - ")
			if len(titlespl) > 4 {
				info.titleRaw = &titleRaw
			} else {
				for i, t := range titlespl {
					if i == 0 {
						info.name = strings.TrimSpace(t)
					} else if i == 1 {
						info.author = strings.TrimSpace(t)
					} else if i == 2 && len(titlespl) == 4 {
						continue
					} else {
						if strings.HasSuffix(strings.ToLower(t), "[archive of our own]") {
							info.fandom = strings.TrimSpace(t[:len(t)-20])
						} else {
							info.fandom = strings.TrimSpace(t)
						}
						break
					}
				}
			}
		}
	}

	// rating
	{
		ratingReg, err := regexp.Compile(`<dd class="rating tags">\n *<ul class="commas">\n *<li><a class="tag" href="\/tags\/[a-zA-Z0-9%]*\/works">[a-zA-Z ]*<\/a><\/li>`)
		if err != nil {
			return FFInfo{}, err
		}

		ratingm := ratingReg.FindIndex([]byte(ffhtml))
		if ratingm == nil {
			info.rating = "[NOT FOUND]"
		} else {
			subratReg, err := regexp.Compile(`>[a-zA-Z ][a-zA-Z ]*<`)
			if err != nil {
				return FFInfo{}, err
			}
			srm := subratReg.FindIndex([]byte(ffhtml[ratingm[0]:ratingm[1]]))
			if srm == nil {
				panic("could not find rating info in `" + ffhtml[ratingm[0]:ratingm[1]] + "`")
			}

			ratingRaw := ffhtml[ratingm[0]:ratingm[1]][srm[0]+1 : srm[1]]

			info.rating = ratingRaw[:len(ratingRaw)-1]
		}
	}

	{
		r, _ := regexp.Compile(`<dd class="freeform tags">\n *<ul class="commas">\n *<li><a class="tag" href="/tags/.*/works">.*</a></li>`)
		r2, _ := regexp.Compile(` *<a class="tag" href="/tags/.*/works">`)
		s := string(r.Find([]byte(ffhtml)))
		s = s[:len(s)-9]
		s = strings.Join(strings.Split(s, "<li>")[1:], "<li>")

		tags := []string{}
		for _, t := range strings.Split(s, "</a></li><li>") {
			t = strings.TrimSpace(t[len(r2.FindString(t)):])
			if t != "" {
				tags = append(tags, strings.ReplaceAll(t, "&#39;", "'"))
			}
		}

		info.tags = "`" + strings.Join(tags, "`, `") + "`"
	}

	// stats
	{
		statsReg, err := regexp.Compile(`<dl class="stats"><dt class="published">Published:<\/dt><dd class="published">[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]<\/dd><dt class="status">(Completed|Updated):<\/dt><dd class="status">[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]<\/dd><dt class="words">Words:<\/dt><dd class="words">[0-9]*,?[0-9]*<\/dd><dt class="chapters">Chapters:<\/dt><dd class="chapters">[0-9]*\/([0-9]*|\?)<\/dd><dt class="comments">Comments:<\/dt><dd class="comments">[0-9]*,?[0-9]*<\/dd><dt class="kudos">Kudos:<\/dt><dd class="kudos">[0-9]*,?[0-9]*<\/dd><dt class="bookmarks">Bookmarks:<\/dt><dd class="bookmarks"><a href="\/works\/[0-9]*\/bookmarks">[0-9]*,?[0-9]*<\/a><\/dd><dt class="hits">Hits:<\/dt><dd class="hits">[0-9]*,?[0-9]*`)
		if err != nil {
			return FFInfo{}, err
		}
		indiSReg := assert(regexp.Compile(`class="(published|status|words|chapters|comments|kudos|hits)">`))

		statsm := statsReg.FindIndex([]byte(ffhtml))
		if statsm == nil {
			info.stats = "[NOT FOUND]"
		} else {
			statsRaw := []string{}
			secs := "[ERROR]"

			for _, stat := range strings.Split(strings.ReplaceAll(ffhtml[statsm[0]:statsm[1]], "</dd><dt ", "</dt><dd "), "</dt><dd ")[1:] {
				if !strings.Contains(stat, ":") && indiSReg.Find([]byte(stat)) != nil {
					statsRaw = append(statsRaw, stat)
				} else if m := assert(regexp.Compile("(Completed|Updated)")).Find([]byte(stat)); m != nil {
					secs = strings.ToLower(string(m))
				}
			}

			_ = secs

			for i := range statsRaw {
				t := string(indiSReg.Find([]byte(statsRaw[i])))[7:]
				t = t[:len(t)-2]
				statsRaw[i] = t
			}

			info.stats = strings.Join(statsRaw, "\n")
		}
	}

	// description
	{
		descReg, err := regexp.Compile(`<blockquote class="userstuff">\n *<p>`)
		if err != nil {
			return FFInfo{}, err
		}

		descm := descReg.FindIndex([]byte(ffhtml))
		if descm == nil {
			info.description = "[NOT FOUND]"
		} else {
			info.description = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.Split(ffhtml[descm[1]:], "</blockquote>")[0]), "<p>", "\n"), "</p>", "")
		}
	}

	// downloads
	{
		downloadsReg, err := regexp.Compile(`<li class="download" aria-haspopup="true">\n *<a href="#">Download</a>\n *<ul class="expandable secondary">\n *<li><a href="/downloads/[0-9]*/[0-9a-zA-Z_&;%\?]*.azw3\?updated_at=[0-9]*">AZW3</a></li>\n *<li><a href="/downloads/[0-9]*/[0-9a-zA-Z_&;%\?]*.epub\?updated_at=[0-9]*">EPUB</a></li>\n *<li><a href="/downloads/[0-9]*/[0-9a-zA-Z_&;%\?]*.mobi\?updated_at=[0-9]*">MOBI</a></li>\n *<li><a href="/downloads/[0-9]*/[0-9a-zA-Z_&;%\?]*.pdf\?updated_at=[0-9]*">PDF</a></li>\n *<li><a href="/downloads/[0-9]*/[0-9a-zA-Z_&;%\?]*.html\?updated_at=[0-9]*">HTML</a></li>\n *</ul>\n *</li>`)
		if err != nil {
			return FFInfo{}, err
		}

		downloadsm := downloadsReg.FindIndex([]byte(ffhtml))
		if downloadsm == nil {
			info.downloads = struct {
				azw3 string
				epub string
				mobi string
				pdf  string
				html string
			}{
				azw3: "[NOT FOUND]",
				epub: "[NOT FOUND]",
				mobi: "[NOT FOUND]",
				pdf:  "[NOT FOUND]",
				html: "[NOT FOUND]",
			}
		} else {
			downspl := strings.Split(strings.TrimSpace(ffhtml[downloadsm[0]:downloadsm[1]]), "\n")

			azw3 := strings.TrimSpace(downspl[3])[13:]
			epub := strings.TrimSpace(downspl[4])[13:]
			mobi := strings.TrimSpace(downspl[5])[13:]
			pdf := strings.TrimSpace(downspl[6])[13:]
			html := strings.TrimSpace(downspl[7])[13:]

			info.downloads = struct {
				azw3 string
				epub string
				mobi string
				pdf  string
				html string
			}{
				azw3: "https://archiveofourown.org" + azw3[:len(azw3)-15],
				epub: "https://archiveofourown.org" + epub[:len(epub)-15],
				mobi: "https://archiveofourown.org" + mobi[:len(mobi)-15],
				pdf:  "https://archiveofourown.org" + pdf[:len(pdf)-14],
				html: "https://archiveofourown.org" + html[:len(html)-15],
			}
		}
	}

	return info, nil
}
