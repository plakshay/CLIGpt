package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// Gemini conversation roles.
const (
	RoleUser  = "user"
	RoleModel = "model"
)

// Message mirrors the role/content pair we keep in session history.
type Message struct {
	Role    string
	Content string
}

// Client wraps the Gemini SDK with the model we want to use.
type Client struct {
	api   *genai.Client
	model string
}

// New constructs a Client using the given API key and model.
func New(ctx context.Context, apiKey, model string) (*Client, error) {
	api, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &Client{api: api, model: model}, nil
}

// Stream sends the conversation and writes streamed response chunks to onChunk.
// It returns the full assistant reply so the caller can append it to history.
func (c *Client) Stream(ctx context.Context, systemPrompt string, history []Message, onChunk func(string)) (string, error) {
	cfg := &genai.GenerateContentConfig{}
	if systemPrompt != "" {
		cfg.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		}
	}

	var full string
	for resp, err := range c.api.Models.GenerateContentStream(ctx, c.model, toContents(history), cfg) {
		if err != nil {
			if ctx.Err() != nil {
				return full, ctx.Err()
			}
			return full, fmt.Errorf("stream: %w", err)
		}
		chunk := resp.Text()
		if chunk == "" {
			continue
		}
		full += chunk
		onChunk(chunk)
	}
	return full, nil
}

func toContents(history []Message) []*genai.Content {
	out := make([]*genai.Content, 0, len(history))
	for _, m := range history {
		out = append(out, &genai.Content{
			Role:  m.Role,
			Parts: []*genai.Part{{Text: m.Content}},
		})
	}
	return out
}