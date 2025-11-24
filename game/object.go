package game

import (
	"github.com/SunshineZzzz/purego-sdl3/sdl"
)

// 背景
type background struct {
	// 纹理
	texture *sdl.Texture
	// 宽度
	width float32
	// 高度
	height float32
	// 偏移量
	offset float32
	// 速度
	speed float32
}

// 玩家
type player struct {
	// 纹理
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 速度
	speed float32
	// 当前生命值
	currentHealth int32
	// 最大生命值
	maxHealth int32
	// 冷却时间
	coolDown uint64
	// 上次射击时间
	lastShootTime uint64
}

// 玩家子弹
type projectilePlayer struct {
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 速度
	speed float32
	// 伤害
	damage int32
}

// 敌机
type enemy struct {
	// 纹理
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 速度
	speed float32
	// 当前生命值
	currentHealth int32
	// 冷却时间
	coolDown uint64
	// 上次射击时间
	lastShootTime uint64
}

// 敌人子弹
type projectileEnemy struct {
	// 纹理
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 方向
	direction sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 速度
	speed float32
	// 伤害
	damage int32
}

// 爆炸
type explosion struct {
	// 序列帧动画对应的纹理
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 当前帧
	currentFrame float32
	// 总帧数
	totalFrame float32
	// 开始时间
	startTime uint64
	// 动画帧率，1秒多少张图片，totalFrame/fps = 播放时间
	fps uint32
}

// 物品类型
type itemType int32

const (
	itemTypeLife itemType = iota
	itemTypeShield
	itemTypeTime
)

// 物品
type item struct {
	// 纹理
	texture *sdl.Texture
	// 位置
	position sdl.FPoint
	// 方向
	direction sdl.FPoint
	// 宽度
	width float32
	// 高度
	height float32
	// 速度
	speed float32
	// 弹跳次数
	bounceCount int32
	// 物品类型
	itemType itemType
}
