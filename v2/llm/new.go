package llm

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/genai"
)

type GoogleAI struct {
	CallbacksHandler callbacks.Handler
	opts             Options
	model            string // Track current model for reasoning detection
	client           *genai.Client
}

var (
	_ llms.Model          = (*GoogleAI)(nil)
	_ llms.ReasoningModel = (*GoogleAI)(nil)
)

func New(ctx context.Context, opts ...Option) (*GoogleAI, error) {
	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	gai := &GoogleAI{
		opts:  clientOptions,
		model: clientOptions.DefaultModel,
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: gai.opts.clientOpts.APIKey,
	})
	if err != nil {
		return nil, err
	}

	gai.client = client

	return gai, nil
}
