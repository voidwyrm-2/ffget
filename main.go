package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/voidwyrm-2/ffget/api"
)

func SelectorFlag(name, value, usage string, selectFrom []string) struct{ get func() string } {
	ptr := flag.String(name, value, usage)
	return struct{ get func() string }{
		get: func() string {
			found := false
			flag.Visit(func(f *flag.Flag) {
				if f.Name == name {
					found = true
				}
			})
			if found && !slices.Contains(selectFrom, *ptr) && len(selectFrom) > 0 {
				fmt.Println(fmt.Sprintf("flag '-%s' only allows '%s'", name, func() string {
					if len(selectFrom) == 1 {
						return selectFrom[0]
					}
					return strings.Join(selectFrom[:len(selectFrom)-1], "', '") + "' or '" + selectFrom[len(selectFrom)-1] + "'"
				}()))
				os.Exit(1)
			}
			return *ptr
		},
	}
}

//go:embed version.txt
var version string

func _main() error {
	version = strings.TrimSpace(version)

	showVersion := flag.Bool("v", false, "Shows the current FFGet version")
	url := flag.String("u", "", "The URL to the fanfiction")
	showInfo := flag.Bool("i", false, "Gets the title, description, etc from the fanfiction")
	download := SelectorFlag("d", "none", "Download the specified format of the fanfiction", []string{"azw3", "epub", "mobi", "pdf", "html"})
	downloadOutput := flag.String("o", "", "The file to download to instead of [fic name].[download type]")

	flag.Parse()
	*url = strings.TrimSpace(*url)
	*downloadOutput = strings.TrimSpace(*downloadOutput)

	_ = download
	_ = downloadOutput

	if *showVersion {
		fmt.Println(version)
		return nil
	}

	if *url == "" {
		return errors.New("flag '-u' is required")
	}

	info, err := api.New(*url)
	if err != nil {
		return err
	}

	if *showInfo {
		return err
	}

	if download.get() != "none" {
		namec := ""
		for _, c := range info.Name {
			if c >= 0 && c <= 255 && c != ' ' && c != '\t' && c != '\n' {
				namec = string(c)
			} else {
				namec += "_"
			}
		}

		if strings.TrimSpace(*downloadOutput) == "" {
			*downloadOutput = namec
		}

		ficContent := []byte{}
		switch download.get() {
		case "azw3":
			ficContent, err = info.DownloadAzw3()
		case "epub":
			ficContent, err = info.DownloadEpub()
		case "mobi":
			ficContent, err = info.DownloadMobi()
		case "pdf":
			ficContent, err = info.DownloadPdf()
		case "html":
			ficContent, err = info.DownloadHtml()
		}

		if err != nil {
			return err
		}

		file, err := os.Create(*downloadOutput + "." + download.get())
		defer file.Close()
		if err != nil {
			return err
		}

		_, err = file.Write(ficContent)
		return err
	}

	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
