package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/gocolly/colly/v2"
	"github.com/mozillazg/go-slugify"
)

var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36"

type ScraperDefinition struct {
	ScraperID      string   `json:"scraper_id"`
	SiteID         string   `json:"site_id"`
	Studio         string   `json:"studio"`
	SiteIcon       string   `json:"site_icon"`
	AllowedDomains []string `json:"allowed_domains"`
	StartURL       string   `json:"start_url"`
	SiteOnhtml     struct {
		Selector        string   `json:"selector"`
		VisitAttr       string   `json:"visit_attr"`
		SkipKnown       bool     `json:"skip_known"`
		SkipURLContains []string `json:"skip_url_contains"`
	} `json:"site_onhtml"`
	PaginationOnhtml struct {
		Selector  string `json:"selector"`
		VisitAttr string `json:"visit_attr"`
		SkipKnown bool   `json:"skip_known"`
	} `json:"pagination_onhtml"`
	SceneOnhtml struct {
		Selector   string `json:"selector"`
		NeededVars []struct {
			VarName     string   `json:"var_name"`
			CollyMethod string   `json:"colly_method"`
			CollyArgs   []string `json:"colly_args"`
		} `json:"needed_vars"`
	} `json:"scene_onhtml"`
}

type ScrapedScene struct {
	SceneID     string   `json:"_id"`
	SiteID      string   `json:"scene_id"`
	SceneType   string   `json:"scene_type"`
	Title       string   `json:"title"`
	Studio      string   `json:"studio"`
	Site        string   `json:"site"`
	Covers      []string `json:"covers"`
	Gallery     []string `json:"gallery"`
	Tags        []string `json:"tags"`
	Cast        []string `json:"cast"`
	Filenames   []string `json:"filename"`
	Duration    int      `json:"duration"`
	Synopsis    string   `json:"synopsis"`
	Released    string   `json:"released"`
	HomepageURL string   `json:"homepage_url"`
}

func arrayToInterface(a []string) []interface{} {
	var i []interface{}
	for _, s := range a {
		i = append(i, s)
	}
	return i
}

func interfaceToArray(i []interface{}) []string {
	var s []string
	for _, x := range i {
		s = append(s, strings.TrimSpace(x.(string)))
	}
	return s
}

func createCollector(domains ...string) *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains(domains...),
		colly.CacheDir("./scraper_cache"),
		colly.UserAgent(userAgent),
	)

	c = createCallbacks(c)
	return c
}

func cloneCollector(c *colly.Collector) *colly.Collector {
	x := c.Clone()
	x = createCallbacks(x)
	return x
}

func createCallbacks(c *colly.Collector) *colly.Collector {
	const maxRetries = 15

	c.OnRequest(func(r *colly.Request) {
		attempt := r.Ctx.GetAny("attempt")

		if attempt == nil {
			r.Ctx.Put("attempt", 1)
		}

		fmt.Println("visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		attempt := r.Ctx.GetAny("attempt").(int)

		if r.StatusCode == 429 {
			fmt.Println("Error:", r.StatusCode, err)

			if attempt <= maxRetries {
				unCache(r.Request.URL.String(), c.CacheDir)
				fmt.Println("Waiting 2 seconds before next request...")
				r.Ctx.Put("attempt", attempt+1)
				time.Sleep(2 * time.Second)
				r.Request.Retry()
			}
		}
	})

	return c
}

func unCache(URL string, cacheDir string) {
	sum := sha1.Sum([]byte(URL))
	hash := hex.EncodeToString(sum[:])
	dir := path.Join(cacheDir, hash[:2])
	filename := path.Join(dir, hash)
	if err := os.Remove(filename); err != nil {
		fmt.Println(err)
	}
}

func Scrape(configFile string, parserFile string) {
	scraperConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var scraper ScraperDefinition
	json.Unmarshal(scraperConfig, &scraper)

	sceneCollector := createCollector(scraper.AllowedDomains...)
	siteCollector := createCollector(scraper.AllowedDomains...)

	scraperParser, err := ioutil.ReadFile(parserFile)
	if err != nil {
		panic(err)
	}
	script := tengo.NewScript(scraperParser)
	script.SetImports(stdlib.GetModuleMap("fmt", "text", "times"))

	sceneCollector.OnHTML(scraper.SceneOnhtml.Selector, func(e *colly.HTMLElement) {
		scene := ScrapedScene{}
		for _, v := range scraper.SceneOnhtml.NeededVars {
			switch v.CollyMethod {
			case "ChildText":
				val := e.ChildText(v.CollyArgs[0])
				_ = script.Add(v.VarName, val)
			case "ChildTexts":
				val := e.ChildTexts(v.CollyArgs[0])
				_ = script.Add(v.VarName, arrayToInterface(val))
			case "ChildAttr":
				val := e.ChildAttr(v.CollyArgs[0], v.CollyArgs[1])
				_ = script.Add(v.VarName, val)
			case "ChildAttrs":
				val := e.ChildAttrs(v.CollyArgs[0], v.CollyArgs[1])
				_ = script.Add(v.VarName, arrayToInterface(val))
			}
		}
		_ = script.Add("url", e.Request.URL.String())
		// run the script
		parsed, err := script.RunContext(context.Background())
		if err != nil {
			panic(err)
		}

		// retrieve values
		scene.SceneType = "VR"
		scene.Studio = scraper.Studio
		scene.Site = scraper.SiteID
		scene.SiteID = strings.TrimSpace(parsed.Get("siteID").String())
		scene.SceneID = slugify.Slugify(scene.Site + "-" + scene.SiteID)
		scene.HomepageURL = strings.TrimSpace(parsed.Get("homepageURL").String())
		scene.Title = strings.TrimSpace(parsed.Get("title").String())
		scene.Duration = parsed.Get("duration").Int()
		scene.Synopsis = strings.TrimSpace(parsed.Get("synopsis").String())
		scene.Covers = append(scene.Covers, strings.TrimSpace(parsed.Get("coverURL").String()))
		scene.Gallery = interfaceToArray(parsed.Get("galleryURLS").Array())
		scene.Cast = interfaceToArray(parsed.Get("cast").Array())

		fmt.Printf("\n\nScraped scene: %+v\n\n", scene)
	})

	siteCollector.OnHTML(scraper.SiteOnhtml.Selector, func(e *colly.HTMLElement) {
		u := e.Request.AbsoluteURL(e.Attr("href"))
		shouldVisit := true
		if scraper.SiteOnhtml.SkipURLContains != nil {
			for _, s := range scraper.SiteOnhtml.SkipURLContains {
				if strings.Contains(u, s) {
					shouldVisit = false
					break
				}
			}
		}
		if shouldVisit {
			sceneCollector.Visit(u)
		}
	})

	siteCollector.OnHTML(scraper.PaginationOnhtml.Selector, func(e *colly.HTMLElement) {
		u := e.Request.AbsoluteURL(e.Attr("href"))
		siteCollector.Visit(u)
	})

	siteCollector.Visit(scraper.StartURL)
}
