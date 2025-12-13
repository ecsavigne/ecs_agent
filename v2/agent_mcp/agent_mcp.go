package agent_mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ecsavigne/ecs_agent/v2/agent"
	"github.com/ecsavigne/ecs_agent/v2/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

var (
	ErrorGetFechingTools             = errors.New("Error fetching tools")
	ErrorEmptySession                = errors.New("Empty session")
	ErrorCreatingRootPrompt          = errors.New("Error creating root prompt")
	ErrorNotContentInPrompt          = errors.New("Not content in prompt")
	ErrorHavePathPromptAndRootPrompt = errors.New("pathPrompt and rootPrompt can't have at the same	time.")
)

type TYPE_AGENT string

const (
	TYPE_AGENT_ONE_SHOT_ZERO     TYPE_AGENT = "OneShotZeroAgent"
	TYPE_AGENT_CONVERSATIONAL    TYPE_AGENT = "ConversationalAgent"
	TYPE_AGENT_OPEN_AI_FUNCTIONS TYPE_AGENT = "OpenAIFunctionsAgent"
)

func (ta *TYPE_AGENT) Enum() TYPE_AGENT {
	if ta == nil {
		return ""
	}

	return *ta
}

func (ta TYPE_AGENT) GetTypeAgent() *TYPE_AGENT {
	if ta == "" {
		return nil
	}

	return &ta
}

type TYPE_STORAGE_MEMORY string

type TYPE_MEMORY string

func (tm *TYPE_MEMORY) Enum() TYPE_MEMORY {
	if tm == nil {
		return ""
	}

	return *tm
}

type MCPAdapterTool struct {
	Function    *llms.FunctionDefinition
	name        string
	description string
	client      *mcpClient
}

func NewMCPAdapterTool(ctx context.Context, cl *mcpClient, function *llms.FunctionDefinition) *MCPAdapterTool {
	return &MCPAdapterTool{
		client:      cl,
		Function:    function,
		name:        function.Name,
		description: function.Description,
	}
}

func (m *MCPAdapterTool) Name() string {
	return m.name
}
func (m *MCPAdapterTool) Description() string {
	// Construir descripción enriquecida con el schema
	var sb strings.Builder
	sb.WriteString(m.description)
	sb.WriteString("\n\n")
	sb.WriteString("⚠️ IMPORTANT: Provide arguments as a valid JSON object.\n\n")

	// Agregar información de parámetros
	param := m.Function.Parameters.(map[string]any)
	if props, ok := param["properties"].(map[string]any); ok {
		sb.WriteString("Required JSON format:\n{\n")
		for paramName, paramInfo := range props {
			if info, ok := paramInfo.(map[string]any); ok {
				paramType := "string"
				if t, ok := info["type"].(string); ok {
					paramType = t
				}
				paramDesc := ""
				if d, ok := info["description"].(string); ok {
					paramDesc = d
				}
				fmt.Fprintf(&sb, "  \"%s\": <%s> // %s\n", paramName, paramType, paramDesc)
			}
		}
		sb.WriteString("}\n")
	}

	return sb.String()
}

func (m *MCPAdapterTool) Call(ctx context.Context, jsonInput string) (string, error) {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		return "", fmt.Errorf("error al deserializar entrada LLM: %w", err)
	}

	result, err := m.client.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      m.Name(),
		Arguments: input,
	})

	if err != nil {
		// if strings.Contains(err.Error(), "session not found") {
		// 	retry := 10
		// 	ticker := time.NewTicker(2 * time.Second)
		// 	for retry > 0 {
		// 		select {
		// 		case <-ticker.C:
		// 			m.cs = m.cs.Reconnect()
		// 			retry--
		// 			// ticker.Stop()
		// 		}
		// 	}
		// 	// reconectar al mcp server
		// }

		fmt.Println("Error calling tool: ", err)
		return "", fmt.Errorf("Error calling tool: %v", err)
	}

	output := strings.Builder{}
	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Fprintf(&output, "%s\n", textContent.Text)
		}
	}

	return output.String(), nil
}

