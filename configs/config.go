package configs

import _ "embed"

//go:embed config.yaml
var Config []byte

//go:embed rbac_model.conf
var RBACModel []byte
