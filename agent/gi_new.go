package agent

import (
	"fmt"

	ia "github.com/ecsavigne/ecs_agent/config"
)

type gaiNew struct {
	base
}

func newgaiNew(c ...ia.ConfigMod) (agent *gaiNew) {
	var a *gaiNew

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in newgaiNew: ", r)
		}
	}()

	b := base{}
	b.setConfigToBase(c...)

	a = &gaiNew{base: b}
	a.setRootPrompt()

	return a
}
