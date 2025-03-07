package search

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pejman-hkh/gdp/gdp"
	"github.com/voidwyrm-2/ffget/api/fic"
)

func Parse(url string) ([]fic.FFInfo, error) {
	ff := []fic.FFInfo{}

	content, err := download(url)
	if err != nil {
		return []fic.FFInfo{}, err
	}

	first := gdp.Default(string(content))

	second := first.Find("body").Eq(0).GetElementById("outer").GetElementById("inner").GetElementById("main")

	ficEntries := second.Find("ol").Eq(1).GetChildren()

	check, err := regexp.Compile(`work_[0-9]+`)
	if err != nil {
		return []fic.FFInfo{}, err
	}

	for ti := range ficEntries.Length() {
		id := ficEntries.Eq(ti).Attr("id")
		if check.MatchString(id) {
			f, err := fic.New("https://archiveofourown.org/works/" + strings.Split(id, "_")[1])
			if err != nil {
				return []fic.FFInfo{}, err
			}

			ff = append(ff, f)
		}

		fmt.Printf("%d of %d\n", ti+1, ficEntries.Length())
	}

	return ff, nil
}
