package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/thoas/go-funk"
)

func main() {
	scrapersDir := "scrapers"
	files, err := ioutil.ReadDir(scrapersDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		scraperDir := filepath.Join(scrapersDir, f.Name())

		sfs, err := ioutil.ReadDir(scraperDir)
		if err != nil {
			log.Fatal(err)
		}
		var scraperFiles []string
		for _, sf := range sfs {
			scraperFiles = append(scraperFiles, sf.Name())
		}

		if funk.Contains(scraperFiles, "config.json") && funk.Contains(scraperFiles, f.Name()+".tengo") {
			configFile := filepath.Join(scraperDir, "config.json")
			parserFile := filepath.Join(scraperDir, f.Name()+".tengo")
			fmt.Printf("Scraping %s...", f.Name())
			Scrape(configFile, parserFile)
		} else {
			fmt.Printf("%s missing required files. Skipping.", scraperDir)
		}
	}
}
