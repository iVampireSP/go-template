package main

import "github.com/iVampireSP/go-template/bootstrap"

func main() {
	app := bootstrap.CreateApplication()
	app.Run("foundation", "Go Project Template")
}
