package game

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/SunshineZzzz/purego-sdl3/img"
	"github.com/SunshineZzzz/purego-sdl3/sdl"
	"github.com/SunshineZzzz/purego-sdl3/ttf"
)

var (
	instance *Game
	once     sync.Once
)

func GetInstance() *Game {
	once.Do(func() {
		instance = &Game{
			fps:          60,
			frameTime:    1000000000 / 60,
			isRunning:    false,
			windowWidth:  600,
			windowHeight: 800,
			sdlWindow:    nil,
			sdlRenderer:  nil,
			titleFont:    nil,
			textFont:     nil,
			deltaTime:    float32(0.0),
			isFullscreen: false,
			currentScene: nil,
			finalScore:   0,
			leaderBoard:  make(map[uint32][]string),
		}
	})
	return instance
}

type Game struct {
	// FPS
	fps uint32
	// 每帧时间间隔，单位纳秒
	frameTime uint64
	// 是否运行中
	isRunning bool
	// 窗口宽高
	windowWidth  int32
	windowHeight int32
	// SDL窗口
	sdlWindow *sdl.Window
	// SDL渲染器
	sdlRenderer *sdl.Renderer
	// 背景
	nearStars background
	farStars  background
	// 标题字体
	titleFont *ttf.Font
	// 文字字体
	textFont *ttf.Font
	// 两帧时间差，秒
	deltaTime float32
	// 是否全屏
	isFullscreen bool
	// 当前场景
	currentScene iscene
	// 最终得分
	finalScore uint32
	// 排行榜
	leaderBoard map[uint32][]string
}

func (g *Game) Init() error {
	// 初始化 SDL
	if !sdl.Init(sdl.InitVideo | sdl.InitAudio | sdl.InitEvents) {
		return fmt.Errorf("sdl init error,%s", sdl.GetError())
	}

	// 创建窗口
	g.sdlWindow = sdl.CreateWindow("SDL Tutorial", g.windowWidth, g.windowHeight, sdl.WindowResizable)
	if g.sdlWindow == nil {
		return fmt.Errorf("sdl create window error,%s", sdl.GetError())
	}

	// 创建渲染器
	g.sdlRenderer = sdl.CreateRenderer(g.sdlWindow, "")
	if g.sdlRenderer == nil {
		return fmt.Errorf("sdl create renderer error,%s", sdl.GetError())
	}

	// 设置渲染器的逻辑尺寸，实现自动的视口缩放和坐标映射
	if !sdl.SetRenderLogicalPresentation(g.sdlRenderer, g.windowWidth, g.windowHeight, sdl.LogicalPresentationLetterbox) {
		return fmt.Errorf("sdl set render logical presentation error,%s", sdl.GetError())
	}

	// 初始化 TTF
	if !ttf.Init() {
		return fmt.Errorf("ttf init error,%s", sdl.GetError())
	}

	// 初始化近背景卷轴
	g.nearStars.speed = 30.0
	g.nearStars.texture = img.LoadTexture(g.sdlRenderer, "assets/image/Stars-A.png")
	if g.nearStars.texture == nil {
		return fmt.Errorf("load near stars texture error,%s", sdl.GetError())
	}
	sdl.GetTextureSize(g.nearStars.texture, &g.nearStars.width, &g.nearStars.height)
	g.nearStars.width /= 2
	g.nearStars.height /= 2
	// 初始化远背景卷轴
	g.farStars.speed = 30.0
	g.farStars.texture = img.LoadTexture(g.sdlRenderer, "assets/image/Stars-B.png")
	if g.farStars.texture == nil {
		return fmt.Errorf("load far stars texture error,%s", sdl.GetError())
	}
	sdl.GetTextureSize(g.farStars.texture, &g.farStars.width, &g.farStars.height)
	g.farStars.width /= 2
	g.farStars.height /= 2

	// 载入字体
	g.titleFont = ttf.OpenFont("assets/font/VonwaonBitmap-16px.ttf", 64.0)
	if g.titleFont == nil {
		return fmt.Errorf("load title font error,%s", sdl.GetError())
	}
	g.textFont = ttf.OpenFont("assets/font/VonwaonBitmap-16px.ttf", 32.0)
	if g.textFont == nil {
		return fmt.Errorf("load text font error,%s", sdl.GetError())
	}

	// 载入排行榜
	g.loadData()

	// 创建标题场景
	g.currentScene = &sceneTitle{}
	g.currentScene.init()

	g.isRunning = true
	return nil
}

