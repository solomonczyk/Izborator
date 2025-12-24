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
func (c *Client) GenerateSelectors(ctx context.Context, htmlSnippet string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert in CSS selectors and Web Scraping.
Analyze the provided HTML snippet of an e-commerce product page.
Identify the unique CSS selectors for the following fields:
1. Name (Product Title)
2. Price (Current price value)
3. Image (Main product image URL)
4. Description (Product details text)

Return ONLY a JSON object with keys: "name", "price", "image", "description".
Do not include markdown formatting.
If a field is not found, use null.

HTML Snippet:
%s`, htmlSnippet)

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
