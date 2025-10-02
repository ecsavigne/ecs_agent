package agent

import (
	"context"

	ia "github.com/ecsavigne/ecs_agent/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/prompts"
)

type gemini struct {
	base
}

func newGemini(c ...ia.ConfigModel) (agent *gemini) {
	var (
		model  = "gemini-2.5-flash"
		apiKey = ""
		mv     = make(map[string]any)
		tool   = new(ia.ConfigTool)
		tpl    = ""
		a      *gemini
	)

	defer func() {
		if r := recover(); r != nil {
			agent = a
		}
	}()

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

	llm, _ := googleai.New(
		context.Background(),
		googleai.WithDefaultModel(model),
		googleai.WithAPIKey(apiKey),
	)

	promptRoot := prompts.NewPromptTemplate(tpl, []string{""})

	a = &gemini{
		base: base{
			llm:            llm,
			tpl:            &promptRoot,
			varTpl:         mv,
			welcomeMessage: "",
			history:        []llms.MessageContent{},
			ConfigTool:     tool,
			Typ:            GEMINI,
		},
	}

	a.setRootPrompt()

	return a
}
