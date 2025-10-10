package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/ecsavigne/ecs_agent/agent"
	ia "github.com/ecsavigne/ecs_agent/config"

	"github.com/tmc/langchaingo/llms"
)

func getMailNotRead(sender string, maxResults int) string {
	// return "Show me the emails not read from tools getMailNotRead\n"
	return `Remitente: faturadigital@lightvirtual.com.br
	Body: Paga la factura de electricidad, atentamente empresa light 
	`
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
	// iaChat := agent.New(agent.GEMINI, ia.ConfigModel{
	// 	// iaChat := agent.New(agent.DEEKSEEK, ia.ConfigModel{
	// 	// RootPrompt: "Eres un especialisata en **{{.Especiality}}** y tu nombre es **{{.Nombre}}**. da un mensaje de bienvenida de una oración simple.",
	// 	APIKey: "ApiKey",
	// 	// TemplateVar: map[string]any{"Nombre": "MailBot", "Especiality": "Analisis de emails"},
	// 	// Tool:        tools,
	// })
	// iaChat := agent.New(config.DEEKSEEK,
	// iaChat := agent.New(config.GEMINI,
	iaChat := agent.New(ia.NEW_GAI,
		ia.WithAPIKey("Your_ApiKey"),
		ia.WithRootPrompt("Eres un especialisata en **{{.Especiality}}** y tu nombre es **{{.Nombre}}**. da un mensaje de bienvenida de una oración simple."),
		ia.WithTemplateVar(map[string]any{"Nombre": "MailBot GAI", "Especiality": "Analisis de emails"}),
		ia.WithTool(tools),
	)

	fmt.Printf("\033[92mIAChat:\033[0m\n%s\n", iaChat.GetWelcomeMessage())
	// fmt.Printf("\033[92mYou:\033[0m\n%s\n", "Neceto informacion sobre el real madrid en esta temporada.")
	/*Ex. use of chains */
	//  prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
	//     prompts.NewHumanMessagePromptTemplate(
	//         "Neceto informacion sobre el .{team} en esta temporada.",
	//         []string{}, // Sin variables
	//     ),
	// })

	// chain := iaChat.NewLLMChain(prompts.NewPromptTemplate("Neceto informacion sobre el Real Madrid de futbol en esta temporada 2025-2026.", []string{""}))
	// out, e := iaChat.Run(context.Background(), chain, "")
	// if e != nil {
	// 	fmt.Println("Error executing chain: ", e.Error())
	// }
	// fmt.Printf("\033[92mIAChat:\033[0m\n%+v\n", out)

	for {
		var input string
		reader := bufio.NewScanner(os.Stdin)
		fmt.Print("\033[94mYou:\033[0m\n")
		if reader.Scan() {
			input = reader.Text()
			// input := "Neceto informacion sobre el Real Madrid de futbol en esta temporada 2025-2026."
			if input == "exit" {
				os.Exit(0)
			}

			buff := bytes.Buffer{}
			str := iaChat.Ask(input, func(ctx context.Context, chunk []byte) error {
				buff.Write(chunk)
				return nil
			})

			if buff.Len() == 0 {
				buff.WriteString(str)
			}
			// iaChat.Ask(input, nil)

			fmt.Printf("\033[92mIAChat:\033[0m\n%s\n", buff.String())
			// input = "exit"
			// fmt.Printf("\033[94mYou:\033[0m\n")
		}
	}
}
