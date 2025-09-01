package main

import (
	"server/internal"
)

func main() {
	internal.Init("hypertube.yml")
	internal.StartServer()
}
