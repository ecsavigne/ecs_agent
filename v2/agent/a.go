package agent

import (
	"context"

	"github.com/ecsavigne/ecs_agent/config"
	"github.com/ecsavigne/ecs_agent/error_ia"
	internal_config "github.com/ecsavigne/ecs_agent/internal/config"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

type LLM interface {
	Reset()
	// GetBase returns the base of the LLM
	GetBase() *base
	GetWelcomeMessage() string
	SetWelcomeMessage(string)
	SetTpl(string, ...bool) *base
	Format(map[string]any) (string, error)
	Ask(string, config.FuncStream, ...bool) string
	NewLLMChain(prompts.FormatPrompter, ...chains.ChainCallOption) *chains.LLMChain
	Run(context.Context, chains.Chain, any, ...chains.ChainCallOption) (string, error)
	Call(context.Context, chains.Chain, map[string]any, ...chains.ChainCallOption) (map[string]any, error)
}

// New returns a new instance of the LLM given by typeModel and c.
// typeModel must be one of config.DEEKSEEK or config.GEMINI.
// c must be a config.ConfigModel.
// If typeModel is not recognized, New panics with "Model not found".
func New(typeModel config.TYPE_MODEL, c ...config.ConfigMod) LLM {
	// if c.APIKey == "" {
	// 	panic(error_ia.ErrorApiKeyNotSet)
	// }
	switch typeModel {
	case config.DEEKSEEK:
		c = append(c, config.WithModel("deepseek-chat"))
		c = append(c, internal_config.WithTypeAgent(config.DEEKSEEK))
		return newDeekSeek(c...)
	case config.GEMINI:
		return newGemini(c...)
	case config.NEW_GAI:
		c = append(c, internal_config.WithTypeAgent(config.NEW_GAI))
		return newgaiNew(c...)
	default:
		panic(error_ia.ErrorAgentNotFound)
	}
}
