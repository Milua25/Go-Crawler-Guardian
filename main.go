package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// List of User Agents
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", randUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func discoverLinks(resp *http.Response, baseUrl string) []string {
	if resp != nil {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		foundUrls := []string{}

		if doc != nil {
			doc.Find("a").Each(func(i int, selection *goquery.Selection) {
				// For each item found, get the title
				res, _ := selection.Attr("href")
				foundUrls = append(foundUrls, res)
			})
		}
		return foundUrls
	} else {
		return []string{}
	}
}

func checkRelative(href, baseUrl string) string {
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseUrl, href)
	} else {
		return href
	}
}

func resolveRelativeLinks(href string, baseUrl string) (bool, string) {
	resultHref := checkRelative(href, baseUrl)
	baseParse, _ := url.Parse(baseUrl)
	resultParse, _ := url.Parse(resultHref)

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		} else {
			return false, ""
		}
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)

func Crawl(targetUrl, baseUrl string) []string {
	fmt.Println(targetUrl)
	tokens <- struct{}{}
	resp, _ := getRequest(targetUrl)
	<-tokens
	links := discoverLinks(resp, baseUrl)

	foundUrls := []string{}

	for _, link := range links {
		ok, correctLink := resolveRelativeLinks(link, baseUrl)
		if ok {
			if correctLink != "" {
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}
	//	ParseHTML(resp)
	return foundUrls
}

//func ()  {
//
//}

func randUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)

	return userAgents[randNum]
}

func main() {
	workList := make(chan []string)

	var n int
	n++

	baseDomain := "https://www.theguardian.com"

	// send the domain to the workList
	go func() {
		workList <- []string{"https://www.theguardian.com"}
	}()

	seen := make(map[string]bool)

	for ; n > 0; n-- {
		list := <-workList

		for _, link := range list {
			if !seen[link] {
				seen[link] = true
				n++
				go func(link, baseUrl string) {
					foundLinks := Crawl(link, baseDomain)
					if foundLinks != nil {
						workList <- foundLinks
					}
				}(link, baseDomain)
			}
		}
	}
}