type MCPClientConfig struct {
	// configuration for mcp
	Endpoint, Version, Name string
	// configuration for mcp
	// Print logs of ping
	ShowPing bool
	// configuration for mcp
	// interval between pings, default value 5s
	TimePing time.Duration
	// Configuration for Agent
	// path of file with content root prompt priority 1
	PathRootPrompt string
	// Configuration for Agent
	// str root prompt priority 2
	RootPrompt string
	// Configuration for Agent
	// value for vars for template style {{.var}}, ej: {"var": "value"}
	VarTemplate map[string]any
	// Configuration for Agent
	// name of vars for template style {{.var}}, ej: ["var"]
	InputVars []string
	// Configuration for Agent
	// Say true if you want to use memory. the momery is aplicable if TYPE_AGENT is TYPE_AGENT_CONVERSATIONAL or TYPE_AGENT_OPEN_AI_FUNCTIONS
	IsMemory bool
	// Configuration for Agent
	// Type Memory for Agent is aplicable only if IsMemory is true
	// Types:
	// - Buffer Memory: tores all conversation messages in a simple buffer. Best for short conversations where you want complete history.
	// - Window Buffer Memory: Maintains a sliding window of recent messages. Useful when you want to limit context length while preserving recent history.
	// - Token Buffer Memory: Manages memory based on token count rather than message count. Provides precise control over context size for LLM token limits.
	// - Summary Memory: Automatically summarizes older conversation history while keeping recent messages intact. Balances context preservation with memory efficiency.
	// - Chat Message History: Provides a lower-level interface for managing individual chat messages. Useful for custom memory implementations.
	TypeMemory TYPE_MEMORY
	// Configuration for Agent
	// Say true if you want to use storage memory
	IsStorageMemory bool
	// Configuration for Agent
	// Type Memory for Agent is aplicable only if IsStorageMemory is true
	// Types:
	// - In-Memory: Fast, temporary storage (default).
	// - File-based: Simple persistence to local files.
	// - Database: SQL or NoSQL database integration
	// - Redis: High-performance, distributed memory storage
	// - Custom: Implement your own storage backend
	TypeStorageMemory string
	// Configuration for Agent
	// typeAgent:
	// - OpenAIFunctions: for Function Calling(generation json)
	// - Conversational:  for ReAct (Reasoning & Action)
	// - OneShotZero: for General/Adaptable (Can to be ReAct o Function Calling)
	TypeAgent TYPE_AGENT
}

type mcpClient struct {
	*mcp.Client
	retry int
	// isConnect      bool
	session *mcp.ClientSession
	*MCPClientConfig
}

func connectClient(ctx context.Context, cl *mcp.Client, endPoint string) (cs *mcp.ClientSession, err error) {

	return cl.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endPoint}, nil)
}

func createClientAndSessionMcp(ctx context.Context, mcp_client *mcpClient, mcpCC MCPClientConfig) (err error) {
	if mcpCC.Endpoint == "" {
		panic("Opsss!!!!   Endpoint not defined")
	}

	name := mcpCC.Name
	if name == "" {
		name = "Client-Mcp"
	}

	version := mcpCC.Version
	if version == "" {
		version = "1.0.0"
	}

	if mcpCC.TimePing == 0 {
		mcpCC.TimePing = 5
	}

	mcpCC.Name = name
	mcpCC.Version = version

	mcpCT := mcpCC

	mcp_client.MCPClientConfig = &mcpCT

	mcp_client.Client = mcp.NewClient(&mcp.Implementation{Name: name, Version: version}, nil)

	mcp_client.session, err = connectClient(ctx, mcp_client.Client, mcpCC.Endpoint)

	return err
}

func NewMCPClient(ctx context.Context, mcpCC MCPClientConfig) *mcpClient {

	mcp_client := new(mcpClient)

	createClientAndSessionMcp(ctx, mcp_client, mcpCC)

	// go mcp_client.ping(ctx)

	return mcp_client
}

