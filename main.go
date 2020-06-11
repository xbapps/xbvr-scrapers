package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr-scrapers/scrapers"
)

func sceneWriter(wg *sync.WaitGroup, i *uint64, scenes <-chan scrapers.ScrapedScene) {
	defer wg.Done()

	for scene := range scenes {
		atomic.AddUint64(i, 1)
		fmt.Printf("\n\nScraped scene: %+v\n\n", scene)
	}
}

func scrapeDir(dir string, scraperWG *sync.WaitGroup, collectedScenes chan scrapers.ScrapedScene) {
	sfs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var scraperFiles []string
	for _, sf := range sfs {
		scraperFiles = append(scraperFiles, sf.Name())
	}
	fmt.Println(scraperFiles)

	_, scraperName := filepath.Split(dir)
	if funk.Contains(scraperFiles, "config.json") && funk.Contains(scraperFiles, scraperName+".tengo") {
		configFile := filepath.Join(dir, "config.json")
		parserFile := filepath.Join(dir, scraperName+".tengo")
		fmt.Printf("Scraping %s...\n", dir)
		fmt.Println(configFile, parserFile)

		scraperWG.Add(1)
		go scrapers.Scrape(scraperWG, configFile, parserFile, collectedScenes)
	} else {
		fmt.Printf("%s missing required files. Skipping.\n", dir)
	}
}

func main() {
	var scraper = flag.String("scraper", "", "process a single specified scraper")
	flag.Parse()

	scrapersDir := "scrapers"
	files, err := ioutil.ReadDir(scrapersDir)
	if err != nil {
		log.Fatal(err)
	}
	t0 := time.Now()

	var scraperWG, writerWG sync.WaitGroup
	collectedScenes := make(chan scrapers.ScrapedScene, 250)
	var sceneCount uint64

	writerWG.Add(1)
	go sceneWriter(&writerWG, &sceneCount, collectedScenes)

	if *scraper != "" {
		scraperDir := filepath.Join(scrapersDir, *scraper)
		scrapeDir(scraperDir, &scraperWG, collectedScenes)
	} else {
		for _, f := range files {
			scraperDir := filepath.Join(scrapersDir, f.Name())

			if info, err := os.Stat(scraperDir); err == nil && !info.IsDir() {
				continue
			}

			scrapeDir(scraperDir, &scraperWG, collectedScenes)
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
