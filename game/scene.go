package game

import (
	"github.com/SunshineZzzz/purego-sdl3/sdl"
)

// 场景接口
type iscene interface {
	init()
	update(deltaTime float32)
	render()
	clean()
	handleEvent(event sdl.Event)
}
