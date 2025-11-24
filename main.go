package main

import (
	"fmt"
	"sdlshoot/game"
)

func main() {
	game := game.GetInstance()
	if err := game.Init(); err != nil {
		fmt.Println(err)
		return
	}
	game.Run()
	game.Clean()
}
