package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	priceInCent, err := parseCents(price)
	if err != nil {
		log.Fatalf("parsing euro string to cents: %v", err)
	}
	fmt.Printf("current price: %d.%02d EUR\n", priceInCent/100, priceInCent%100)
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

// isDigits reports whether s is non-empty and consists only of the ASCII digits 0-9.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// parseCents converts a price string in the form "479" or "479.99"
// into an amount in cents. It returns an error for any other format.
func parseCents(s string) (int64, error) {
	euroPart, centPart, hasDot := strings.Cut(s, ".")
	if !isDigits(euroPart) {
		return 0, fmt.Errorf("invalid price: %q", s)
	}
	parsedEuro, err := strconv.ParseInt(euroPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing euros in %q: %w", s, err)
	}
	switch {
	case !hasDot:
		centPart = "00"
	case len(centPart) == 1:
		centPart += "0"
	case len(centPart) == 2:
	default:
		// "479." or "479.999"
		return 0, fmt.Errorf("invalid price: %q", s)
	}
	if !isDigits(centPart) {
		return 0, fmt.Errorf("invalid price: %q", s)
	}
	parsedCent, err := strconv.ParseInt(centPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing cents in %q: %w", s, err)
	}
	return parsedEuro*100 + parsedCent, nil
}
