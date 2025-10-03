//lint:file-ignore ST1005 This is a false positive
package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	ia "github.com/ecsavigne/ecs_agent/config"
	"github.com/ecsavigne/ecs_agent/error_ia"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type base struct {
	llm            any
	tpl            *prompts.PromptTemplate
	varTpl         map[string]any
	welcomeMessage string
	history        []llms.MessageContent
	ConfigTool     *ia.ConfigTool
	Typ            TYPE_AGENT
}

func (b *base) getHistory() []llms.MessageContent {
	return b.history
}

func (b *base) setHistory(history []llms.MessageContent) {
	b.history = history
}

// func (b *base) getConfigTool() *ia.ConfigTool {
// 	return b.ConfigTool
// }

// func (b *base) setConfigTool(tool *ia.ConfigTool) {
// 	b.ConfigTool = tool
// }

func (b base) getGemini() *googleai.GoogleAI {
	if v, ok := b.llm.(*googleai.GoogleAI); ok {
		return v
	}
	return nil
}

func (b *base) getDeekSeek() *openai.LLM {
	if v, ok := b.llm.(*openai.LLM); ok {
		return v
	}

	return nil
}

func (b *base) setRootPrompt() {
	// promptStr, _ := b.tpl.Format(b.varTpl)
	promptStr, _ := b.tpl.Format(b.varTpl)

	msg := b.createMessage(llms.ChatMessageTypeSystem, promptStr)
	// b.history = append(b.history, msg)
	b.setHistory(append(b.getHistory(), msg))

	// switch b.Typ {
	// case GEMINI:
	// 	{
	// 		msg = b.createMessage(llms.ChatMessageTypeHuman, "")
	// 		// b.history = append(b.history, msg)
	// 		b.setHistory(append(b.getHistory(), msg))
	// 	}
	// }

	// completion := b.generateCompletion()
	// if completion == nil {
	// 	fmt.Println("Error generating root prompt")
	// }

	// respText := b.getContent(completion)
	// if respText != "" {
	// 	msg = b.createMessage(llms.ChatMessageTypeAI, respText)
	// 	b.welcomeMessage = respText
	// 	// b.history = append(b.history, msg)
	// 	b.setHistory(append(b.getHistory(), msg))
	// }

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

func (b *base) generateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {

	switch b.Typ {
	case GEMINI:
		if b.getGemini() == nil {
			return nil, fmt.Errorf("Gemini LLM not set")
		}
		return b.getGemini().GenerateContent(ctx, messages, options...)
	case DEEKSEEK:
		if b.getDeekSeek() == nil {
			return nil, fmt.Errorf("DeekSeek LLM not set")
		}
		return b.getDeekSeek().GenerateContent(ctx, messages, options...)
	default:
		return nil, nil
	}
}

func (b *base) generateCompletion(isTool ...bool) *llms.ContentResponse {
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
			return b.generateCompletion()
		} else if errors.Is(err, llms.ErrQuotaExceeded) { // ErrQuotaExceeded
			// Handle quota exceeded
			fmt.Println(error_ia.ErrorQuotaExceeded)
			return nil
		} else {
			// Handle other errors
			log.Printf("%s. Error is: %v", error_ia.ErrorGettingCompletion, err)

		}
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

				// switch b.Typ {
				// case GEMINI:
				// 	{ // Append tool_use to messageHistory
				msg = b.createMessageToolCall(tool_calls[0])
				// b.history = append(b.history, msg)
				b.setHistory(append(b.getHistory(), msg))
				// 	}
				// }

				// create message role tool
				msg = b.createMessageTool(func_name, tool_call_id, strResp)
				// add message to history
				// b.history = append(b.history, msg)
				b.setHistory(append(b.getHistory(), msg))
				// call generate completion again
				completion = b.generateCompletion()
				return completion
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

func (b *base) SetTpl(prompt_path string, is_path ...bool) *base {
	if len(is_path) > 0 && is_path[0] {
		prompt_path = ia.LoadPromptFromFile(prompt_path)
	}
	tpl := prompts.NewPromptTemplate(prompt_path, []string{""})
	b.tpl = &tpl

	return b
}

func (b *base) Format(var_tpl map[string]any) (string, error) {
	return b.tpl.Format(var_tpl)
}

func (b *base) Ask(question string, isTool ...bool) string {
	msg := b.createMessage(llms.ChatMessageTypeHuman, question)
	b.history = append(b.history, msg)

	completion := b.generateCompletion(true)
	content := b.getContent(completion)
	msg = b.createMessage(llms.ChatMessageTypeAI, content)
	b.history = append(b.history, msg)

	return content
}
