package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/tmc/langchaingo/llms"
)

type FuncStream func(ctx context.Context, chunk []byte) error

// args[0] is the callback_func and args[1] is the args(array with name args in order))
type FuncArgs [2]any

func (f FuncArgs) Get(pos int) any {
	return f[pos]
}

/*
Ej: de configuracion de ConfigTool

	var tools = &ia.ConfigTool{
	        Tool: []llms.Tool{
	            {
	                Type: "function",
	                Function: &llms.FunctionDefinition{
	                    Name:        "getMailNotRead",
	                    Description: "Obtener emails no leídos dado un remitente ej: notifications@github.com",
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
	                                // "default":     3,
	                            },
	                        },
	                        "required": []string{"sender", "maxResults"},
	                    },
	                    Strict: true,
	                },
	            },
	        },
	        Funcs: map[string]ia.FuncArgs{
	            "getMailNotRead": {getMailNotRead, []string{"sender,maxResults"}},
	        },
	    }
*/
type ConfigTool struct {
	// Definition of tools
	Tool []llms.Tool
	// map of functions key: nameFunc, value: is FuncArgs
	Funcs map[string]FuncArgs
}

// config for configure llm
type ConfigModel struct {
	APIKey string
	Model  string
	// One prompt string for configure with the root system
	RootPrompt string
	// Info prompt for configure with the root system located in a file
	RootPromptPath string
	// Values for configure var in the root prompt if exists (key: nameVar, value: valueVar)
	TemplateVar map[string]any
	Tool        *ConfigTool
}

func LoadPromptFromFile(pathPrompt string) string {
	bytes, e := os.ReadFile(pathPrompt)
	if e != nil {
		msg := fmt.Sprintf("Occurred one error in [ LoadPromptFromFile ] When read Promp file, error is: %s\n", e)
		panic(msg)
	}

	return string(bytes)
}

func ExecuteFunction(fn any, args string, paramsStr []string) string {
	func_exec := reflect.ValueOf(fn)
	funcType := func_exec.Type()
	params := []reflect.Value{}

	// Control the number of arguments
	if funcType.NumIn() != len(paramsStr) {
		panic(fmt.Sprintf("The number of arguments expected is %d, but received to %d ", funcType.NumIn(), len(paramsStr)))
	}

	argsAny := map[string]any{}
	e := json.Unmarshal([]byte(args), &argsAny)
	if e != nil {
		panic(e)
	}

	for i, v := range paramsStr {
		// Get type i param of the function
		expectedType := funcType.In(i)
		params = append(params, reflect.ValueOf(argsAny[v]).Convert(expectedType))
	}

	res := func_exec.Call(params)
	if len(res) > 0 {
		return fmt.Sprintf("%v", res[0])
	}

	return ""
}
