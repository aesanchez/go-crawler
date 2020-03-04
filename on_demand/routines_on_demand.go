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

var rootURL string = "southerncrossbritannia.com"

type Node struct {
	url      string
	children map[*Node]struct{} //ToDo: convert to a simple slice
}

var sitemap Node
var siteLinks = map[string]struct{}{}
var mutexLinks = &sync.Mutex{}
var mutexRunning = &sync.Mutex{}
var routinesRunning = 0
var mutexLast = &sync.Mutex{} //ToDo: semaphore or waiting gruops

func main() {
	start := time.Now()
	sitemap.url = "/"
	sitemap.children = make(map[*Node]struct{})
	siteLinks = make(map[string]struct{})
	siteLinks[sitemap.url] = struct{}{}

	mutexLast.Lock()
	routinesRunning = 1
	buildSitemap(sitemap)
	mutexLast.Lock()
	fmt.Println("\nSitemap of", rootURL, ":")
	printSitemap()
	fmt.Println("Number of links:", len(siteLinks))
	fmt.Println("Time elapsed:", time.Since(start))
}

func buildSitemap(n Node) {
	// fmt.Println("\nIniciando analisis de", n.url)
	links := getLinksFromURL(n.url)
	for l := range links {
		newNode := Node{url: l}
		newNode.children = make(map[*Node]struct{})
		n.children[&newNode] = struct{}{}
		mutexRunning.Lock()
		routinesRunning++
		fmt.Println("+", routinesRunning)
		mutexRunning.Unlock()
		go buildSitemap(newNode)
	}
	// leave
	mutexRunning.Lock()
	routinesRunning--
	if routinesRunning == 0 {
		mutexLast.Unlock()
	}
	fmt.Println("-", routinesRunning)
	mutexRunning.Unlock()
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

func getLinksFromURL(URI string) map[string]struct{} {
	links := map[string]struct{}{}
	str := getHTMLFromURL(URI)
	// recursive path through the HTML nodes tree
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					a.Val = validateURL(a.Val)
					if a.Val == "" {
						continue
					}
					mutexLinks.Lock()
					if _, ok := siteLinks[a.Val]; !ok { //check if it was processed before
						siteLinks[a.Val] = struct{}{} //update
						links[a.Val] = struct{}{}
					}
					mutexLinks.Unlock()

				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		log.Fatal(err)
	}
	rec(doc)

	return links
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
