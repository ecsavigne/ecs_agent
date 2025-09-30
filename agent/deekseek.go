package agent

import (
	"context"
	"fmt"

	ia "github.com/ecsavigne/ecs_agent/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type deekSeek struct {
	llm            *openai.LLM
	tpl            *prompts.PromptTemplate
	varTpl         map[string]any
	welcomeMessage string
	history        []llms.MessageContent
	ConfigTool     *ia.ConfigTool
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

func (a *deekSeek) createMessage(role llms.ChatMessageType, msg string) llms.MessageContent {
	return llms.MessageContent{Role: role, Parts: []llms.ContentPart{llms.TextPart(msg)}}
}

func (a *deekSeek) createMessageToolCall(tool_call llms.ToolCall) llms.MessageContent {
	toolResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:           tool_call.ID,
				Type:         tool_call.Type,
				FunctionCall: tool_call.FunctionCall,
			},
		},
	}
	return toolResponse
}

func (a *deekSeek) createMessageTool(nameTool, tool_call_id, msg string) llms.MessageContent {
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

func (a *deekSeek) generateCompletion(isUsr ...bool) *llms.ContentResponse {
	opts := []llms.CallOption{
		// llms.WithMaxTokens(300),
		// llms.WithJSONMode(),
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

	// textStr := ""
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
				// Append tool_use to messageHistory
				msg := a.createMessageToolCall(tool_calls[0])
				a.history = append(a.history, msg)
				// create message role tool
				msg = a.createMessageTool(func_name, tool_call_id, strResp)
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

func (a *deekSeek) getContent(completion *llms.ContentResponse) string {
	if completion != nil && len(completion.Choices) > 0 {
		return completion.Choices[0].Content
	}

	return ""
}

func (a *deekSeek) setRootPrompt() {
	promptStr, _ := a.tpl.Format(a.varTpl)
	msg := a.createMessage(llms.ChatMessageTypeSystem, promptStr)
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

func (a *deekSeek) Reset() {
	a.history = []llms.MessageContent{}
	a.setRootPrompt()
}

func (a *deekSeek) GetWelcomeMessage() string {
	return a.welcomeMessage
}

func (a *deekSeek) SetWelcomeMessage(msg string) {
	a.welcomeMessage = msg
}

func (a *deekSeek) Ask(question string) string {
	msg := a.createMessage(llms.ChatMessageTypeHuman, question)
	a.history = append(a.history, msg)

	completion := a.generateCompletion(true)
	content := a.getContent(completion)
	msg = a.createMessage(llms.ChatMessageTypeAI, content)
	a.history = append(a.history, msg)

	return content
}
