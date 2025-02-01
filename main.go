package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

func exprin(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

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
				exprin(fmt.Sprintf("flag '-%s' only allows '%s'", name, func() string {
					if len(selectFrom) == 1 {
						return selectFrom[0]
					}
					return strings.Join(selectFrom[:len(selectFrom)-1], "', '") + "' or '" + selectFrom[len(selectFrom)-1] + "'"
				}()))
			}
			return *ptr
		},
	}
}

//go:embed version.txt
var version string

var debug *bool

func dprint(a ...any) {
	if *debug {
		fmt.Println(a...)
	}
}

func main() {
	version = strings.TrimSpace(version)
	/*
		parser := argparse.NewParser("ffget", "An AO3 client written in Go")

		showVersion := parser.Flag("v", "version", &argparse.Options{Required: false, Default: false, Help: "Shows the current FFGet version"})
		url := parser.String("u", "url", &argparse.Options{Required: true, Help: "The URL to the fanfiction"})
		info := parser.Flag("i", "info", &argparse.Options{Required: false, Default: false, Help: "Gets the title, description, etc from the fanfiction"})
		download := parser.Selector("d", "download", []string{"azw3", "epub", "mobi", "pdf", "html"}, &argparse.Options{Required: false, Default: "none", Help: "Download the specified format of the fanfiction"})
		downloadOutput := parser.String("o", "output", &argparse.Options{Required: false, Default: "", Help: "The file to download to instead of [fic name].[download type]"})

		err := parser.Parse(os.Args)
		if err != nil {
			fmt.Print(parser.Usage(err))
			return
		}
	*/

	debug = flag.Bool("debug", false, "Prints debug messages")
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
		return
	}

	if *url == "" {
		fmt.Println("flag '-u' is required")
		os.Exit(1)
	}

	*url = cleanFFLink(*url)

	html, err := getHTML(*url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ffinfo, err := getFFInfo(html)
	if err != nil {
		fmt.Println(err.Error())
	}

	if *showInfo {
		formattedStats := fmt.Sprintf(`%s
%s
%s
%s
%s
%s
%s
%s`, ffinfo.stats.published,
			ffinfo.stats.status,
			ffinfo.stats.words,
			ffinfo.stats.chapters,
			ffinfo.stats.comments,
			ffinfo.stats.kudos,
			ffinfo.stats.bookmarks,
			ffinfo.stats.hits)

		fmt.Printf(`from link: %s
		
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
===============`+"\n", strings.Split(*url, "?")[0], ffinfo.name, ffinfo.author, strings.Join(ffinfo.ficCating.fandoms, "', '"), ffinfo.ficCating.rating, strings.Join(ffinfo.ficCating.categories, ", "), strings.Join(ffinfo.ficCating.tags, "',\t\t\n'"), formattedStats, ffinfo.summary)
		return
	}

	if download.get() != "none" {
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
			switch download.get() {
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

		if err := writeFile(*downloadOutput+"."+download.get(), downloadedContent); err != nil {
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
