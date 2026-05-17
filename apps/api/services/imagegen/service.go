package imagegenvc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const defaultModel = "gpt-image-1"

// ErrImageGenFailed is returned when the OpenAI image generation call fails.
var ErrImageGenFailed = errors.New("image generation failed")

// Service generates images from text prompts.
type Service interface {
	GenerateImage(ctx context.Context, prompt string) ([]byte, error)
}

type service struct {
	client openai.Client
	model  string
}

// New returns a Service backed by the OpenAI Images API.
// If model is empty, defaultModel ("gpt-image-1") is used.
func New(apiKey, model string) Service {
	if model == "" {
		model = defaultModel
	}
	c := openai.NewClient(option.WithAPIKey(apiKey))
	return &service{client: c, model: model}
}

func (s *service) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	params := openai.ImageGenerateParams{
		Prompt: prompt,
		Model:  openai.ImageModel(s.model),
		N:      openai.Int(1),
		Size:   openai.ImageGenerateParamsSize1024x1024,
	}
	// dall-e-2 and dall-e-3 require explicit b64_json format; gpt-image-1+ always returns base64.
	if s.model == "dall-e-2" || s.model == "dall-e-3" {
		params.ResponseFormat = openai.ImageGenerateParamsResponseFormatB64JSON
	}

	resp, err := s.client.Images.Generate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrImageGenFailed, err)
	}
	if len(resp.Data) == 0 || resp.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("%w: empty response from API", ErrImageGenFailed)
	}

	data, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return data, nil
}
