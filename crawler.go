package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

// URLSet list or urls
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// URL data from sitemap
type URL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
	News    News     `xml:"news"`
}

// News url detail data form sitemap
type News struct {
	XMLName         xml.Name `xml:"news"`
	PublicationDate string   `xml:"publication_date"`
	Title           string   `xml:"title"`
	Keywords        string   `xml:"keywords"`
}

// CrawlResult just url and stauts code
type CrawlResult struct {
	url        string
	statusCode int
}

var worker int = 10
var timeSleep int = 500

func openXMLFile(filename string) ([]byte, error) {
	xmlFile, err := os.Open(filename)
	if err != nil {
		return []byte{}, err
	}

	xmlData, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return []byte{}, err
	}

	return xmlData, nil
}

func parseXMLFile(filename string) (URLSet, error) {
	xmlData, err := openXMLFile(filename)
	if err != nil {
		return URLSet{}, err
	}

	var urlset URLSet
	xml.Unmarshal(xmlData, &urlset)
	return urlset, nil
}

func crawler(wg *sync.WaitGroup, tasks <-chan URL,
	results chan<- CrawlResult, instance int) {
	for url := range tasks {
		time.Sleep(time.Duration(timeSleep) * time.Millisecond)
		resp, err := http.Get(url.Loc)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("[worker %d ]: %s: %d\n", instance, url.Loc, resp.StatusCode)
		crawlresult := CrawlResult{
			url:        url.Loc,
			statusCode: resp.StatusCode,
		}
		results <- crawlresult
	}
	wg.Done()
}

func waitGroup(urlset URLSet) {
	var wg sync.WaitGroup
	tasks := make(chan URL, len(urlset.URLs)+5)
	results := make(chan CrawlResult, len(urlset.URLs)+5)

	// launch 5 worker
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go crawler(&wg, tasks, results, i+1)
	}

	// send tasks
	for _, url := range urlset.URLs {
		tasks <- url
	}
	close(tasks)

	wg.Wait()

	success := 0
	for i := 0; i < len(urlset.URLs); i++ {
		result := <-results
		fmt.Printf("[main] result %s: %d\n", result.url, result.statusCode)
		if result.statusCode == 200 {
			success++
		}
	}
	fmt.Printf("Total urls %d, success %d\n", len(urlset.URLs), success)
}

func crawlerWorker(tasks <-chan URL, results chan<- CrawlResult, instance int) {
	for url := range tasks {
		time.Sleep(time.Duration(timeSleep) * time.Millisecond)
		resp, err := http.Get(url.Loc)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("[worker %d ]: %s: %d\n", instance, url.Loc, resp.StatusCode)
		crawlresult := CrawlResult{
			url:        url.Loc,
			statusCode: resp.StatusCode,
		}
		results <- crawlresult
	}
}

func workerPool(urlset URLSet) {
	tasks := make(chan URL, len(urlset.URLs)+5)
	results := make(chan CrawlResult, len(urlset.URLs)+5)

	// run 5 worker
	for i := 0; i < worker; i++ {
		go crawlerWorker(tasks, results, i+1)
	}

	// send tasks
	for _, url := range urlset.URLs {
		tasks <- url
	}
	close(tasks)

	success := 0
	for i := 0; i < len(urlset.URLs); i++ {
		result := <-results
		fmt.Printf("[main] result %s: %d\n", result.url, result.statusCode)
		if result.statusCode == 200 {
			success++
		}
	}
	fmt.Printf("Total urls %d, success %d\n", len(urlset.URLs), success)
}

func main() {
	var filename = flag.String("filename", "sitemap-news.xml", "xml file")
	var method = flag.String("method", "wait-group", "method type wait-group|worker-pool")
	flag.Parse()

	urlset, err := parseXMLFile(*filename)
	if err != nil {
		fmt.Println(err.Error())
	}

	start := time.Now()

	switch {
	case *method == "wait-group":
		fmt.Printf("[main ] %s\n", *method)
		waitGroup(urlset)
	case *method == "worker-pool":
		fmt.Printf("[main] %s\n", *method)
		workerPool(urlset)
	}
	elapsed := time.Since(start)
	fmt.Printf("Wait Group took %s\n", elapsed)
}
