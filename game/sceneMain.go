package game

import (
	"container/list"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/SunshineZzzz/purego-sdl3/img"
	"github.com/SunshineZzzz/purego-sdl3/sdl"
	"github.com/SunshineZzzz/purego-sdl3/ttf"
)

// 标题场景
type sceneMain struct {
	// 随机数生成器
	rand *rand.Rand
	// 游戏结束定时器
	timerEnd float32
	// 分数
	score uint32
	// 背景音乐
	bgm *oggPlayer
	// uiHealth纹理
	uiHealth *sdl.Texture
	// 分数字体
	scoreFont *ttf.Font
	// 音效map
	sounds map[string]*wavPlayer
	// 玩家
	player player
	// 是否死亡
	isDead bool
	// 玩家子弹模板
	projectilePlayerTemplate projectilePlayer
	// 玩家子弹列表
	projectilesPlayer *list.List
	// 敌人模板
	enemyTemplate enemy
	// 敌人列表
	enemies *list.List
	// 敌人子弹模板
	projectileEnemyTemplate projectileEnemy
	// 敌人子弹列表
	projectilesEnemy *list.List
	// 爆炸模板
	explosionTemplate explosion
	// 爆炸列表
	explosions *list.List
	// 声明物品模板
	itemLifeTemplate item
	// 物品列表
	items *list.List
}

var _ iscene = (*sceneMain)(nil)

func (s *sceneMain) init() {
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.isDead = false
	s.timerEnd = 0.0
	s.score = 0
	s.projectilesPlayer = list.New()
	s.enemies = list.New()
	s.projectilesEnemy = list.New()
	s.explosions = list.New()
	s.items = list.New()

	// 读取并播放背景音乐
	bgm, err := newOggPlayer("assets/music/03_Racing_Through_Asteroids_Loop.ogg")
	if err != nil {
		panic(err)
	}
	s.bgm = bgm
	s.bgm.SetLoop(true)
	s.bgm.Play()

	// 读取uiHealth纹理
	s.uiHealth = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/Health UI Black.png")
	if s.uiHealth == nil {
		panic("load uiHealth texture error")
	}

	// 载入字体
	s.scoreFont = ttf.OpenFont("assets/font/VonwaonBitmap-12px.ttf", 24.0)
	if s.scoreFont == nil {
		panic("load score font error")
	}

	// 读取音效
	s.sounds = make(map[string]*wavPlayer)
	s.sounds["player_shoot"], err = newWavPlayer("assets/sound/laser_shoot4.wav")
	if err != nil {
		panic(err)
	}
	s.sounds["enemy_shoot"], err = newWavPlayer("assets/sound/xs_laser.wav")
	if err != nil {
		panic(err)
	}
	s.sounds["player_explode"], err = newWavPlayer("assets/sound/explosion1.wav")
	if err != nil {
		panic(err)
	}
	s.sounds["enemy_explode"], err = newWavPlayer("assets/sound/explosion3.wav")
	if err != nil {
		panic(err)
	}
	s.sounds["hit"], err = newWavPlayer("assets/sound/eff11.wav")
	if err != nil {
		panic(err)
	}
	s.sounds["get_item"], err = newWavPlayer("assets/sound/eff5.wav")
	if err != nil {
		panic(err)
	}

	// 初始化玩家
	s.player.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/SpaceShip.png")
	s.player.speed = 300.0
	s.player.currentHealth = 3
	s.player.maxHealth = 3
	s.player.coolDown = 300
	s.player.lastShootTime = 0
	if s.player.texture == nil {
		panic("load player texture error")
	}
	sdl.GetTextureSize(s.player.texture, &s.player.width, &s.player.height)
	s.player.width /= 5.0
	s.player.height /= 5.0
	s.player.position.X = float32(GetInstance().windowWidth)/2.0 - s.player.width/2.0
	s.player.position.Y = float32(GetInstance().windowHeight) - s.player.height

	// 初始化玩家子弹模板
	s.projectilePlayerTemplate.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/laser-1.png")
	if s.projectilePlayerTemplate.texture == nil {
		panic("load projectile player texture error")
	}
	sdl.GetTextureSize(s.projectilePlayerTemplate.texture, &s.projectilePlayerTemplate.width, &s.projectilePlayerTemplate.height)
	s.projectilePlayerTemplate.width /= 4.0
	s.projectilePlayerTemplate.height /= 4.0
	s.projectilePlayerTemplate.speed = 600.0
	s.projectilePlayerTemplate.damage = 1

	// 初始化敌人模板
	s.enemyTemplate.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/insect-2.png")
	if s.enemyTemplate.texture == nil {
		panic("load enemy texture error")
	}
	sdl.GetTextureSize(s.enemyTemplate.texture, &s.enemyTemplate.width, &s.enemyTemplate.height)
	s.enemyTemplate.width /= 4.0
	s.enemyTemplate.height /= 4.0
	s.enemyTemplate.speed = 150.0
	s.enemyTemplate.currentHealth = 2
	s.enemyTemplate.coolDown = 2000
	s.enemyTemplate.lastShootTime = 0

	// 初始化敌人子弹模板
	s.projectileEnemyTemplate.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/bullet-1.png")
	if s.projectileEnemyTemplate.texture == nil {
		panic("load projectile enemy texture error")
	}
	sdl.GetTextureSize(s.projectileEnemyTemplate.texture, &s.projectileEnemyTemplate.width, &s.projectileEnemyTemplate.height)
	s.projectileEnemyTemplate.width /= 2.0
	s.projectileEnemyTemplate.height /= 2.0
	s.projectileEnemyTemplate.speed = 400.0
	s.projectileEnemyTemplate.damage = 1

	// 初始化爆炸模板
	s.explosionTemplate.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/effect/explosion.png")
	if s.explosionTemplate.texture == nil {
		panic("load explosion texture error")
	}
	sdl.GetTextureSize(s.explosionTemplate.texture, &s.explosionTemplate.width, &s.explosionTemplate.height)
	s.explosionTemplate.totalFrame = s.explosionTemplate.width / s.explosionTemplate.height
	s.explosionTemplate.width = s.explosionTemplate.height
	s.explosionTemplate.fps = 10

	// 初始化物品模板
	s.itemLifeTemplate.texture = img.LoadTexture(GetInstance().sdlRenderer, "assets/image/bonus_life.png")
	if s.itemLifeTemplate.texture == nil {
		panic("load item life texture error")
	}
	sdl.GetTextureSize(s.itemLifeTemplate.texture, &s.itemLifeTemplate.width, &s.itemLifeTemplate.height)
	s.itemLifeTemplate.width /= 4.0
	s.itemLifeTemplate.height /= 4.0
	s.itemLifeTemplate.speed = 200.0
	s.itemLifeTemplate.bounceCount = 3
	s.itemLifeTemplate.itemType = itemTypeLife
}

