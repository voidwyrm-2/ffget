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

func getFFRaw(url *string) (string, error) {
	if i := strings.Index(*url, "?"); i != -1 {
		*url = (*url)[:i]
	}
	*url = strings.TrimSpace(*url)
	*url += "?hide_banner=true&amp;view_adult=true&amp;=true&amp;view_full_work=true"

	return getHTML(*url)
}

type FFInfo struct {
	titleRaw                          *string
	name, author, fandom, description string
	downloads                         struct{ azw3, epub, mobi, pdf, html string }
}

func getFFInfo(ffhtml string) (FFInfo, error) {
	info := FFInfo{}
	info.titleRaw = nil

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

	{
		descReg, err := regexp.Compile(`<blockquote class="(userstuff|userstuff summary)">\n *<p>`)
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
