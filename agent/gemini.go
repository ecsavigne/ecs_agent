package agent

import (
	"fmt"

	"github.com/ecsavigne/ecs_agent/config"
)

type gemini struct {
	base
}

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
