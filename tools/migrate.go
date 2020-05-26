package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/xbapps/xbvr-scrapers/scrapers"
)

func main() {
	var scraper = flag.String("scraper", "", "path to scraper code")
	flag.Parse()

	scraperCode, err := ioutil.ReadFile(*scraper)
	if err != nil {
		panic(err)
	}

	def := scrapers.ScraperDefinition{}
	def.SiteOnhtml.SkipKnown = true
	def.SiteOnhtml.SkipURLContains = []string{"/join"}
	def.SiteOnhtml.VisitAttr = "href"
	def.PaginationOnhtml.VisitAttr = "href"

	reg := regexp.MustCompile(`registerScraper\(\"(?P<scraperId>.+)\", \"(?P<siteId>.+)\", \"(?P<scraperIcon>.+)\", (.+)\)`)
	match := reg.FindSubmatch(scraperCode)
	def.ScraperID = string(match[1])
	def.SiteID = string(match[2])
	def.SiteIcon = string(match[3])

	col := regexp.MustCompile(`createCollector\((\".+\")\)`)
	match = col.FindSubmatch(scraperCode)
	d := strings.Split(string(match[1]), ",")
	for i, v := range d {
		d[i] = strings.Trim(strings.TrimSpace(v), "\"")
	}

	def.AllowedDomains = d

	start := regexp.MustCompile(`siteCollector.Visit\(\"(.+)\"\)`)
	match = start.FindSubmatch(scraperCode)
	def.StartURL = string(match[1])

	studio := regexp.MustCompile(`sc.Studio = \"(.+)\"`)
	match = studio.FindSubmatch(scraperCode)
	def.Studio = string(match[1])

	so := regexp.MustCompile(`sceneCollector\.OnHTML\([\x60"](.+)[\x60"]`)
	match = so.FindSubmatch(scraperCode)
	def.SceneOnhtml.Selector = string(match[1])

	prettyJSON, err := json.MarshalIndent(def, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))
}
