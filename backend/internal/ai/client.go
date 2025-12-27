package ai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Client обертка для работы с OpenAI API
type Client struct {
	api   *openai.Client
	model string
}

// New создает новый AI клиент
func New(token string, model string) *Client {
	if model == "" {
		model = openai.GPT4oMini // Дешевая и умная модель
	}
	return &Client{
		api:   openai.NewClient(token),
		model: model,
	}
}

// SelectorsResult результат генерации селекторов
type SelectorsResult struct {
	Name        string `json:"name"`
	Price       string `json:"price"`
	Image       string `json:"image"`
	Description string `json:"description"`
}

// GenerateSelectors просит AI найти селекторы в HTML
// siteType может быть "ecommerce" или "service_provider"
func (c *Client) GenerateSelectors(ctx context.Context, htmlSnippet string, siteType string) (string, error) {
	var pageTypeDesc string
	var extractionInstructions string
	
	if siteType == "service_provider" {
		pageTypeDesc = "service provider price list page (cenovnik, прайс-лист)"
		extractionInstructions = `For service providers, the page may contain:
- HTML tables (<table>) with service names and prices in rows
- Price lists (cenovnik) with structured data
- Service names in one column, prices in another column
- Multiple services on one page
- Div-based lists with service items (not just tables)

CRITICAL: For table-based price lists, you MUST provide selectors that can extract MULTIPLE rows.

Extract selectors for:
1. Name: Service name selector that works for EACH row/item
   - If in table: "table tr td:first-child" or "table tbody tr td:nth-child(1)"
   - If in div list: "div.service-item" or "div.price-item" or similar
   - MUST select individual service names, not the entire table
2. Price: Service price selector that works for EACH row/item
   - If in table: "table tr td:last-child" or "table tbody tr td:nth-child(2)" or column with price
   - If in div list: "div.service-item .price" or similar
   - MUST select individual prices, not the entire table
3. Image: Service image or provider logo (optional)
   - Can be empty if no images available
4. Description: Service description or details (optional)
   - Can be empty if no descriptions available

EXAMPLES of good selectors for tables:
- Name: "table tbody tr td:first-child" (selects first cell of each row)
- Price: "table tbody tr td:last-child" (selects last cell of each row)
- Or: "table tbody tr td:nth-child(1)" and "table tbody tr td:nth-child(2)"

EXAMPLES for div-based lists:
- Name: "div.service-list > div .service-name"
- Price: "div.service-list > div .service-price"

IMPORTANT: The selectors MUST allow extracting multiple services from the same page.`
	} else {
		pageTypeDesc = "e-commerce product page"
		extractionInstructions = `For e-commerce pages, extract:
1. Name: Product Title
2. Price: Current price value
3. Image: Main product image URL
4. Description: Product details text`
	}

	var nameFieldDesc string
	if siteType == "service_provider" {
		nameFieldDesc = "Service name"
	} else {
		nameFieldDesc = "Product Title"
	}

	prompt := fmt.Sprintf(`You are an expert in CSS selectors and Web Scraping.
Analyze the provided HTML snippet of a %s.
%s

CRITICAL: You must return CSS SELECTORS (like ".product-title" or "h1.name"), NOT the actual data values!

Identify the unique CSS selectors for the following fields:
1. Name (%s) - CSS selector that matches the element containing the name (e.g., "h1.product-title", ".product-name", "table td:first-child")
2. Price (Current price value) - CSS selector that matches the element containing the price (e.g., ".price", ".product-price", "table td:last-child")
3. Image (Main product/service image URL, optional) - CSS selector for image element (e.g., "img.product-image", "img[src]")
4. Description (Product/service details text, optional) - CSS selector for description element

IMPORTANT:
- Return CSS SELECTORS (strings like ".class-name", "#id", "div > span"), NOT the actual text content
- Example CORRECT: {"name": "h1.product-title", "price": ".price-value"}
- Example WRONG: {"name": "iPhone 15 Pro", "price": "129999 RSD"} <- This is data, not a selector!

CRITICAL for service providers (price lists):
- The page contains MULTIPLE services, not just one
- If data is in a table (<table>), provide selectors for TABLE CELLS (td), not rows (tr)
- Example: "table tbody tr td:first-child" for name, "table tbody tr td:last-child" for price
- These selectors will match ALL rows, allowing extraction of all services
- If data is in divs, provide selectors that match each service item individually
- The scraper will iterate over all matches, so each selector should target ONE field per service

Return ONLY a JSON object with keys: "name", "price", "image", "description".
Each value must be a CSS selector string (like ".class" or "#id"), NOT the actual content.
Do not include markdown formatting.
If a field is not found, use null or empty string.

HTML Snippet:
%s`, pageTypeDesc, extractionInstructions, nameFieldDesc, htmlSnippet)

	resp, err := c.api.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1, // Нам нужна точность, а не креативность
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	// Очистка от markdown, если AI все-таки его добавил
	result := resp.Choices[0].Message.Content
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSpace(result)

	return result, nil
}
