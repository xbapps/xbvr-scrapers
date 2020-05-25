package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thoas/go-funk"
)

func sceneWriter(wg *sync.WaitGroup, i *uint64, scenes <-chan ScrapedScene) {
	defer wg.Done()

	for scene := range scenes {
		atomic.AddUint64(i, 1)
		fmt.Printf("\n\nScraped scene: %+v\n\n", scene)
	}
}

func main() {
	scrapersDir := "scrapers"
	files, err := ioutil.ReadDir(scrapersDir)
	if err != nil {
		log.Fatal(err)
	}
	t0 := time.Now()

	var scraperWG, writerWG sync.WaitGroup
	collectedScenes := make(chan ScrapedScene, 250)
	var sceneCount uint64

	writerWG.Add(1)
	go sceneWriter(&writerWG, &sceneCount, collectedScenes)

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

			scraperWG.Add(1)
			go Scrape(&scraperWG, configFile, parserFile, collectedScenes)
		} else {
			fmt.Printf("%s missing required files. Skipping.", scraperDir)
		}
	}
	// Wait for scrapers to complete
	scraperWG.Wait()

	// Wait for sceneWriter threads to complete
	close(collectedScenes)
	writerWG.Wait()

	fmt.Printf("Scraped %v scenes in %s\n",
		sceneCount,
		time.Now().Sub(t0).Round(time.Second))
}