func (s *sceneMain) update(deltaTime float32) {
	s.keyboardControl(deltaTime)
	s.updatePlayerProjectiles(deltaTime)
	s.updateEnemyProjectiles(deltaTime)
	s.spawEnemy()
	s.updateEnemies(deltaTime)
	s.updatePlayer(deltaTime)
	s.updateExplosions(deltaTime)
	s.updateItems(deltaTime)
	if s.isDead {
		// 3秒后切换到标题场景
		s.changeSceneDelayed(deltaTime, 3)
	}
}

func (s *sceneMain) render() {
	// 渲染玩家子弹
	s.renderPlayerProjectiles()
	// 渲染敌机子弹
	s.renderEnemyProjectiles()
	// 渲染玩家
	if !s.isDead {
		ds := sdl.FRect{X: s.player.position.X, Y: s.player.position.Y, W: s.player.width, H: s.player.height}
		sdl.RenderTexture(GetInstance().sdlRenderer, s.player.texture, nil, &ds)
	}
	// 渲染敌人
	s.renderEnemies()
	// 渲染物品
	s.renderItems()
	// 渲染爆炸效果
	s.renderExplosions()
	// 渲染UI
	s.renderUI()
}

func (s *sceneMain) clean() {
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
	if s.uiHealth != nil {
		sdl.DestroyTexture(s.uiHealth)
		s.uiHealth = nil
	}
	if s.scoreFont != nil {
		ttf.CloseFont(s.scoreFont)
		s.scoreFont = nil
	}
	for _, sound := range s.sounds {
		sound.Close()
	}
	s.sounds = nil
	if s.player.texture != nil {
		sdl.DestroyTexture(s.player.texture)
		s.player.texture = nil
	}
	if s.projectilePlayerTemplate.texture != nil {
		sdl.DestroyTexture(s.projectilePlayerTemplate.texture)
		s.projectilePlayerTemplate.texture = nil
	}
	if s.projectileEnemyTemplate.texture != nil {
		sdl.DestroyTexture(s.projectileEnemyTemplate.texture)
		s.projectileEnemyTemplate.texture = nil
	}
	if s.explosionTemplate.texture != nil {
		sdl.DestroyTexture(s.explosionTemplate.texture)
		s.explosionTemplate.texture = nil
	}
	if s.itemLifeTemplate.texture != nil {
		sdl.DestroyTexture(s.itemLifeTemplate.texture)
		s.itemLifeTemplate.texture = nil
	}
	s.projectilesPlayer = nil
	s.enemies = nil
	s.projectilesEnemy = nil
	s.explosions = nil
	s.items = nil
}

