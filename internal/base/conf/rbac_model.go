package conf

import (
	"go-template/configs"
)

func (c *Config) GetRBACModel() string {
	return string(configs.RBACModel)
}
