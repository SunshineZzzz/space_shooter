package game

import (
	"github.com/SunshineZzzz/purego-sdl3/sdl"
)

// 标题场景
type sceneTitle struct {
	// 背景音乐
	bgm *oggPlayer
	// 定时器
	timer float32
}

var _ iscene = (*sceneTitle)(nil)

func (s *sceneTitle) init() {
	bgm, err := newOggPlayer("assets/music/06_Battle_in_Space_Intro.ogg")
	if err != nil {
		panic(err)
	}
	s.bgm = bgm
	s.bgm.SetLoop(true)
	s.bgm.Play()
}

func (s *sceneTitle) update(deltaTime float32) {
	s.timer += deltaTime
	if s.timer > 1.0 {
		s.timer -= 1.0
	}
}

func (s *sceneTitle) render() {
	// 渲染标题文字
	titleText := "SDL太空战机"
	GetInstance().renderTextCentered(titleText, 0.4, true)

	// 渲染普通文字
	if s.timer < 0.5 {
		instructions := "按 J 键开始游戏"
		GetInstance().renderTextCentered(instructions, 0.8, false)
	}
}

func (s *sceneTitle) clean() {
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
}

func (s *sceneTitle) handleEvent(event sdl.Event) {
	if event.Type() == sdl.EventKeyDown {
		if event.Key().Scancode == sdl.ScancodeJ {
			GetInstance().changeScene(&sceneMain{})
		}
	}
}