func (s *sceneMain) handleEvent(event sdl.Event) {
	if event.Type() == sdl.EventKeyDown {
		if event.Key().Scancode == sdl.ScancodeEscape {
			GetInstance().changeScene(&sceneTitle{})
		}
	}
}

func (s *sceneMain) keyboardControl(deltaTime float32) {
	if s.isDead {
		return
	}

	// 获取键盘状态
	keyboardState := sdl.GetKeyboardState()
	if keyboardState[sdl.ScancodeW] {
		s.player.position.Y -= deltaTime * s.player.speed
	}
	if keyboardState[sdl.ScancodeS] {
		s.player.position.Y += deltaTime * s.player.speed
	}
	if keyboardState[sdl.ScancodeA] {
		s.player.position.X -= deltaTime * s.player.speed
	}
	if keyboardState[sdl.ScancodeD] {
		s.player.position.X += deltaTime * s.player.speed
	}

	// 限制飞机的移动范围
	if s.player.position.X < 0.0 {
		s.player.position.X = 0.0
	}
	if s.player.position.X > float32(GetInstance().windowWidth)-s.player.width {
		s.player.position.X = float32(GetInstance().windowWidth) - s.player.width
	}
	if s.player.position.Y < 0.0 {
		s.player.position.Y = 0.0
	}
	if s.player.position.Y > float32(GetInstance().windowHeight)-s.player.height {
		s.player.position.Y = float32(GetInstance().windowHeight) - s.player.height
	}

	// 控制子弹发射
	if keyboardState[sdl.ScancodeSpace] {
		currentTime := sdl.GetTicks()
		if currentTime-s.player.lastShootTime > s.player.coolDown {
			s.shootPlayer()
			s.player.lastShootTime = currentTime
		}
	}
}

func (s *sceneMain) shootPlayer() {
	// 在这里实现发射子弹的逻辑
	projectile := s.projectilePlayerTemplate
	projectile.position.X = s.player.position.X + s.player.width/2 - projectile.width/2
	projectile.position.Y = s.player.position.Y
	s.projectilesPlayer.PushBack(&projectile)
	s.sounds["player_shoot"].Play()
}

func (s *sceneMain) renderPlayerProjectiles() {
	for e := s.projectilesPlayer.Front(); e != nil; e = e.Next() {
		projectile := e.Value.(*projectilePlayer)
		ds := sdl.FRect{X: projectile.position.X, Y: projectile.position.Y, W: projectile.width, H: projectile.height}
		sdl.RenderTexture(GetInstance().sdlRenderer, projectile.texture, nil, &ds)
	}
}

func (s *sceneMain) renderEnemyProjectiles() {
	for e := s.projectilesEnemy.Front(); e != nil; e = e.Next() {
		projectile := e.Value.(*projectileEnemy)
		ds := sdl.FRect{X: projectile.position.X, Y: projectile.position.Y, W: projectile.width, H: projectile.height}
		var angle float64 = math.Atan2(float64(projectile.direction.Y), float64(projectile.direction.X))*180/math.Pi - 90.0
		sdl.RenderTextureRotated(GetInstance().sdlRenderer, projectile.texture, nil, &ds, angle, nil, sdl.FlipNone)
	}
}

