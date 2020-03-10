package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"net/url"

	"golang.org/x/net/html"
)

var rootURL string = "godoc.org"

type Node struct {
	url      string
	children map[*Node]struct{} //ToDo: convert to a simple slice
}

var sitemap Node
var siteLinks = map[string]struct{}{}
var mutexLinks = &sync.Mutex{}
var wg sync.WaitGroup

func main() {
	start := time.Now()
	sitemap.url = "/"
	sitemap.children = make(map[*Node]struct{})
	siteLinks = make(map[string]struct{})
	siteLinks[sitemap.url] = struct{}{}

	wg.Add(1)
	buildSitemap(sitemap)
	wg.Wait()

	fmt.Println("\nSitemap of", rootURL, ":")
	printSitemap()
	fmt.Println("Number of links:", len(siteLinks))
	fmt.Println("Time elapsed:", time.Since(start))
}

func buildSitemap(node Node) {
	// fmt.Println("\nIniciando analisis de", n.url)
	str := getHTMLFromURL(node.url)
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		log.Fatal(err)
	}

	// recursive path through the HTML nodes tree
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}
				a.Val = validateURL(a.Val)
				if a.Val == "" {
					continue
				}
				mutexLinks.Lock()
				_, exists := siteLinks[a.Val]
				if !exists {
					siteLinks[a.Val] = struct{}{} //update
					mutexLinks.Unlock()

					newNode := Node{url: a.Val, children: make(map[*Node]struct{})}
					node.children[&newNode] = struct{}{}

					wg.Add(1)
					go buildSitemap(newNode)
				} else {
					mutexLinks.Unlock()
				}

			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	rec(doc)

	wg.Done()
}

func getHTMLFromURL(URI string) string {
	res, err := http.Get("https://" + rootURL + URI)
	if err != nil {
		log.Fatal(err)
	}
	str, err := ioutil.ReadAll(res.Body) //dumps the body of the response into a string
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(str)
}

func printSitemap() {
	var recPath func(level int, node Node)
	recPath = func(level int, node Node) {
		for i := 0; i < level; i++ {
			fmt.Print("   ")
		}
		fmt.Println(node.url)
		if node.children != nil && len(node.children) > 0 {
			for pointer := range node.children {
				recPath(level+1, *pointer)
			}
		}
	}
	recPath(0, sitemap)
}

func validateURL(s string) (res string) {
	// fmt.Println(">Testing", s)
	u, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	// ignore invalid scheme, external links, empty strings and rootURL
	if (u.Scheme != "" && u.Scheme != "https" && u.Scheme != "http") ||
		(u.Host != "" && u.Host != rootURL) ||
		s == "" || s == "/" {
		return ""
	}
	res = u.Path
	if !strings.HasPrefix(res, "/") {
		res = "/" + res //format to be consistent
	}
	res = strings.TrimSuffix(res, "/")
	// fmt.Println("\t-->", res)
	return res
}
