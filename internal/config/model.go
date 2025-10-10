package config

import "github.com/ecsavigne/ecs_agent/config"

func WithTypeAgent(typeAgent config.TYPE_AGENT) config.ConfigMod {
	return func(c *config.ConfigModel) {
		c.SetTypeAgent(typeAgent)
	}
}
