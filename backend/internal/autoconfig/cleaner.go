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

	// Удаляем мусор
	doc.Find("script, style, iframe, noscript, svg, footer, header, nav").Remove()

	// Очищаем комментарии (goquery их не парсит, но на всякий случай)
	// Основная очистка через удаление ненужных элементов

	// Для MVP просто берем body и обрезаем, если слишком длинно
	body := doc.Find("body")
	if body.Length() == 0 {
		// Если body нет, берем весь документ
		html, _ := doc.Html()
		return truncateHTML(html), nil
	}

	html, _ := body.Html()

	// Простое обрезание по длине (например, 20000 символов),
	// чтобы влезть в дешевый контекст. Обычно цена и имя в начале body.
	return truncateHTML(html), nil
}

// truncateHTML обрезает HTML до разумного размера
func truncateHTML(html string) string {
	if len(html) > 20000 {
		// Обрезаем, но стараемся не обрезать посередине тега
		truncated := html[:20000]
		// Находим последний закрывающий тег
		lastTag := strings.LastIndex(truncated, ">")
		if lastTag > 0 {
			truncated = truncated[:lastTag+1]
		}
		return truncated
	}
	return html
}
