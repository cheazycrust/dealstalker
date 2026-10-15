package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Please provide a URL of a product for price tracking")
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/spike-geizhals <geizhals-product-url>")
		os.Exit(1)
	}
	rawURL := os.Args[1]
	fmt.Printf("URL: %s\n", rawURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		log.Fatalf("creating request: %v", err)
	}

	req.Header.Set("User-Agent", "DealStalker/0.1 (+https://github.com/cheazycrust/dealstalker)")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	fmt.Println("status code:", resp.StatusCode)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("parsing html: %v", err)
	}

	price, ok := findMetaContent(doc, "og:price:amount")
	if !ok {
		log.Fatal("price not found on page")
	}
	fmt.Println("current price:", price)
}

// getAttr returns the value of the attribute named key on node n,
// or an empty string if n has no such attribute.
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// findMetaContent searches the tree rooted at n for a
// <meta property="..." content="..."> element whose property matches
// and returns its content value. The boolean reports whether a
// matching element was found.

func findMetaContent(n *html.Node, property string) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "meta" && getAttr(n, "property") == property {
		return getAttr(n, "content"), true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		val, ok := findMetaContent(c, property)
		if ok {
			return val, true
		}
	}

	return "", false
}