type agentMcp struct {
	mcpClient *mcpClient
	// agent     *agents.OneShotZeroAgent
	agent    agents.Agent
	executor *agents.Executor
	llm      llms.Model
	tools    []tools.Tool
	opts     []agents.Option
	memory   schema.Memory
}

func (agt *agentMcp) ping(ctx context.Context) {
	cl := agt.mcpClient
	ticker := time.NewTicker(cl.MCPClientConfig.TimePing * time.Second)
	defer ticker.Stop()

	var err error
	for range ticker.C {
		t := time.Now().Format(time.StampMilli)
		if cl.MCPClientConfig.ShowPing {
			fmt.Printf("[%s] - Ping to MCP server\n", t)
		}

		if cl.session == nil {
			if cl.MCPClientConfig.ShowPing {
				fmt.Printf("[%s] - Empty client session\n", t)
			}

			err = ErrorEmptySession
		} else {
			err = cl.session.Ping(ctx, &mcp.PingParams{})
		}

		if err != nil {
			if cl.MCPClientConfig.ShowPing {
				fmt.Printf("[%s] - Erro in Ping is.... : %s\n", t, err.Error())
			}

			// reconectar al mcp server
			var sessionTemp *mcp.ClientSession
			sessionTemp, err = connectClient(ctx, cl.Client, cl.MCPClientConfig.Endpoint)
			if err == nil {
				if sessionTemp != nil {
					if cl.MCPClientConfig.ShowPing {
						fmt.Printf("[%s] - Reconected with MCP server, session ID: %s\n", t, sessionTemp.ID())
					}
					cl.session = sessionTemp
					if len(agt.tools) == 0 {
						err = agt.getTools(ctx)
						agt.agent = createAgent(ctx, agt.mcpClient.TypeAgent.GetTypeAgent(), agt.llm, agt.tools, agt.opts...)
						agt.setExecutor(ctx)
						// agt.executor = agents.NewExecutor(agt.agent)
					}
				}
			} else {
				if cl.MCPClientConfig.ShowPing {
					fmt.Printf("[%s] - Reconected with MCP server fail error is: %s\n", t, err.Error())
				}
			}
		} else {
			if cl.MCPClientConfig.ShowPing {
				fmt.Printf("[%s] - Pong from MCP server\n", t)
			}
		}
	}

}

func (agt *agentMcp) getTools(ctx context.Context) error {
	var tlc []tools.Tool

	if agt.mcpClient.session == nil {
		return ErrorEmptySession
	}

	toolsResult, err := agt.mcpClient.session.ListTools(context.Background(), nil)
	if err != nil {
		fmt.Println("Err: ", err.Error())
		return ErrorGetFechingTools
	}

	for _, tool := range toolsResult.Tools {
		param := tool.InputSchema.(map[string]any)

		Function := &llms.FunctionDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  param,
			Strict:      true,
		}

		MCPAdapterTool := NewMCPAdapterTool(ctx, agt.mcpClient, Function)
		tlc = append(tlc, MCPAdapterTool)
	}

	agt.tools = tlc

	return nil
}

func createAgent(cxt context.Context, typeA *TYPE_AGENT, llm llms.Model, tools []tools.Tool, opts ...agents.Option) agents.Agent {
	var agt agents.Agent
	_ = cxt

	switch typeA.Enum() {
	case TYPE_AGENT_CONVERSATIONAL:
		agt = agents.NewConversationalAgent(llm, tools, opts...)
	case TYPE_AGENT_OPEN_AI_FUNCTIONS:
		agt = agents.NewOpenAIFunctionsAgent(llm, tools, opts...)
	default:
		agt = agents.NewOneShotAgent(llm, tools, opts...)
		// case TYPE_AGENT_ONE_SHOT_ZERO:
	}

	return agt
}

