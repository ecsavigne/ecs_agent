//lint:file-ignore ST1005 This is a false positive
package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	ia "github.com/ecsavigne/ecs_agent/v2/config"
	"github.com/ecsavigne/ecs_agent/v2/error_ia"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

type base struct {
	llm            llms.Model
	tpl            *prompts.PromptTemplate
	varTpl         map[string]any
	welcomeMessage string
	history        []llms.MessageContent
	ConfigTool     *ia.ConfigTool
	Typ            ia.TYPE_MODEL
}

func (b *base) setConfigToBase(optsModel ...ia.ConfigMod) {
	ConfigModel := ia.DefaultConfigModel()

	for _, opt := range optsModel {
		opt(&ConfigModel)
	}

	tpl := ia.GetTplText(ConfigModel)
	promptRoot := prompts.NewPromptTemplate(tpl, []string{""})

	*b = base{
		tpl:            &promptRoot,
		varTpl:         ConfigModel.TemplateVar,
		welcomeMessage: "",
		history:        []llms.MessageContent{},
		ConfigTool:     ConfigModel.Tool,
	}

	switch ConfigModel.GetTypeAgent() {
	case ia.DEEKSEEK:
		b.llm = deekseekLLM(ConfigModel.APIKey, ConfigModel.Model)
		b.Typ = ia.DEEKSEEK
	case ia.GEMINI:
		b.llm = geminiLLM(ConfigModel.APIKey, ConfigModel.Model)
		b.Typ = ia.GEMINI
	case ia.NEW_GAI:
		b.llm = newGaiLLM(ConfigModel.APIKey, ConfigModel.Model)
		b.Typ = ia.NEW_GAI
	}

}

func (b *base) getHistory() []llms.MessageContent {
	return b.history
}

func (b *base) setHistory(history []llms.MessageContent) {
	b.history = history
}

func (b base) GetLLM() llms.Model {
	return b.llm
}

func (b *base) setRootPrompt() {
	promptStr, e := b.tpl.Format(b.varTpl)
	if e != nil || promptStr == "" {
		promptStr = `Eres un especialista en cualquier area. Tu nombre es Bot. Tu mensaje de bienvenida sera, ej: """Hola soy bot tu especialista en cualquier area"""`
	}

	msg := b.createMessage(llms.ChatMessageTypeSystem, promptStr)
	b.setHistory(append(b.getHistory(), msg))

	switch b.Typ {
	case ia.GEMINI, ia.NEW_GAI:
		{
			msg = b.createMessage(llms.ChatMessageTypeHuman, "quien eres tu?")
			b.setHistory(append(b.getHistory(), msg))
		}
	}

	completion := b.generateCompletion(nil)
	if completion == nil {
		// fmt.Println("Error generating root prompt")
		return
	}

	respText := b.getContent(completion)
	if respText != "" {
		msg = b.createMessage(llms.ChatMessageTypeAI, respText)
		b.welcomeMessage = respText
		b.setHistory(append(b.getHistory(), msg))
	}

}

func (*base) createMessage(role llms.ChatMessageType, msg string) llms.MessageContent {
	return llms.MessageContent{Role: role, Parts: []llms.ContentPart{llms.TextPart(msg)}}
}

