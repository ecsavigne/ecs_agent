package agent

import (
	"context"
	"log"

	_llm "github.com/ecsavigne/ecs_agent/v2/llm"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

func geminiLLM(apiKey, model string) *googleai.GoogleAI {
	llm, e := googleai.New(
		context.Background(),
		googleai.WithDefaultModel(model),
		googleai.WithAPIKey(apiKey),
	)

	if e != nil {
		log.Fatalf("Error create llm gemini, error is: %v", e)
	}

	return llm
}

func deekseekLLM(apiKey, model string) *openai.LLM {
	llm, e := openai.New(
		openai.WithModel(model),
		openai.WithToken(apiKey),
		openai.WithBaseURL("https://api.deepseek.com"),
		// openai.WithResponseFormat(openai.ResponseFormatJSON),
	)
	if e != nil {
		log.Fatalf("Error create llm deekseek, error is: %v", e)
	}

	return llm
}

func newGaiLLM(apiKey, model string) *_llm.GoogleAI {
	llm, e := _llm.New(
		context.Background(),
		_llm.WithDefaultModel(model),
		_llm.WithAPIKey(apiKey),
	)

	if e != nil {
		log.Fatalf("Error create llm newGai, error is: %v", e)
	}

	return llm
}
