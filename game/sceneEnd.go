package game

import (
	"fmt"
	"strconv"

	"github.com/SunshineZzzz/purego-sdl3/sdl"
)

type sceneEnd struct {
	// 背景音乐
	bgm *oggPlayer
	// 是否正在输入
	isTyping bool
	// 名字
	name string
	// 闪烁的光标计时器
	blinkTimer float32
}

var _ iscene = (*sceneEnd)(nil)

func (s *sceneEnd) init() {
	s.isTyping = true
	s.blinkTimer = 1.0

	// 载入背景音乐
	bgm, err := newOggPlayer("assets/music/06_Battle_in_Space_Intro.ogg")
	if err != nil {
		panic(err)
	}
	s.bgm = bgm
	s.bgm.SetLoop(true)
	s.bgm.Play()

	if !sdl.TextInputActive(GetInstance().sdlWindow) {
		sdl.StartTextInput(GetInstance().sdlWindow)
	}
	if !sdl.TextInputActive(GetInstance().sdlWindow) {
		fmt.Printf("failed to start text input: %s\n", sdl.GetError())
	}
}

func (s *sceneEnd) update(deltaTime float32) {
	s.blinkTimer -= deltaTime
	if s.blinkTimer < 0.0 {
		s.blinkTimer += 1.0
	}
}

func (s *sceneEnd) render() {
	if s.isTyping {
		s.renderPhase1()
	} else {
		s.renderPhase2()
	}
}

func (s *sceneEnd) clean() {
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
}

func (s *sceneEnd) handleEvent(event sdl.Event) {
	if s.isTyping {
		if event.Type() == sdl.EventTextInput {
			ti := event.Text()
			s.name += ti.Text()
		}
		if event.Type() == sdl.EventKeyDown {
			if event.Key().Scancode == sdl.ScancodeReturn {
				s.isTyping = false
				sdl.StopTextInput(GetInstance().sdlWindow)
				if s.name == "" {
					s.name = "无名氏"
				}
				GetInstance().insertLeaderBoard(GetInstance().finalScore, s.name)
			}
			if event.Key().Scancode == sdl.ScancodeBackspace {
				if len(s.name) == 0 {
					return
				}
				runes := []rune(s.name)
				if len(runes) > 0 {
					runes = runes[:len(runes)-1]
					s.name = string(runes)
				}
			}
		}
	} else {
		if event.Type() == sdl.EventKeyDown {
			if event.Key().Scancode == sdl.ScancodeJ {
				GetInstance().changeScene(&sceneMain{})
			}
		}
	}
}

func (s *sceneEnd) renderPhase1() {
	score := GetInstance().finalScore
	scoreText := "你的得分是：" + strconv.FormatUint(uint64(score), 10)
	gameOver := "Game Over"
	instrutionText := "请输入你的名字，按回车键确认："
	GetInstance().renderTextCentered(scoreText, 0.1, false)
	GetInstance().renderTextCentered(gameOver, 0.4, true)
	GetInstance().renderTextCentered(instrutionText, 0.6, false)

	if s.name != "" {
		p := GetInstance().renderTextCentered(s.name, 0.8, false)
		if s.blinkTimer < 0.5 {
			GetInstance().renderTextPos("_", p.X, p.Y, true)
		}
	} else {
		if s.blinkTimer < 0.5 {
			GetInstance().renderTextCentered("_", 0.8, false)
		}
	}
}

func (s *sceneEnd) renderPhase2() {
	GetInstance().renderTextCentered("得分榜", 0.05, true)
	posY := 0.2 * float32(GetInstance().windowHeight)
	i := 1
	for _, entry := range GetInstance().getSortedLeaderboard() {
		for _, name := range entry.names {
			GetInstance().renderTextPos(fmt.Sprintf("%d. %s %d", i, name, entry.score), 100.0, posY, true)
			posY += 45
			i++
		}
	}
	if s.blinkTimer < 0.5 {
		GetInstance().renderTextCentered("按 J 键重新开始游戏", 0.85, false)
	}
}
