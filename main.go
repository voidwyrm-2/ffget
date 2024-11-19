package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

//go:embed version.txt
var version string

func main() {
	version = strings.TrimSpace(version)
	parser := argparse.NewParser("ffget", "An AO3 client written in Go")

	version := parser.Flag("v", "version", &argparse.Options{Required: false, Default: false, Help: "Shows the current FFGet version"})
	url := parser.String("u", "url", &argparse.Options{Required: true, Help: "The URL to the fanfiction"})
	info := parser.Flag("i", "info", &argparse.Options{Required: false, Default: false, Help: "Gets the title, description, etc from the fanfiction"})
	download := parser.Selector("d", "download", []string{"azw3", "epub", "mobi", "pdf", "html"}, &argparse.Options{Required: false, Default: "none", Help: "Download the specified format of the fanfiction"})
	downloadOutput := parser.String("o", "output", &argparse.Options{Required: false, Default: "", Help: "The file to download to instead of [fic name].[download type]"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	if *version {
		fmt.Println(version)
		return
	}

	html, err := getFFRaw(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ffinfo, err := getFFInfo(html)
	if err != nil {
		fmt.Println(err.Error())
	}

	if *info {
		if ffinfo.titleRaw != nil {
			fmt.Printf("raw title: %s\ndescription:\n%s\n", *ffinfo.titleRaw, ffinfo.description)
		} else {
			fmt.Printf("name: %s\nauthor: %s\nfandom: %s\ndescription:\n%s\n", ffinfo.name, ffinfo.author, ffinfo.fandom, ffinfo.description)
		}
		return
	}

	if *download != "none" {
		namec := []byte{}
		for _, c := range ffinfo.name {
			if c >= 0 && c <= 255 && c != ' ' && c != '\t' && c != '\n' {
				namec = append(namec, byte(c))
			} else {
				namec = append(namec, '_')
			}
		}

		if strings.TrimSpace(*downloadOutput) == "" {
			*downloadOutput = string(namec)
		}

		downURL := func() (l string) {
			switch *download {
			case "azw3":
				l = ffinfo.downloads.azw3
			case "epub":
				l = ffinfo.downloads.epub
			case "mobi":
				l = ffinfo.downloads.mobi
			case "pdf":
				l = ffinfo.downloads.pdf
			case "html":
				l = ffinfo.downloads.html
			}
			return
		}()

		downloadedContent, err := getHTML(downURL)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if err := writeFile(*downloadOutput+"."+*download, downloadedContent); err != nil {
			fmt.Println(err.Error())
			return
		}

		/*
			fmt.Println("downloading '" + downURL + "' to '" + *downloadOutput + "'...")
			com := []string{"-c", "curl '" + downURL + "' -o " + *downloadOutput + "." + *download}
			// fmt.Println(com)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd := exec.Command("sh", com...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(stderr.String())
				fmt.Println(err.Error())
			} else {
				fmt.Println("downloading finished")
			}
		*/
	}
}