func createMemory(typeMemory TYPE_MEMORY) schema.Memory {
	switch typeMemory.Enum() {
	default:
		return memory.NewConversationWindowBuffer(1000)
	}

}

func (agt *agentMcp) setRootPrompt(ctx context.Context) error {
	if agt.mcpClient.RootPrompt == "" && agt.mcpClient.PathRootPrompt == "" {
		return ErrorNotContentInPrompt
	}

	if agt.mcpClient.RootPrompt != "" && agt.mcpClient.PathRootPrompt != "" {
		return ErrorHavePathPromptAndRootPrompt
	}

	content := ""
	inputVar := make([]string, 0)
	varTemp := make(map[string]any)

	if agt.mcpClient.RootPrompt != "" {
		content = agt.mcpClient.RootPrompt
	} else if agt.mcpClient.PathRootPrompt != "" {
		content = config.LoadPromptFromFile(agt.mcpClient.PathRootPrompt)
	}

	if len(agt.mcpClient.InputVars) > 0 {
		inputVar = agt.mcpClient.InputVars
	}

	if len(agt.mcpClient.VarTemplate) > 0 {
		varTemp = agt.mcpClient.VarTemplate
	}

	template := prompts.NewPromptTemplate(
		content,
		inputVar,
	)

	var err error
	agt.mcpClient.RootPrompt, err = template.Format(varTemp)
	if err != nil {
		return err
	}

	_, err = chains.Run(ctx, agt.executor, agt.mcpClient.RootPrompt)
	if err != nil {
		return ErrorCreatingRootPrompt
	}

	return nil
}

func (agt *agentMcp) setExecutor(ctx context.Context) {
	agt.executor = agents.NewExecutor(agt.agent, agt.opts...)

	err := agt.setRootPrompt(ctx)
	if errors.Is(err, ErrorHavePathPromptAndRootPrompt) {
		panic(err)
	}
}

// NewAgentMcp returns a new instance of the MCP agent given by llm, mcpConfig, and opts.
// llm must be either a llms.Model or an agent.LLM.
// mcpConfig must be a MCPClientConfig.
// opts can be of type agents.Option.
// If llm is a llms.Model, a new OneShotAgent is created with that model.
// If llm is an agent.LLM, a new OneShotAgent is created with the llm's base model.
// The returned agent is a *agentMcp, which contains the OneShotAgent and the mcpConfig.
func NewAgentMcp(llm any, mcpConfig MCPClientConfig, opts ...agents.Option) *agentMcp {
	agtMCP := &agentMcp{}

	ctx := context.Background()

	agtMCP.mcpClient = NewMCPClient(ctx, mcpConfig)

	// discoveryTools
	err := agtMCP.getTools(ctx)

	if errors.Is(err, ErrorEmptySession) {
		agtMCP.tools = make([]tools.Tool, 0)
	}

	if llm, ok := llm.(llms.Model); ok {
		agtMCP.llm = llm
	}

	if llm, ok := llm.(agent.LLM); ok {
		agtMCP.llm = llm.GetBase().GetLLM()
	}

	if mcpConfig.IsMemory {
		agtMCP.memory = createMemory(mcpConfig.TypeMemory)
		opts = append(opts, agents.WithMemory(agtMCP.memory))
	}

	typeAgt := mcpConfig.TypeAgent.GetTypeAgent()
	agtMCP.agent = createAgent(ctx, typeAgt, agtMCP.llm, agtMCP.tools, opts...)

	agtMCP.opts = opts
	// agtMCP.executor = agents.NewExecutor(agtMCP.agent, opts...)
	agtMCP.setExecutor(ctx)

	go agtMCP.ping(ctx)

	return agtMCP
}

func (agt *agentMcp) Run(ctx context.Context, input string) (string, error) {
	result, err := chains.Run(ctx, agt.executor, input)
	if err != nil {
		return "", fmt.Errorf("Error executing chain: %w", err)
	}

	return result, nil
}
