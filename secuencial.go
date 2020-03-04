package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"net/url"

	"golang.org/x/net/html"
)

var rootURL string = "godoc.com"

type Node struct {
	url      string
	children map[*Node]struct{}
}

var sitemap Node
var siteLinks = map[string]struct{}{}

func main() {
	start := time.Now()
	sitemap.url = "/"
	sitemap.children = make(map[*Node]struct{})
	siteLinks = make(map[string]struct{})
	siteLinks[sitemap.url] = struct{}{}

	buildSitemap(sitemap)
	fmt.Println("\nSitemap of", rootURL, ":")
	fmt.Println("Number of links:", len(siteLinks))
	printSitemap()
	fmt.Println("Time elapsed:", time.Since(start))
}

func buildSitemap(n Node) {
	fmt.Println("\nIniciando analisis de", n.url)
	links := getLinksFromUrl(n.url)
	for l := range links {
		newNode := Node{url: l}
		newNode.children = make(map[*Node]struct{})
		n.children[&newNode] = struct{}{}

		// update used links
		siteLinks[l] = struct{}{}

	}
	for child := range n.children {
		buildSitemap(*child)
	}

}

func getLinksFromUrl(current_url string) map[string]struct{} {
	//get html
	res, err := http.Get("http://" + rootURL + current_url)
	if err != nil {
		log.Fatal(err)
	}
	// ioutil.ReadAll dumps the body of the response into a string
	str, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	// str contains the full body, now we have to strip all links

	links := map[string]struct{}{}

	// recursive path through the HTML nodes tree
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					a.Val = validateURL(a.Val)
					if a.Val == "" {
						continue
					} else if _, ok := siteLinks[a.Val]; !ok {
						// a.Val es posible candidato pero me tengo que fijar que no lo haya agregado ya
						links[a.Val] = struct{}{}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	doc, err := html.Parse(strings.NewReader(string(str)))
	if err != nil {
		log.Fatal(err)
	}
	f(doc)

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
	fmt.Println(">Testing", s)
	u, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	// ignore mailto, external links, empty strings and the rooturl
	if u.Scheme == "mailto" ||
		(u.Host != "" && u.Host != rootURL) ||
		s == "" || s == "/" {
		fmt.Println("  Ignorando", u.Host, s)
		return ""
	}
	res = u.Path
	if !strings.HasPrefix(res, "/") {
		res = "/" + res //agrego para que sea consistente
	}
	res = strings.TrimSuffix(res, "/")
	fmt.Println("-->", res)
	return res
}
