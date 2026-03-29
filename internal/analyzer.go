package analyzer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type AnalysisResult struct {
	URL           string
	HTMLVersion   string
	Title         string
	Headings      map[string]int
	InternalLinks int
	ExternalLinks int
	Inaccessible  int
	HasLoginForm  bool
}

func Analyze(ctx context.Context, targetURL string) (*AnalysisResult, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("URL is not reachable: %v", err)
	}
	defer resp.Body.Close()

	log.Println("Checking the submitted URL for errors and analyzing the content")
	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	rootNode, err := html.Parse(resp.Body)
	if err != nil {
		log.Println("Failed to parse HTML structure")
		return nil, fmt.Errorf("failed to parse HTML structure")
	}

	doc := goquery.NewDocumentFromNode(rootNode)

	log.Println("Extracting information from the page")
	res := &AnalysisResult{
		URL:         targetURL,
		HTMLVersion: detectHTMLVersion(rootNode),
		Title:       doc.Find("title").Text(),
		Headings:    make(map[string]int),
	}

	log.Println("Counting headings and links")
	for i := 1; i <= 6; i++ {
		tag := fmt.Sprintf("h%d", i)
		res.Headings[tag] = doc.Find(tag).Length()
	}

	log.Println("Analyzing links on the page")
	var extLinks []string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.HasPrefix(href, "http") {
			res.ExternalLinks++
			extLinks = append(extLinks, href)
		} else if href != "" && !strings.HasPrefix(href, "#") {
			res.InternalLinks++
		}
	})

	log.Println("Checking accessibility of external links")
	res.Inaccessible = checkAccessibility(ctx, extLinks)

	log.Println("Checking for login form")
	res.HasLoginForm = doc.Find("form input[type='password']").Length() > 0

	log.Println("Analysis complete, returning results")
	return res, nil
}

func detectHTMLVersion(n *html.Node) string {
	for c := n.FirstChild; c != nil; c = n.NextSibling {
		if c.Type == html.DoctypeNode {
			if strings.ToLower(c.Data) == "html" && len(c.Attr) == 0 {
				return "HTML5"
			}
			for _, a := range c.Attr {
				if a.Key == "public" {
					return a.Val
				}
			}
		}
		if res := detectHTMLVersion(c); res != "Unknown" {
			return res
		}
	}
	log.Println("Could not detect HTML version, defaulting to Unknown")
	return "Unknown"
}

func checkAccessibility(ctx context.Context, links []string) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	failedCount := 0
	client := &http.Client{Timeout: 5 * time.Second}

	for _, link := range links {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "HEAD", l, nil)
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				mu.Lock()
				failedCount++
				mu.Unlock()
			}
		}(link)
	}
	wg.Wait()
	log.Println("Completed accessibility checks for external links")
	return failedCount
}
