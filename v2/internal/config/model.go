package config

import "github.com/ecsavigne/ecs_agent/v2/config"

func WithTypeAgent(typeAgent config.TYPE_MODEL) config.ConfigMod {
	return func(c *config.ConfigModel) {
		c.SetTypeModel(typeAgent)
	}
}
