package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ecsavigne/ecs_agent/v2/llm/calloption"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNoContentInResponse   = errors.New("no content in generation response")
	ErrUnknownPartInResponse = errors.New("unknown part type in generation response")
	ErrInvalidMimeType       = errors.New("invalid mime type on content")
)

const (
	CITATIONS            = "citations"
	SAFETY               = "safety"
	RoleSystem           = "system"
	RoleModel            = "model"
	RoleUser             = "user"
	RoleTool             = "tool"
	ResponseMIMETypeJson = "application/json"
)

// Call implements the [llms.Model] interface.
func (g *GoogleAI) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, g, prompt, options...)
}

func (g *GoogleAI) SupportsReasoning() bool {
	switch {
	case strings.Contains(g.model, "gemini-2.5"), strings.Contains(g.model, "gemini-2.0"),
		strings.Contains(g.model, "gemini-3"), strings.Contains(g.model, "gemini-4"),
		(strings.Contains(g.model, "gemini-exp") && strings.Contains(g.model, "thinking")):
		return true
	default:
		return false
	}
}

// convertParts converts between a sequence of langchain parts and genai parts.
func convertParts(parts []llms.ContentPart) ([]genai.Part, error) {
	convertedParts := make([]genai.Part, 0, len(parts))
	// for _, part := range parts {
	// 	var out genai.Part

	// 	switch p := part.(type) {
	// 	case llms.TextContent:
	// 		out = genai.Text(p.Text)
	// 	case llms.BinaryContent:
	// 		out = genai.Blob{MIMEType: p.MIMEType, Data: p.Data}
	// 	case llms.ImageURLContent:
	// 		typ, data, err := imageutil.DownloadImageData(p.URL)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		out = genai.ImageData(typ, data)
	// 	case llms.ToolCall:
	// 		fc := p.FunctionCall
	// 		var argsMap map[string]any
	// 		if err := json.Unmarshal([]byte(fc.Arguments), &argsMap); err != nil {
	// 			return convertedParts, err
	// 		}
	// 		out = genai.FunctionCall{
	// 			Name: fc.Name,
	// 			Args: argsMap,
	// 		}
	// 	case llms.ToolCallResponse:
	// 		out = genai.FunctionResponse{
	// 			Name: p.Name,
	// 			Response: map[string]any{
	// 				"response": p.Content,
	// 			},
	// 		}
	// 	}

	// 	convertedParts = append(convertedParts, out)
	// }
	return convertedParts, nil
}

func convertTool(llm_tools []llms.Tool) (tool []*genai.Tool) {
	if len(llm_tools) == 0 {
		return
	}
	// FunctionDeclarations []*FunctionDeclaration

	switch llm_tools[0].Type {
	case "function":
		param := genai.Schema{}
		if v, ok := llm_tools[0].Function.Parameters.(map[string]any); ok {
			b, err := json.Marshal(v)
			if err == nil {
				_ = json.Unmarshal(b, &param)
			}
		}

		tool = append(tool, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        llm_tools[0].Function.Name,
					Description: llm_tools[0].Function.Description,
					Parameters:  &param,
					// Strict:      llm_tools[0].Function.Strict,
				},
			},
		})
		return tool
	default:
		return
	}
}

func convertCallOption(options ...llms.CallOption) genai.GenerateContentConfig {
	optsLegacy := llms.CallOptions{}

	for _, opt := range options {
		opt(&optsLegacy)
	}

	optsModern := genai.GenerateContentConfig{
		CandidateCount:   int32(optsLegacy.CandidateCount),
		MaxOutputTokens:  int32(optsLegacy.MaxTokens),
		Temperature:      proto.Float32(float32(optsLegacy.Temperature)),
		TopP:             proto.Float32(float32(optsLegacy.TopP)),
		TopK:             proto.Float32(float32(optsLegacy.TopK)),
		FrequencyPenalty: proto.Float32(float32(optsLegacy.FrequencyPenalty)),
		PresencePenalty:  proto.Float32(float32(optsLegacy.PresencePenalty)),
		Seed:             proto.Int32(int32(optsLegacy.Seed)),
		StopSequences:    optsLegacy.StopWords,
		ResponseMIMEType: optsLegacy.ResponseMIMEType,
		Tools:            convertTool(optsLegacy.Tools),
	}

	return optsModern
}