func (s *sceneMain) updatePlayerProjectiles(deltaTime float32) {
	// 子弹超出屏幕外边界的距离
	margin := float32(32.0)
	for e := s.projectilesPlayer.Front(); e != nil; {
		next := e.Next()

		projectile := e.Value.(*projectilePlayer)
		projectile.position.Y -= projectile.speed * deltaTime
		// 检查子弹是否超出屏幕
		if projectile.position.Y+margin < 0 {
			s.projectilesPlayer.Remove(e)
		} else {
			for e := s.enemies.Front(); e != nil; e = e.Next() {
				enemy := e.Value.(*enemy)
				enemyRect := sdl.FRect{
					X: enemy.position.X,
					Y: enemy.position.Y,
					W: enemy.width,
					H: enemy.height,
				}
				projectileRect := sdl.FRect{
					X: projectile.position.X,
					Y: projectile.position.Y,
					W: projectile.width,
					H: projectile.height,
				}
				if sdl.HasRectIntersectionFloat(enemyRect, projectileRect) {
					enemy.currentHealth -= projectile.damage
					s.projectilesPlayer.Remove(e)
					s.sounds["hit"].Play()
					break
				}
			}
		}

		e = next
	}
}

func (s *sceneMain) updateEnemyProjectiles(deltaTime float32) {
	// 子弹超出屏幕外边界的距离
	margin := float32(32.0)

	for e := s.projectilesEnemy.Front(); e != nil; {
		next := e.Next()

		projectile := e.Value.(*projectileEnemy)
		projectile.position.X += projectile.speed * projectile.direction.X * deltaTime
		projectile.position.Y += projectile.speed * projectile.direction.Y * deltaTime
		if projectile.position.Y > float32(GetInstance().windowHeight)+margin ||
			projectile.position.Y < -margin ||
			projectile.position.X < -margin ||
			projectile.position.X > float32(GetInstance().windowWidth)+margin {
			s.projectilesEnemy.Remove(e)
		} else {
			projectileRect := sdl.FRect{
				X: projectile.position.X,
				Y: projectile.position.Y,
				W: projectile.width,
				H: projectile.height,
			}
			playerRect := sdl.FRect{
				X: s.player.position.X,
				Y: s.player.position.Y,
				W: s.player.width,
				H: s.player.height,
			}
			if sdl.HasRectIntersectionFloat(playerRect, projectileRect) && !s.isDead {
				s.player.currentHealth -= projectile.damage
				s.projectilesEnemy.Remove(e)
				s.sounds["hit"].Play()
				break
			}
		}

		e = next
	}
}

func (s *sceneMain) spawEnemy() {
	dis := s.rand.Float32()
	if dis > 1.0/60.0 {
		return
	}
	enemy := s.enemyTemplate
	enemy.position.X = s.rand.Float32() * (float32(GetInstance().windowWidth) - enemy.width)
	enemy.position.Y = -enemy.height
	s.enemies.PushBack(&enemy)
}

func (s *sceneMain) updateEnemies(deltaTime float32) {
	for e := s.enemies.Front(); e != nil; {
		next := e.Next()

		enemy := e.Value.(*enemy)
		enemy.position.Y += enemy.speed * deltaTime

		if enemy.position.Y > float32(GetInstance().windowHeight) {
			s.enemies.Remove(e)
		} else {
			currentTime := sdl.GetTicks()
			if enemy.currentHealth <= 0 {
				s.enemyExplode(enemy)
				s.enemies.Remove(e)
				e = next
				continue
			}
			if currentTime-enemy.lastShootTime > enemy.coolDown && !s.isDead {
				s.shootEnemy(enemy)
				enemy.lastShootTime = currentTime
			}
		}

		e = next
	}
}

func (s *sceneMain) enemyExplode(enemy *enemy) {
	currentTime := sdl.GetTicks()
	explosion := s.explosionTemplate
	explosion.position.X = enemy.position.X + enemy.width/2 - explosion.width/2
	explosion.position.Y = enemy.position.Y + enemy.height/2 - explosion.height/2
	explosion.startTime = currentTime
	s.explosions.PushBack(&explosion)
	s.sounds["enemy_explode"].Play()
	if s.rand.Float32() < 0.5 {
		s.dropItem(enemy)
	}
	s.score += 10
}

func (s *sceneMain) shootEnemy(enemy *enemy) {
	projectile := s.projectileEnemyTemplate
	projectile.position.X = enemy.position.X + enemy.width/2 - projectile.width/2
	projectile.position.Y = enemy.position.Y
	projectile.direction = s.getDirection(enemy)
	s.projectilesEnemy.PushBack(&projectile)
	s.sounds["enemy_shoot"].Play()
}

