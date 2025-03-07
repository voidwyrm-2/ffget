package api

import (
	"strconv"
	"strings"
)

func CleanLink(url string) string {
	url = strings.TrimSpace(url)
	if i := strings.Index(url, "/chapter/"); i != -1 {
		url = url[:i]
	} else if i := strings.Index(url, "#"); i != -1 {
		url = url[:i]
	} else if i = strings.Index(url, "?"); i != -1 {
		url = url[:i]
	} else if _, err := strconv.Atoi(url); err == nil {
		url = "https://archiveofourown.org/works/" + url
	}

	url = strings.TrimSpace(url)
	url += "?hide_banner=true&amp;view_adult=true&amp;=true&amp;view_full_work=true"

	return url
}