func convertMessage(message llms.MessageContent) *genai.Content {
	var (
		// messageGenai *genai.Content
		content *genai.Content
		Arg     = map[string]any{}
	)

	// for _, message := range messages {
	// var parts []*genai.Part

	switch v := message.Parts[0].(type) {
	case llms.TextContent:
		content = genai.NewContentFromText(
			string(v.Text),
			genai.Role(string(message.Role)),
		)
	// case llms.ImageURLContent:
	// 	parts = []*genai.Part{{ImageURL: v.URL}}
	case llms.BinaryContent:
		// parts = []*genai.Part{{InlineData: &genai.Blob{MIMEType: v.MIMEType, Data: v.Data}}}
	case llms.ToolCall:
		_ = json.Unmarshal([]byte(v.FunctionCall.Arguments), &Arg)
		content = genai.NewContentFromFunctionCall(
			v.FunctionCall.Name,
			Arg,
			genai.Role(message.Role),
		)
		// parts = []*genai.Part{{
		// 	FunctionCall: &genai.FunctionCall{
		// 		ID:   v.ID,
		// 		Name: v.FunctionCall.Name,
		// 		Args: Arg,
		// 	},
		// }}
	case llms.ToolCallResponse:
		// parts = []*genai.Part{{
		// 	FunctionResponse: &genai.FunctionResponse{
		// 		ID:   v.ToolCallID,
		// 		Name: v.Name,
		// 		Args: Arg,
		// 	},
		// }}
		_ = json.Unmarshal([]byte(v.Content), &Arg)
		content = genai.NewContentFromFunctionResponse(
			v.Name,
			Arg,
			genai.Role(message.Role),
		)
	default:
		content = genai.NewContentFromText(
			"Unknown message type",
			genai.Role(message.Role),
		)
	}

	// messagesGenai = append(messagesGenai, content)
	// }

	return content
}

func (g *GoogleAI) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (r *llms.ContentResponse, e error) {
	config := convertCallOption(options...)

	return g.GenerateContentNew(ctx, messages, &config)
}

// response = &llms.ContentResponse{
func generateResponse(generate_resp *genai.GenerateContentResponse) (resp *llms.ContentResponse) {
	var (
		content   = ""
		candidate = generate_resp.Candidates
	)

	switch {
	case candidate[0].Content.Parts[0].Text != "":
		content = candidate[0].Content.Parts[0].Text
	case candidate[0].Content.Parts[0].VideoMetadata != nil:
	case candidate[0].Content.Parts[0].InlineData != nil:
	case candidate[0].Content.Parts[0].FileData != nil:
	case len(candidate[0].Content.Parts[0].ThoughtSignature) > 0:
	case candidate[0].Content.Parts[0].FunctionCall != nil:
	case candidate[0].Content.Parts[0].CodeExecutionResult != nil:
	case candidate[0].Content.Parts[0].ExecutableCode != nil:
	case candidate[0].Content.Parts[0].FunctionResponse != nil:
	case candidate[0].Content.Parts[0].Thought:

	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:    content,
				StopReason: string(candidate[0].FinishReason),
				GenerationInfo: map[string]any{
					"role":            string(candidate[0].Content.Role),
					"create_time":     generate_resp.CreateTime,
					"model_version":   generate_resp.ModelVersion,
					"token_count":     candidate[0].TokenCount,
					"finish_reason":   candidate[0].FinishReason,
					"finish_message":  candidate[0].FinishMessage,
					"prompt_feedback": generate_resp.PromptFeedback,
					"response_id":     generate_resp.ResponseID,
					"usage_metadata":  generate_resp.UsageMetadata,
				},
			},
		},
	}
}

// options can be of type llms.CallOption or calloption.CallOption
// func (g *GoogleAI) GenerateContentNew(ctx context.Context, messages []llms.MessageContent, options ...calloption.CallOption) (*llms.ContentResponse, error) {
func (g *GoogleAI) GenerateContentNew(ctx context.Context, messages []llms.MessageContent,
	config *genai.GenerateContentConfig, options ...calloption.CallOption) (*llms.ContentResponse, error) {
	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	var (
		response *llms.ContentResponse
		contents = []*genai.Content{}
	)

	for _, opt := range options {
		opt(config)
	}

	if messages[0].Role == llms.ChatMessageTypeSystem {
		config.SystemInstruction = convertMessage(messages[0])
		messages = messages[1:]
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	for _, message := range messages {
		switch message.Role {
		case llms.ChatMessageTypeSystem:
			config.SystemInstruction = convertMessage(message)
		case llms.ChatMessageTypeAI:
			message.Role = llms.ChatMessageType("model")
		case llms.ChatMessageTypeHuman:
			message.Role = llms.ChatMessageType("user")
		}

		contents = append(contents, convertMessage(message))
	}

	// Update the tracked model if it was overridden
	// effectiveModel := opts.Model
	// if effectiveModel != "" && effectiveModel != g.model {
	// 	g.model = effectiveModel
	// }

	model := g.client.Models

	// set model.ResponseMIMEType from either opts.JSONMode or opts.ResponseMIMEType
	switch {
	case config.ResponseMIMEType != "" && config.ResponseJsonSchema != nil:
		return nil, fmt.Errorf("conflicting options, can't use JSONMode and ResponseMIMEType together")
		// case opts.ResponseMIMEType != "" && !opts.JSONMode:
		// 	model.ResponseMIMEType = opts.ResponseMIMEType
		// case opts.ResponseMIMEType == "" && opts.JSONMode:
		// 	model.ResponseMIMEType = ResponseMIMETypeJson
	}

	generateContentResponse, err := model.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return nil, err
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	gr := generateContentResponse
	cr := (*llms.ContentResponse)(nil)
	if gr != nil && gr.Candidates != nil {
		cr = generateResponse(gr)
	}
	// return generateContentResponse, nil
	return cr, nil
}