func (*base) createMessageTool(nameTool, tool_call_id, msg string) llms.MessageContent {
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

func (*base) createMessageToolCall(tool_call llms.ToolCall) llms.MessageContent {
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

func (b *base) generateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	return b.GetLLM().GenerateContent(ctx, messages, options...)
}

func (b *base) generateCompletion(fnStream ia.FuncStream, isTool ...bool) *llms.ContentResponse {
	opts := []llms.CallOption{
		// llms.WithMaxTokens(300),
		// llms.WithJSONMode(),
		llms.WithTemperature(0.2),
		llms.WithTopP(0.4),

		// llms.WithStreamingFunc(steamTest),
	}

	if b.Typ != ia.NEW_GAI {
		opts = append(opts, llms.WithFrequencyPenalty(1.5))
	}

	var (
		completion *llms.ContentResponse
		msg        llms.MessageContent
		err        error
	)

	if len(isTool) > 0 && isTool[0] {
		if b.ConfigTool != nil && b.ConfigTool.Funcs != nil {
			opts = append(opts, []llms.CallOption{
				llms.WithTools(b.ConfigTool.Tool),
				llms.WithFunctionCallBehavior(llms.FunctionCallBehaviorAuto),
			}...)
		}
	}

	// textStr := ""
	// completion, err = b.llm.GenerateContent(
	if !(len(isTool) > 0 && isTool[0]) && fnStream != nil {
		opts = append(opts, llms.WithStreamingFunc(fnStream))
	}

	completion, err = b.generateContent(
		context.Background(),
		b.history,
		opts...,
	)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, llms.ErrUnexpectedChatMessageType) { // ErrRateLimit
			// Handle rate limiting
			time.Sleep(time.Second * 60)
			return b.generateCompletion(nil)
		} else if errors.Is(err, llms.ErrQuotaExceeded) { // ErrQuotaExceeded
			// Handle quota exceeded
			fmt.Println(error_ia.ErrorQuotaExceeded)
			return nil
		} else {
			// Handle other errors
			log.Printf("%s. Error is: %v", error_ia.ErrorGettingCompletion, err)
		}
	}

	if completion == nil {
		return nil
	}

	tool_calls := completion.Choices[0].ToolCalls
	if tool_calls != nil {
		call_func := tool_calls[0].FunctionCall
		name := call_func.Name
		params := call_func.Arguments
		tool_call_id := tool_calls[0].ID

		for func_name, fnArg := range b.ConfigTool.Funcs {
			if func_name == name {
				// execute function and get response
				strResp := ia.ExecuteFunction(fnArg.Get(0), params, fnArg.Get(1).([]string))

				msg = b.createMessageToolCall(tool_calls[0])
				b.setHistory(append(b.getHistory(), msg))

				// create message role tool
				msg = b.createMessageTool(func_name, tool_call_id, strResp)
				// add message to history
				b.setHistory(append(b.getHistory(), msg))
				// call generate completion again
				return b.generateCompletion(fnStream)
			}
		}
	}

	return completion
}

func (*base) getContent(completion *llms.ContentResponse) string {
	if completion != nil && len(completion.Choices) > 0 {
		return completion.Choices[0].Content
	}

	return ""
}

func (b *base) Reset() {
	b.history = []llms.MessageContent{}
	// b.setRootPrompt()
}

func (b *base) GetWelcomeMessage() string {
	return b.welcomeMessage
}

func (b *base) SetWelcomeMessage(msg string) {
	b.welcomeMessage = msg
}

// SetTpl sets the prompt template for the agent.
// The prompt template is used to generate the welcome message and to
// provide context for the AI to generate responses.
// If is_path is true, the prompt_path is expected to be a file path
// containing the prompt template. Otherwise, prompt_path is expected to
// contain the prompt template directly.
func (b *base) SetTpl(prompt_path string, is_path ...bool) *base {
	if len(is_path) > 0 && is_path[0] {
		prompt_path = ia.LoadPromptFromFile(prompt_path)
	}
	tpl := prompts.NewPromptTemplate(prompt_path, []string{""})
	b.tpl = &tpl

	return b
}

// Format returns a formatted string based on the prompt template and the
// given map of variables. It returns an error if the formatting fails.
// The format string is a Go template string, where variables are denoted
// as {{.VarName}}. The map of variables should contain the values for these
// variables.
func (b *base) Format(var_tpl map[string]any) (string, error) {
	return b.tpl.Format(var_tpl)
}

func (b *base) Ask(question string, fnStream ia.FuncStream, isTool ...bool) string {
	_isTool := false
	if len(isTool) > 0 {
		_isTool = isTool[0]
	}

	msg := b.createMessage(llms.ChatMessageTypeHuman, question)
	b.history = append(b.history, msg)

	completion := b.generateCompletion(fnStream, _isTool)
	content := b.getContent(completion)
	msg = b.createMessage(llms.ChatMessageTypeAI, content)
	b.history = append(b.history, msg)

	return content
}

func (b *base) NewLLMChain(prompt prompts.FormatPrompter, opts ...chains.ChainCallOption) *chains.LLMChain {
	return chains.NewLLMChain(b.GetLLM(), prompt, opts...)
}

func (*base) Run(ctx context.Context, c chains.Chain, input any, options ...chains.ChainCallOption) (string, error) {
	return chains.Run(ctx, c, input, options...)
}

func (*base) Call(ctx context.Context, c chains.Chain, input map[string]any, options ...chains.ChainCallOption) (map[string]any, error) {
	return chains.Call(ctx, c, input, options...)
}
