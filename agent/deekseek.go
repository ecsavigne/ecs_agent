package agent

import (
	"fmt"

	"github.com/ecsavigne/ecs_agent/config"
)

type deekSeek struct {
	base
}

func newDeekSeek(c ...config.ConfigMod) *deekSeek {
	var a *deekSeek

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in newDeekSeek: ", r)
		}
	}()

	b := base{}
	b.setConfigToBase(c...)

	a = &deekSeek{base: b}
	a.setRootPrompt()

	return a
}
