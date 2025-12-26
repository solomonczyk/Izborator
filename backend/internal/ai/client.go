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

Extract selectors for:
1. Name: Service name (может быть в таблице: table tr td:first-child или отдельный элемент)
2. Price: Service price (может быть в таблице: table tr td:last-child или рядом с названием)
3. Image: Service image or provider logo (optional)
4. Description: Service description or details (optional)

For table-based price lists, use selectors like:
- "table tr" for rows (then extract name from first td, price from last td)
- Or specific table cell selectors if structure is consistent`
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

Identify the unique CSS selectors for the following fields:
1. Name (%s)
2. Price (Current price value)
3. Image (Main product/service image URL, optional)
4. Description (Product/service details text, optional)

IMPORTANT for service providers:
- If data is in a table, provide selectors that can extract multiple rows
- Use table row selectors (e.g., "table tr") if services are listed in rows
- Price may be in a separate column or cell
- For table-based extraction, you can use: "table tr" for rows, then extract name from first td, price from last td

Return ONLY a JSON object with keys: "name", "price", "image", "description".
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