func (g *Game) Run() {
	for g.isRunning {
		frameStart := sdl.GetTicksNS()
		g.handleEvent()
		g.update()
		g.render()
		frameEnd := sdl.GetTicksNS()
		diff := frameEnd - frameStart
		if diff < g.frameTime {
			sdl.DelayNS(g.frameTime - diff)
			g.deltaTime = float32(g.frameTime) / 1e9
			continue
		}
		g.deltaTime = float32(diff) / 1e9
	}
}

func (g *Game) handleEvent() {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		if event.Type() == sdl.EventQuit {
			g.isRunning = false
			return
		}
		if event.Type() == sdl.EventKeyDown {
			if event.Key().Scancode == sdl.ScancodeF4 {
				g.isFullscreen = !g.isFullscreen
				sdl.SetWindowFullscreen(g.sdlWindow, g.isFullscreen)
			}
		}
		if event.Type() == sdl.EventWindowResized {
			g.windowWidth = event.Window().Data1
			g.windowHeight = event.Window().Data2
			sdl.SetRenderLogicalPresentation(g.sdlRenderer, g.windowWidth, g.windowHeight, sdl.LogicalPresentationLetterbox)
		}
		g.currentScene.handleEvent(event)
	}
}

func (g *Game) update() {
	g.backgroundUpdate(g.deltaTime)
	// 更新当前场景
	g.currentScene.update(g.deltaTime)
}

func (g *Game) render() {
	// 清空渲染器
	sdl.RenderClear(g.sdlRenderer)

	// 渲染星空背景
	g.renderBackground()
	// 渲染当前场景
	g.currentScene.render()

	// 显示更新
	sdl.RenderPresent(g.sdlRenderer)
}

func (g *Game) Clean() {
	if g.currentScene != nil {
		g.currentScene.clean()
		g.currentScene = nil
	}

	if g.nearStars.texture != nil {
		sdl.DestroyTexture(g.nearStars.texture)
		g.nearStars.texture = nil
	}

	if g.farStars.texture != nil {
		sdl.DestroyTexture(g.farStars.texture)
		g.farStars.texture = nil
	}

	if g.titleFont != nil {
		ttf.CloseFont(g.titleFont)
		g.titleFont = nil
	}

	if g.textFont != nil {
		ttf.CloseFont(g.textFont)
		g.textFont = nil
	}

	ttf.Quit()
	sdl.DestroyRenderer(g.sdlRenderer)
	sdl.DestroyWindow(g.sdlWindow)
	sdl.Quit()

	// 保存排行榜数据
	g.saveData()
}

func (g *Game) renderBackground() {
	// 渲染远处的星星
	for posY := g.farStars.offset; posY < float32(g.windowHeight); posY += g.farStars.height {
		for posX := float32(0.0); posX < float32(g.windowWidth); posX += g.farStars.width {
			ds := sdl.FRect{X: posX, Y: posY, W: g.farStars.width, H: g.farStars.height}
			sdl.RenderTexture(g.sdlRenderer, g.farStars.texture, nil, &ds)
		}
	}
	// 渲染近处的星星
	for posY := g.nearStars.offset; posY < float32(g.windowHeight); posY += g.nearStars.height {
		for posX := float32(0.0); posX < float32(g.windowWidth); posX += g.nearStars.width {
			ds := sdl.FRect{X: posX, Y: posY, W: g.nearStars.width, H: g.nearStars.height}
			sdl.RenderTexture(g.sdlRenderer, g.nearStars.texture, nil, &ds)
		}
	}
}

func (g *Game) backgroundUpdate(deltaTime float32) {
	g.nearStars.offset += g.nearStars.speed * deltaTime
	if g.nearStars.offset >= 0 {
		g.nearStars.offset -= g.nearStars.height
	}

	g.farStars.offset += g.farStars.speed * deltaTime
	if g.farStars.offset >= 0 {
		g.farStars.offset -= g.farStars.height
	}
}

