package openrouter

import (
	"context"
	"net/http"
)

const createImagesSuffix = "/images"

type ImageInputReference struct {
	ImageURL ImageURLRef `json:"image_url"`
}

type ImageURLRef struct {
	URL string `json:"url"`
}

type ImageGenerationProvider struct {
	Options map[string]map[string]any `json:"options,omitempty"`
}

type ImageGenerationRequest struct {
	Model             string                   `json:"model"`
	Prompt            string                   `json:"prompt"`
	AspectRatio       string                   `json:"aspect_ratio,omitempty"`
	Background        string                   `json:"background,omitempty"`
	InputReferences   []ImageInputReference    `json:"input_references,omitempty"`
	N                 *int                     `json:"n,omitempty"`
	OutputCompression *int                     `json:"output_compression,omitempty"`
	OutputFormat      string                   `json:"output_format,omitempty"`
	Provider          *ImageGenerationProvider `json:"provider,omitempty"`
	Quality           string                   `json:"quality,omitempty"`
	Resolution        string                   `json:"resolution,omitempty"`
	Seed              *int                     `json:"seed,omitempty"`
	Size              string                   `json:"size,omitempty"`
	Stream            *bool                    `json:"stream,omitempty"`
}

type ImageGenerationData struct {
	B64JSON   string `json:"b64_json"`
	MediaType string `json:"media_type,omitempty"`
}

type ImageGenerationUsage struct {
	CompletionTokens int      `json:"completion_tokens"`
	PromptTokens     int      `json:"prompt_tokens"`
	TotalTokens      int      `json:"total_tokens"`
	Cost             *float64 `json:"cost,omitempty"`
}

type ImageGenerationResponse struct {
	Created int64                 `json:"created"`
	Data    []ImageGenerationData `json:"data"`
	Usage   *ImageGenerationUsage `json:"usage,omitempty"`
}

// CreateImages generates an image from a text prompt via the image generation router.
// API reference: https://openrouter.ai/docs/api/api-reference/images/create-images
func (c *Client) CreateImages(
	ctx context.Context,
	request ImageGenerationRequest,
) (response ImageGenerationResponse, err error) {
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(createImagesSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
