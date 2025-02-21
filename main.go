package main

import (
	"go-template/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		panic(err)
		return
	}
}
