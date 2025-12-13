package agent

import (
	"fmt"

	"github.com/ecsavigne/ecs_agent/config"
)

type gemini struct {
	base
}

// newGemini returns a new instance of the Gemini agent given by typeGemini and c.
// c must be a config.ConfigModel.
// If typeGemini is not recognized, newGemini panics with "Agent not found".
// TODO: Add any additional error handling or checks for the c parameter.
func newGemini(c ...config.ConfigMod) (agent *gemini) {
	var a *gemini

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in newGemini: ", r)
		}
	}()

	b := base{}
	b.setConfigToBase(c...)

	a = &gemini{base: b}
	a.setRootPrompt()

	return a
}

func (g *gemini) GetBase() *base {
	return &g.base
}