func (s *sceneMain) getDirection(enemy *enemy) sdl.FPoint {
	x := (s.player.position.X + s.player.width/2) - (enemy.position.X + enemy.width/2)
	y := (s.player.position.Y + s.player.height/2) - (enemy.position.Y + enemy.height/2)
	length := math.Sqrt(float64(x*x + y*y))
	x /= float32(length)
	y /= float32(length)
	return sdl.FPoint{X: x, Y: y}
}

func (s *sceneMain) renderEnemies() {
	for e := s.enemies.Front(); e != nil; e = e.Next() {
		enemy := e.Value.(*enemy)
		ds := sdl.FRect{X: enemy.position.X, Y: enemy.position.Y, W: enemy.width, H: enemy.height}
		sdl.RenderTexture(GetInstance().sdlRenderer, enemy.texture, nil, &ds)
	}
}

func (s *sceneMain) updatePlayer(float32) {
	if s.isDead {
		return
	}

	if s.player.currentHealth <= 0 {
		s.isDead = true
		currentTime := sdl.GetTicks()
		explosion := s.explosionTemplate
		explosion.position.X = s.player.position.X + s.player.width/2 - explosion.width/2
		explosion.position.Y = s.player.position.Y + s.player.height/2 - explosion.height/2
		explosion.startTime = currentTime
		s.explosions.PushBack(&explosion)
		s.sounds["player_explode"].Play()
		GetInstance().finalScore = s.score
		return
	}
	for e := s.enemies.Front(); e != nil; e = e.Next() {
		enemy := e.Value.(*enemy)
		enemyRect := sdl.FRect{
			X: enemy.position.X,
			Y: enemy.position.Y,
			W: enemy.width,
			H: enemy.height,
		}
		playerRect := sdl.FRect{
			X: s.player.position.X,
			Y: s.player.position.Y,
			W: s.player.width,
			H: s.player.height,
		}
		if sdl.HasRectIntersectionFloat(playerRect, enemyRect) {
			s.player.currentHealth -= 1
			enemy.currentHealth = 0
		}
	}
}

func (s *sceneMain) updateExplosions(float32) {
	currentTime := sdl.GetTicks()
	for e := s.explosions.Front(); e != nil; {
		next := e.Next()

		explosion := e.Value.(*explosion)
		explosion.currentFrame = float32(currentTime-explosion.startTime) / 1000.0 * float32(explosion.fps)
		fmt.Printf("currentFrame: %f, totalFrame: %f\n", explosion.currentFrame, explosion.totalFrame)
		if explosion.currentFrame >= explosion.totalFrame {
			s.explosions.Remove(e)
		}

		e = next
	}
}

func (s *sceneMain) renderExplosions() {
	for e := s.explosions.Front(); e != nil; e = e.Next() {
		explosion := e.Value.(*explosion)

		frameIndex := int32(explosion.currentFrame)
		if frameIndex >= int32(explosion.totalFrame) {
			frameIndex = int32(explosion.totalFrame) - 1
		}
		if frameIndex < 0 {
			frameIndex = 0
		}

		sc := sdl.FRect{
			X: float32(frameIndex) * explosion.width,
			Y: 0.0,
			W: explosion.width,
			H: explosion.height,
		}
		ds := sdl.FRect{
			X: explosion.position.X,
			Y: explosion.position.Y,
			W: explosion.width,
			H: explosion.height,
		}
		// fmt.Printf("sc: %f, %f, %f, %f\n", sc.X, sc.Y, sc.W, sc.H)
		sdl.RenderTexture(GetInstance().sdlRenderer, explosion.texture, &sc, &ds)
	}
}

func (s *sceneMain) dropItem(Enemy *enemy) {
	item := s.itemLifeTemplate
	item.position.X = Enemy.position.X + Enemy.width/2 - item.width/2
	item.position.Y = Enemy.position.Y + Enemy.height/2 - item.height/2
	angle := s.rand.Float64() * 2 * math.Pi
	item.direction.X = float32(math.Cos(angle))
	item.direction.Y = float32(math.Sin(angle))
	s.items.PushBack(&item)
}

