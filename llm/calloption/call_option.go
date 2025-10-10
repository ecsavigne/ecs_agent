package calloption

import (
	"google.golang.org/genai"
)

type CallOption func(*genai.GenerateContentConfig)

func WithHTTPOption(httpOptions genai.HTTPOptions) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.HTTPOptions = &httpOptions
	}
}

func WithSystemInstruction(instruction genai.Content) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.SystemInstruction = &instruction
	}
}

func WithTemperature(temperature float32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.Temperature = &temperature
	}
}

func WithTopP(topP float32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.TopP = &topP
	}
}

func WithTopK(topK float32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.TopK = &topK
	}
}

func WithCandidateCount(candidateCount int32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.CandidateCount = candidateCount
	}
}

func WithMaxOutputTokens(maxOutputTokens int32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.MaxOutputTokens = maxOutputTokens
	}
}

func WithStopSequences(stopSequences []string) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.StopSequences = stopSequences
	}
}

func WithResponseLogprobs(responseLogprobs bool) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ResponseLogprobs = responseLogprobs
	}
}

func WithLogprobs(logprobs int32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.Logprobs = &logprobs
	}
}

func WithPresencePenalty(presencePenalty float32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.PresencePenalty = &presencePenalty
	}
}

func WithFrequencyPenalty(frequencyPenalty float32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.FrequencyPenalty = &frequencyPenalty
	}
}

func WithSeed(seed int32) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.Seed = &seed
	}
}

func WithResponseMIMEType(responseMIMEType string) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ResponseMIMEType = responseMIMEType
	}
}

func WithResponseSchema(responseSchema *genai.Schema) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ResponseSchema = responseSchema
	}
}

func WithResponseJsonSchema(responseJsonSchema any) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ResponseJsonSchema = responseJsonSchema
	}
}

func WithRoutingConfig(routingConfig genai.GenerationConfigRoutingConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.RoutingConfig = &routingConfig
	}
}

func WithModelSelectionConfig(modelSelectionConfig genai.ModelSelectionConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ModelSelectionConfig = &modelSelectionConfig
	}
}

func WithSafetySettings(safetySettings []genai.SafetySetting) CallOption {
	return func(o *genai.GenerateContentConfig) {
		for _, v := range safetySettings {
			o.SafetySettings = append(o.SafetySettings, &v)
		}
	}
}

func WithTools(tools []genai.Tool) CallOption {
	return func(o *genai.GenerateContentConfig) {
		for _, v := range tools {
			o.Tools = append(o.Tools, &v)
		}
	}
}

func WithToolConfig(toolConfig genai.ToolConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ToolConfig = &toolConfig
	}
}

func WithLabels(labels map[string]string) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.Labels = labels
	}
}

func WithCachedContent(cachedContent string) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.CachedContent = cachedContent
	}
}

func WithResponseModalities(responseModalities []string) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ResponseModalities = responseModalities
	}
}

func WithMediaResolution(mediaResolution genai.MediaResolution) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.MediaResolution = mediaResolution
	}
}

func WithSpeechConfig(speechConfig genai.SpeechConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.SpeechConfig = &speechConfig
	}
}

func WithAudioTimestamp(audioTimestamp bool) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.AudioTimestamp = audioTimestamp
	}
}

func WithThinkingConfig(thinkingConfig genai.ThinkingConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ThinkingConfig = &thinkingConfig
	}
}

func WithImageConfig(imageConfig genai.ImageConfig) CallOption {
	return func(o *genai.GenerateContentConfig) {
		o.ImageConfig = &imageConfig
	}
}
