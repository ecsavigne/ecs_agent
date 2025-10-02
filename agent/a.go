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

// New returns a new instance of the LLM given by typeAgent and c.
// typeAgent must be one of config.DEEKSEEK or config.GEMINI.
// c must be a config.ConfigModel.
// If typeAgent is not recognized, New panics with "Agent not found".
func New(typeAgent TYPE_AGENT, c config.ConfigModel) config.LLM {
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