func (g *Game) renderTextCentered(text string, posY float32, isTitle bool) sdl.FPoint {
	color := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	var surface *sdl.Surface = nil

	if isTitle {
		surface = ttf.RenderTextSolid(g.titleFont, text, 0, color)
	} else {
		surface = ttf.RenderTextSolid(g.textFont, text, 0, color)
	}

	texture := sdl.CreateTextureFromSurface(g.sdlRenderer, surface)
	y := (float32(g.windowHeight) - float32(surface.H)) * posY
	rect := sdl.FRect{
		X: float32(g.windowWidth/2 - surface.W/2),
		Y: y,
		W: float32(surface.W),
		H: float32(surface.H),
	}
	sdl.RenderTexture(g.sdlRenderer, texture, nil, &rect)

	sdl.DestroySurface(surface)
	sdl.DestroyTexture(texture)

	// 右上角坐标
	return sdl.FPoint{X: rect.X + rect.W, Y: rect.Y}
}

func (g *Game) renderTextPos(text string, posX float32, posY float32, isLeft bool) {
	color := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	surface := ttf.RenderTextSolid(g.textFont, text, 0, color)
	texture := sdl.CreateTextureFromSurface(g.sdlRenderer, surface)
	var rect sdl.FRect
	if isLeft {
		rect = sdl.FRect{
			X: posX,
			Y: posY,
			W: float32(surface.W),
			H: float32(surface.H),
		}
	} else {
		rect = sdl.FRect{
			X: float32(g.windowWidth - int32(posX) - surface.W),
			Y: posY,
			W: float32(surface.W),
			H: float32(surface.H),
		}
	}
	sdl.RenderTexture(g.sdlRenderer, texture, nil, &rect)
	sdl.DestroySurface(surface)
	sdl.DestroyTexture(texture)
}

func (g *Game) changeScene(scene iscene) {
	if g.currentScene != nil {
		g.currentScene.clean()
		g.currentScene = nil
	}
	g.currentScene = scene
	g.currentScene.init()
}

func (g *Game) insertLeaderBoard(score uint32, name string) {
	scoreNames, ok := g.leaderBoard[score]
	if !ok {
		scoreNames = make([]string, 0, 3)
	}
	scoreNames = append(scoreNames, name)
	g.leaderBoard[score] = scoreNames

	totalCount := 0
	for _, names := range g.leaderBoard {
		totalCount += len(names)
	}
	if totalCount > 8 {
		// 找到最小的分数
		var minScore uint32 = 0
		first := true
		for score := range g.leaderBoard {
			if first {
				minScore = score
				first = false
			} else if score < minScore {
				minScore = score
			}
		}
		// 删除最小分数的记录
		if names, exists := g.leaderBoard[minScore]; exists {
			if len(names) > 1 {
				// 如果有多条记录，只删除最后一条
				g.leaderBoard[minScore] = names[:len(names)-1]
			} else {
				// 如果只有一条记录，删除整个键
				delete(g.leaderBoard, minScore)
			}
		}
	}
}

func (g *Game) saveData() {
	// 保存得分榜的数据
	file, err := os.Create("assets/save.dat")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for score, names := range g.leaderBoard {
		for _, name := range names {
			fmt.Fprintf(file, "%v %v\n", score, name)
		}
	}
}

func (g *Game) loadData() {
	// 加载得分榜的数据
	file, err := os.OpenFile("assets/save.dat", os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(err)
	}
	defer file.Close()
	g.leaderBoard = make(map[uint32][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var score uint32
		var name string
		n, err := fmt.Sscanf(line, "%v %v", &score, &name)
		if err != nil || n != 2 {
			continue
		}
		g.insertLeaderBoard(score, name)
	}
}

func (g *Game) getSortedLeaderboard() []struct {
	score uint32
	names []string
} {
	// 提取所有分数到切片
	scores := make([]uint32, 0, len(g.leaderBoard))
	for score := range g.leaderBoard {
		scores = append(scores, score)
	}
	// 按分数降序排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i] > scores[j]
	})
	// 构建排序后的结果
	var result []struct {
		score uint32
		names []string
	}
	for _, score := range scores {
		result = append(result, struct {
			score uint32
			names []string
		}{
			score: score,
			names: g.leaderBoard[score],
		})
	}
	return result
}