func (s *sceneMain) renderItems() {
	for e := s.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*item)
		itemRect := sdl.FRect{
			X: item.position.X,
			Y: item.position.Y,
			W: item.width,
			H: item.height,
		}
		sdl.RenderTexture(GetInstance().sdlRenderer, item.texture, nil, &itemRect)
	}
}

func (s *sceneMain) updateItems(deltaTime float32) {
	for e := s.items.Front(); e != nil; {
		next := e.Next()

		item := e.Value.(*item)
		item.position.X += item.direction.X * item.speed * deltaTime
		item.position.Y += item.direction.Y * item.speed * deltaTime

		// 处理屏幕边缘反弹
		if item.position.X < 0 && item.bounceCount > 0 {
			item.direction.X = -item.direction.X
			item.bounceCount--
		}
		if item.position.X+item.width > float32(GetInstance().windowWidth) && item.bounceCount > 0 {
			item.direction.X = -item.direction.X
			item.bounceCount--
		}
		if item.position.Y < 0 && item.bounceCount > 0 {
			item.direction.Y = -item.direction.Y
			item.bounceCount--
		}
		if item.position.Y+item.height > float32(GetInstance().windowHeight) && item.bounceCount > 0 {
			item.direction.Y = -item.direction.Y
			item.bounceCount--
		}
		// 如果超出屏幕范围则删除
		if item.position.X+item.width < 0 ||
			item.position.X > float32(GetInstance().windowWidth) ||
			item.position.Y+item.height < 0 ||
			item.position.Y > float32(GetInstance().windowHeight) {
			s.items.Remove(e)
		} else {
			itemRect := sdl.FRect{
				X: item.position.X,
				Y: item.position.Y,
				W: item.width,
				H: item.height,
			}
			playerRect := sdl.FRect{
				X: s.player.position.X,
				Y: s.player.position.Y,
				W: s.player.width,
				H: s.player.height,
			}
			if sdl.HasRectIntersectionFloat(itemRect, playerRect) && !s.isDead {
				s.playerGetItem(item)
				s.items.Remove(e)
			}
		}

		e = next
	}
}

func (s *sceneMain) playerGetItem(item *item) {
	s.score += 5
	if item.itemType == itemTypeLife {
		s.player.currentHealth += 1
		if s.player.currentHealth > s.player.maxHealth {
			s.player.currentHealth = s.player.maxHealth
		}
	}
	s.sounds["get_item"].Play()
}

func (s *sceneMain) changeSceneDelayed(deltaTime float32, delay float32) {
	s.timerEnd += deltaTime
	if s.timerEnd > delay {
		GetInstance().changeScene(&sceneEnd{})
	}
}

func (s *sceneMain) renderUI() {
	// 渲染血条
	sdl.SetTextureColorMod(s.uiHealth, 100, 100, 100)
	x := float32(10)
	y := float32(10)
	size := float32(32)
	offset := float32(40)
	for i := int32(0); i < s.player.maxHealth; i++ {
		rect := sdl.FRect{
			X: x + float32(i)*offset,
			Y: y,
			W: size,
			H: size,
		}
		sdl.RenderTexture(GetInstance().sdlRenderer, s.uiHealth, nil, &rect)
	}
	// reset color
	sdl.SetTextureColorMod(s.uiHealth, 255, 255, 255)
	for i := int32(0); i < s.player.currentHealth; i++ {
		rect := sdl.FRect{
			X: x + float32(i)*offset,
			Y: y,
			W: size,
			H: size,
		}
		sdl.RenderTexture(GetInstance().sdlRenderer, s.uiHealth, nil, &rect)
	}
	// 渲染得分
	text := "SCORE:" + strconv.Itoa(int(s.score))
	color := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	var surface *sdl.Surface = nil
	surface = ttf.RenderTextSolid(s.scoreFont, text, 0, color)
	texture := sdl.CreateTextureFromSurface(GetInstance().sdlRenderer, surface)
	rect := sdl.FRect{
		X: float32(GetInstance().windowWidth - 10 - surface.W),
		Y: float32(10),
		W: float32(surface.W),
		H: float32(surface.H),
	}
	sdl.RenderTexture(GetInstance().sdlRenderer, texture, nil, &rect)
	sdl.DestroySurface(surface)
	sdl.DestroyTexture(texture)
}
