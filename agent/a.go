package agent

import (
	"github.com/ecsavigne/ecs_agent/config"
	"github.com/ecsavigne/ecs_agent/error_ia"
)

type TYPE_AGENT string

const (
	DEEKSEEK TYPE_AGENT = "deekseek"
	GEMINI   TYPE_AGENT = "gemini"
)

type LLM interface {
	Reset()
	GetWelcomeMessage() string
	SetWelcomeMessage(msg string)
	SetTpl(prompt_path string, is_path ...bool) *base
	Format(var_tpl map[string]any) (string, error)
	Ask(question string, isTool ...bool) string
}

// New returns a new instance of the LLM given by typeAgent and c.
// typeAgent must be one of config.DEEKSEEK or config.GEMINI.
// c must be a config.ConfigModel.
// If typeAgent is not recognized, New panics with "Agent not found".
func New(typeAgent TYPE_AGENT, c config.ConfigModel) LLM {
	if c.APIKey == "" {
		panic(error_ia.ErrorApiKeyNotSet)
	}
	switch typeAgent {
	case DEEKSEEK:
		return newDeekSeek(c)
	case GEMINI:
		return newGemini(c)
	default:
		panic(error_ia.ErrorAgentNotFound)
	}
}
