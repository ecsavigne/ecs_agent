package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ecsavigne/ecs_agent/agent"
	ia "github.com/ecsavigne/ecs_agent/config"

	"github.com/tmc/langchaingo/llms"
)

func getMailNotRead(sender string, maxResults int) string {
	return "Show me the emails not read from tools getMailNotRead\n"
}

var tools = &ia.ConfigTool{
	Tool: []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getMailNotRead",
				Description: "Obtener emails no leídos dado un remitente ej: notifications@github.com, si no se informa el remitente será \"\".",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"sender": map[string]any{
							"type":        "string",
							"description": "remitente a buscar. Ej: notifications@github.com",
						},
						"maxResults": map[string]any{
							"type":        "integer",
							"description": "Cantidad de mensaje a buscar. Ej: 3",
						},
					},
					"required": []string{"sender", "maxResults"},
				},
				Strict: true,
			},
		},
	},
	Funcs: map[string]ia.FuncArgs{
		"getMailNotRead": {getMailNotRead, []string{"sender", "maxResults"}},
	},
}

func main() {
	// iaChat := agent.New(agent.DEEKSEEK, ia.ConfigModel{
	// 	RootPrompt:     "Eres un especialisata en **{{.Especiality}}** y tu nombre es **{{.Nombre}}**. da un mensaje de bienvenida de una oración simple.",
	// 	RootPromptPath: "",
	// APIKey: "asdsadsk-c97ed79162a7416a8e18e00b114a4356_v1asdas",
	// 	TemplateVar: map[string]any{"Nombre": "MailBot", "Especiality": "Analisis de emails"},
	// 	Tool:        tools,
	// })
	iaChat := agent.New(agent.GEMINI, ia.ConfigModel{
		RootPrompt:     "Eres un especialisata en **{{.Especiality}}** y tu nombre es **{{.Nombre}}**. da un mensaje de bienvenida de una oración simple.",
		RootPromptPath: "",
		APIKey:         "sadasdasd_AIzaSyDT2m0bCEQoStkgqptr1ZxOPoC4ddc06CEPMsk-c97ed79162a7416a8e18e00b114a4356sk-c97ed79162a7416a8e18e00b114a4356",
		TemplateVar:    map[string]any{"Nombre": "MailBot", "Especiality": "Analisis de emails"},
		Tool:           tools,
	})

	fmt.Printf("\033[92mIAChat:\033[0m\n%s\n", iaChat.GetWelcomeMessage())

	for {
		var input string
		reader := bufio.NewScanner(os.Stdin)
		fmt.Print("\033[94mYou:\033[0m\n")
		if reader.Scan() {
			input = reader.Text()
			// input = "encuentra todos los correos de este remitente notifications@github.com 3"
			// input = "Dame los 3 primeros correos"
			// input = "Que puedes hacer"
			if input == "exit" {
				os.Exit(0)
			}

			fmt.Printf("\033[92mIAChat\033[0m: %s\n", iaChat.Ask(input))
		}

	}
}
