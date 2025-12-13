package llm

import (
	"net/http"

	"cloud.google.com/go/auth"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

type clientOptions struct {
	APIKey string

	// Optional. Backend for GenAI. See Backend constants. Defaults to BackendGeminiAPI unless explicitly set to BackendVertexAI,
	// or the environment variable GOOGLE_GENAI_USE_VERTEXAI is set to "1" or "true".
	Backend genai.Backend

	// Optional. GCP Project ID for Vertex AI. Required for BackendVertexAI.
	// Can also be set via the GOOGLE_CLOUD_PROJECT environment variable.
	// Find your Project ID: https://cloud.google.com/resource-manager/docs/creating-managing-projects#identifying_projects
	Project string

	// Optional. GCP Location/Region for Vertex AI. Required for BackendVertexAI.
	// Can also be set via the GOOGLE_CLOUD_LOCATION or GOOGLE_CLOUD_REGION environment variable.
	// Generative AI locations: https://cloud.google.com/vertex-ai/generative-ai/docs/learn/locations.
	Location string

	// Optional. Google credentials.  If not specified, [Application Default Credentials] will be used.
	//
	// [Application Default Credentials]: https://developers.google.com/accounts/docs/application-default-credentials
	Credentials *auth.Credentials

	// Optional HTTP client to use. If nil, a default client will be created.
	// For Vertex AI, this client must handle authentication appropriately.
	HTTPClient *http.Client

	// Optional HTTP options to override.
	HTTPOptions genai.HTTPOptions

	envVarProvider func() map[string]string
}

// Options is a set of options for GoogleAI and Vertex clients.
type Options struct {
	DefaultModel          string
	DefaultEmbeddingModel string
	DefaultCandidateCount int
	DefaultMaxTokens      int
	DefaultTemperature    float64
	DefaultTopK           int
	DefaultTopP           float64
	HarmThreshold         HarmBlockThreshold

	clientOpts clientOptions
}

func DefaultOptions() Options {
	return Options{
		DefaultModel:          "gemini-2.0-flash",
		DefaultEmbeddingModel: "embedding-001",
		DefaultCandidateCount: 1,
		DefaultMaxTokens:      2048,
		DefaultTemperature:    0.5,
		DefaultTopK:           3,
		DefaultTopP:           0.95,
		HarmThreshold:         HarmBlockOnlyHigh,

		clientOpts: clientOptions{
			Backend:  genai.BackendGeminiAPI,
			Location: "",
			Project:  "",
		},
	}
}

type Option func(*Options)

func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.clientOpts.APIKey = apiKey
	}
}

func WithBackend(apiKey string) Option {
	return func(o *Options) {
		o.clientOpts.APIKey = apiKey
	}
}

func WithProject(p string) Option {
	return func(o *Options) {
		o.clientOpts.Project = p
	}
}

func WithLocation(l string) Option {
	return func(o *Options) {
		o.clientOpts.Location = l
	}
}

func WithCredentials(credentialsOpts *auth.CredentialsOptions) Option {
	return func(o *Options) {
		o.clientOpts.Credentials = auth.NewCredentials(credentialsOpts)
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *Options) {
		o.clientOpts.HTTPClient = httpClient
	}
}

func WithHTTPOption(httpOptions genai.HTTPOptions) Option {
	return func(o *Options) {
		o.clientOpts.HTTPOptions = httpOptions
	}
}

func WithEnvVarProvider(f func() map[string]string) Option {
	return func(o *Options) {
		o.clientOpts.envVarProvider = f
	}
}

func WithGRPCConn(conn *grpc.ClientConn) Option {
	return func(o *Options) {
		o.clientOpts.HTTPClient = &http.Client{}
	}
}

func WithDefaultModel(defaultModel string) Option {
	return func(o *Options) {
		o.DefaultModel = defaultModel
	}
}

func WithDefaultEmbeddingModel(defaultEmbeddingModel string) Option {
	return func(o *Options) {
		o.DefaultEmbeddingModel = defaultEmbeddingModel
	}
}

func WithDefaultCandidateCount(defaultCandidateCount int) Option {
	return func(o *Options) {
		o.DefaultCandidateCount = defaultCandidateCount
	}
}

func WithDefaultMaxTokens(maxTokens int) Option {
	return func(o *Options) {
		o.DefaultMaxTokens = maxTokens
	}
}

func WithDefaultTemperature(defaultTemperature float64) Option {
	return func(o *Options) {
		o.DefaultTemperature = defaultTemperature
	}
}

func WithDefaultTopK(defaultTopK int) Option {
	return func(o *Options) {
		o.DefaultTopK = defaultTopK
	}
}

func WithDefaultTopP(defaultTopP float64) Option {
	return func(o *Options) {
		o.DefaultTopP = defaultTopP
	}
}

func WithHarmThreshold(ht HarmBlockThreshold) Option {
	return func(o *Options) {
		o.HarmThreshold = ht
	}
}

func WithCachedContent(name string) llms.CallOption {
	return func(o *llms.CallOptions) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]interface{})
		}
		o.Metadata["CachedContentName"] = name
	}
}

type HarmBlockThreshold int32

const (
	// HarmBlockUnspecified means threshold is unspecified.
	HarmBlockUnspecified HarmBlockThreshold = 0
	// HarmBlockLowAndAbove means content with NEGLIGIBLE will be allowed.
	HarmBlockLowAndAbove HarmBlockThreshold = 1
	// HarmBlockMediumAndAbove means content with NEGLIGIBLE and LOW will be allowed.
	HarmBlockMediumAndAbove HarmBlockThreshold = 2
	// HarmBlockOnlyHigh means content with NEGLIGIBLE, LOW, and MEDIUM will be allowed.
	HarmBlockOnlyHigh HarmBlockThreshold = 3
	// HarmBlockNone means all content will be allowed.
	HarmBlockNone HarmBlockThreshold = 4
)
