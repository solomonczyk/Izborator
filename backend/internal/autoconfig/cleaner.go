package autoconfig

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CleanHTML удаляет скрипты, стили и лишние атрибуты, оставляя структуру
func CleanHTML(rawHTML string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return "", err
	}

	doc.Find("script, style, iframe, noscript, svg, footer, header, nav").Remove()

	body := doc.Find("body")
	var text string
	if body.Length() == 0 {
		text = doc.Text()
	} else {
		text = body.Text()
	}
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	return truncateHTML(text), nil
}


// truncateHTML обрезает HTML до разумного размера
func truncateHTML(html string) string {
	if len(html) > 20000 {
		return html[:20000]
	}
	return html
}

