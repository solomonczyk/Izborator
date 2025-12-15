package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <URL>")
	}
	url := os.Args[1]

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	c.SetRequestTimeout(60)

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "sr-RS,sr;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		fmt.Printf("H1 found: %s\n", strings.TrimSpace(e.Text))
		fmt.Printf("  Classes: %s\n", e.Attr("class"))
	})

	c.OnHTML(".price", func(e *colly.HTMLElement) {
		fmt.Printf("Price (.price): %s\n", strings.TrimSpace(e.Text))
	})

	c.OnHTML("[data-price-type]", func(e *colly.HTMLElement) {
		fmt.Printf("Price (data-price-type): %s, type: %s\n", strings.TrimSpace(e.Text), e.Attr("data-price-type"))
	})

	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		fmt.Printf("JSON-LD found (length: %d)\n", len(e.Text))
		if strings.Contains(e.Text, "price") {
			fmt.Printf("  Contains 'price': YES\n")
			// Ищем цену в JSON
			idx := strings.Index(e.Text, `"price"`)
			if idx > 0 {
				start := idx - 50
				if start < 0 {
					start = 0
				}
				end := idx + 100
				if end > len(e.Text) {
					end = len(e.Text)
				}
				fmt.Printf("  Context: %s\n", e.Text[start:end])
			}
		}
	})

	c.OnResponse(func(r *colly.Response) {
		html := string(r.Body)
		fmt.Printf("\n=== HTML Analysis ===\n")
		fmt.Printf("Status: %d\n", r.StatusCode)
		fmt.Printf("HTML Length: %d\n", len(html))

		// Ищем название
		if strings.Contains(html, "Dell Laptop XPS") {
			fmt.Printf("Found 'Dell Laptop XPS' in HTML\n")
			idx := strings.Index(html, "Dell Laptop XPS")
			start := idx - 100
			if start < 0 {
				start = 0
			}
			end := idx + 200
			if end > len(html) {
				end = len(html)
			}
			fmt.Printf("Context: %s\n", html[start:end])
		}

		// Ищем цену
		if strings.Contains(html, "RSD") || strings.Contains(html, "din") {
			fmt.Printf("Found price indicators (RSD/din)\n")
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error: %v (Status: %d)\n", err, r.StatusCode)
	})

	fmt.Printf("Fetching: %s\n\n", url)
	if err := c.Visit(url); err != nil {
		log.Fatal(err)
	}
}

