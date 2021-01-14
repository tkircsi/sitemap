package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tkircsi/link"
)

type UrlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	Urls    []Url
}

type Url struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

func main() {
	urlFlag := flag.String("url", "http://gophercises.com", "the url that you want to build a sitemap for")
	maxDepth := flag.Int("depth", 5, "the maximum number of deep level to traverse")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)
	// for _, page := range pages {
	// 	fmt.Println(page)
	// }
	var toXml = UrlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Urls:  []Url{},
	}

	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, Url{
			Loc: page,
		})
	}

	f, err := os.Create("urls.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	xe := xml.NewEncoder(f)
	xe.Indent("", "  ")
	fmt.Fprint(f, xml.Header)
	err = xe.Encode(toXml)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(f)
}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i < maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		for url := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			for _, link := range get(url) {
				nq[link] = struct{}{}
			}
		}
	}
	ret := make([]string, len(seen))
	i := 0
	for k := range seen {
		ret[i] = k
		i++
	}
	return ret
}

func get(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	reqURL := resp.Request.URL
	baseURL := &url.URL{
		Scheme: reqURL.Scheme,
		Host:   reqURL.Host,
	}
	base := baseURL.String()

	return filter(hrefs(resp.Body, base), withPrefix(base))
}

func hrefs(r io.Reader, base string) []string {
	links, err := link.Parse(r)
	if err != nil {
		log.Fatal(err)
	}

	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}
	return ret
}

func filter(links []string, fn func(string) bool) []string {
	var ret []string
	for _, l := range links {
		if fn(l) {
			ret = append(ret, l)
		}
	}
	return ret
}

func withPrefix(prefix string) func(string) bool {
	return func(l string) bool {
		return strings.HasPrefix(l, prefix)
	}
}
