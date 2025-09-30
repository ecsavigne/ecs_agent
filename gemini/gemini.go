package gemini

import (
	ia "agent"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/prompts"
)

type gemini struct {
	llm            *googleai.GoogleAI
	tpl            *prompts.PromptTemplate
	varTpl         map[string]any
	welcomeMessage string
	history        []llms.MessageContent
	ConfigTool     *ia.ConfigTool
}

func NewGemini(c ...ia.ConfigModel) *gemini {
	var (
		model  = "gemini-2.5-flash"
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

	llm, _ := googleai.New(
		context.Background(),
		googleai.WithDefaultModel(model),
		googleai.WithAPIKey(apiKey),
	)

	promptRoot := prompts.NewPromptTemplate(tpl, []string{""})

	a := &gemini{
		llm:            llm,
		tpl:            &promptRoot,
		varTpl:         mv,
		welcomeMessage: "",
		history:        []llms.MessageContent{},
		ConfigTool:     tool,
	}

	a.setRootPrompt()

	return a
}

func (a *gemini) createMessage(role llms.ChatMessageType, msg string) llms.MessageContent {
	return llms.TextParts(role, msg)
}

func (a *gemini) createMessageTool(nameTool, tool_call_id, msg string) llms.MessageContent {
	toolResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: tool_call_id,
				Name:       nameTool,
				Content:    msg,
			},
		},
	}
	return toolResponse
}

// func steamTest(ctx context.Context, chunk []byte) error {
// 	fmt.Println("111111111111111    ", string(chunk), "    2222222222222222222")
// 	return nil
// }

func (a *gemini) generateCompletion(isUsr ...bool) *llms.ContentResponse {
	opts := []llms.CallOption{
		llms.WithTemperature(0.2),
		llms.WithTopP(0.4),
		llms.WithFrequencyPenalty(1.5),
		// llms.WithStreamingFunc(steamTest),
	}

	var (
		completion *llms.ContentResponse
		err        error
	)

	if len(isUsr) > 0 && isUsr[0] {
		if a.ConfigTool != nil && a.ConfigTool.Funcs != nil {
			opts = append(opts, []llms.CallOption{
				llms.WithTools(a.ConfigTool.Tool),
				llms.WithFunctionCallBehavior(llms.FunctionCallBehaviorAuto),
			}...)
		}
	}

	completion, err = a.llm.GenerateContent(
		context.Background(),
		a.history,
		opts...,
	)
	if err != nil {
		fmt.Println("Error generating completion: ", err)
		return nil
	}

	tool_calls := completion.Choices[0].ToolCalls
	if tool_calls != nil {
		call_func := tool_calls[0].FunctionCall
		name := call_func.Name
		params := call_func.Arguments
		tool_call_id := tool_calls[0].ID

		for func_name, fnArg := range a.ConfigTool.Funcs {
			if func_name == name {
				// execute function and get response
				strResp := ia.ExecuteFunction(fnArg.Get(0), params, fnArg.Get(1).([]string))
				// create message role tool
				msg := a.createMessageTool(func_name, tool_call_id, strResp)
				// add message to history
				a.history = append(a.history, msg)
				// call generate completion again
				completion = a.generateCompletion()
				return completion
			}
		}
	}

	return completion
}

func (a *gemini) getContent(completion *llms.ContentResponse) string {
	if completion != nil && len(completion.Choices) > 0 {
		return completion.Choices[0].Content
	}

	return ""
}

func (a *gemini) setRootPrompt() {
	promptStr, _ := a.tpl.Format(a.varTpl)
	msg := a.createMessage(llms.ChatMessageTypeSystem, promptStr)
	a.history = append(a.history, msg)
	msg = a.createMessage(llms.ChatMessageTypeHuman, "")
	a.history = append(a.history, msg)

	completion := a.generateCompletion()
	if completion == nil {
		fmt.Println("Error generating root prompt")
	}

	respText := a.getContent(completion)
	if respText != "" {
		msg = a.createMessage(llms.ChatMessageTypeAI, respText)
		a.welcomeMessage = respText
		a.history = append(a.history, msg)
	}

}

func (a *gemini) Reset() {
	a.history = []llms.MessageContent{}
	a.setRootPrompt()
}

func (a *gemini) GetWelcomeMessage() string {
	return a.welcomeMessage
}

func (a *gemini) SetWelcomeMessage(msg string) {
	a.welcomeMessage = msg
}

func (a *gemini) Ask(question string) string {
	msg := a.createMessage(llms.ChatMessageTypeHuman, question)
	a.history = append(a.history, msg)

	completion := a.generateCompletion(true)
	content := a.getContent(completion)
	msg = a.createMessage(llms.ChatMessageTypeAI, content)
	a.history = append(a.history, msg)

	return content
}
