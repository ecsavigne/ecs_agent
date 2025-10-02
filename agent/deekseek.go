package agent

import (
	ia "github.com/ecsavigne/ecs_agent/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type deekSeek struct {
	base
}

func newDeekSeek(c ...ia.ConfigModel) *deekSeek {
	var (
		model  = "deepseek-chat"
		apiKey = ""
		mv     = make(map[string]any)
		tool   = new(ia.ConfigTool)
		tpl    = ""
	)

	if len(c) > 0 {
		config := c[0]
		if config.APIKey != "" {
			apiKey = config.APIKey
		}
		if config.Model != "" {
			model = config.Model
		}
		if config.RootPrompt != "" {
			tpl = config.RootPrompt
		}
		if config.RootPromptPath != "" {
			tpl = ia.LoadPromptFromFile(config.RootPromptPath)
		}
		if config.TemplateVar != nil {
			mv = config.TemplateVar
		}
		if config.Tool != nil {
			tool = config.Tool
		}
	}

	llm, _ := openai.New(
		openai.WithModel(model),
		openai.WithToken(apiKey),
		openai.WithBaseURL("https://api.deepseek.com"),
		// openai.WithResponseFormat(openai.ResponseFormatJSON),
	)

	promptRoot := prompts.NewPromptTemplate(tpl, []string{""})

	a := &deekSeek{
		base: base{
			llm:            llm,
			tpl:            &promptRoot,
			varTpl:         mv,
			welcomeMessage: "",
			history:        []llms.MessageContent{},
			ConfigTool:     tool,
			Typ:            DEEKSEEK,
		},
	}

	a.setRootPrompt()

	return a
}
