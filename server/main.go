package main

import (
	"server/internal"
	"server/internal/services"
)

func main() {
	internal.Init("hypertube.yml")
	services.Logger.Info("Run all configurations!")
	internal.StartServer()
}
