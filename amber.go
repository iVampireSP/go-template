package main

import (
	"go-template/cmd"
)

// @title           Leaflow Amber
// @version         1.0
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		panic(err)
		return
	}
}
