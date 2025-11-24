package game

import (
	"fmt"
	"unsafe"

	"github.com/SunshineZzzz/purego-sdl3/sdl"
)

// 全局WAV播放器句柄管理
var wavPlayerHandles = struct {
	handles map[uint32]*wavPlayer
	nextID  uint32
}{
	handles: make(map[uint32]*wavPlayer),
	nextID:  1,
}

func registerWavPlayer(p *wavPlayer) uint32 {
	id := wavPlayerHandles.nextID
	wavPlayerHandles.handles[id] = p
	wavPlayerHandles.nextID++
	return id
}
func getWavPlayer(id uint32) *wavPlayer {
	return wavPlayerHandles.handles[id]
}
func unregisterWavPlayer(id uint32) {
	delete(wavPlayerHandles.handles, id)
}

// WAV播放器
type wavPlayer struct {
	// SDL音频流
	stream *sdl.AudioStream
	// WAV音频数据
	audioBuf *uint8
	// WAV音频数据长度
	audioLen int
	// 当前播放位置
	dataPos int
	// 正在播放
	isPlaying bool
	// 是否循环播放
	loop bool
	// 音频规格
	spec *sdl.AudioSpec
	// id
	id uint32
}

func newWavPlayer(soundFilePath string) (*wavPlayer, error) {
	// 打开文件IO流
	ioStream := sdl.IOFromFile(soundFilePath, "rb")
	if ioStream == nil {
		return nil, fmt.Errorf("failed to open WAV file: %s", sdl.GetError())
	}
	// 自动释放了
	// defer sdl.CloseIO(ioStream)

	// 使用SDL直接加载WAV文件
	var audioBuf *uint8
	var audioLen uint32
	spec := &sdl.AudioSpec{}
	// 加载WAV数据
	success := sdl.LoadWAVIO(ioStream, true, spec, &audioBuf, &audioLen)
	if !success {
		return nil, fmt.Errorf("failed to load WAV data: %s", sdl.GetError())
	}

	player := &wavPlayer{
		audioBuf:  audioBuf,
		audioLen:  int(audioLen),
		spec:      spec,
		dataPos:   0,
		isPlaying: false,
		loop:      false,
	}

	// 注册WAV播放器
	player.id = registerWavPlayer(player)

	// 创建音频流
	callback := sdl.NewAudioStreamCallback(wavAudioCallback)
	player.stream = sdl.OpenAudioDeviceStream(
		sdl.AudioDeviceDefaultPlayback,
		spec,
		callback,
		unsafe.Pointer(uintptr(player.id)),
	)

	if player.stream == nil {
		sdl.Free(unsafe.Pointer(audioBuf))
		return nil, fmt.Errorf("failed to open audio stream: %s", sdl.GetError())
	}

	return player, nil
}

// WAV音频回调函数
func wavAudioCallback(userdata unsafe.Pointer, stream *sdl.AudioStream, additionalAmount, totalAmount int32) {
	id := uint32(uintptr(userdata))
	player := getWavPlayer(id)

	// 安全检查
	if player == nil || player.id != id || !player.isPlaying || player.audioLen == 0 {
		return
	}

	// 计算剩余数据量
	remaining := player.audioLen - player.dataPos
	if remaining <= 0 {
		if player.loop {
			player.dataPos = 0
			remaining = player.audioLen
		} else {
			player.isPlaying = false
			sdl.AudioStreamDevicePaused(stream)
			return
		}
	}

	// 推送数据到音频流
	neededBytes := int(additionalAmount)
	dataToSend := min(neededBytes, remaining)
	if dataToSend > 0 {
		data := (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(player.audioBuf)) + uintptr(player.dataPos)))
		// player.audioBuf+player.dataPos
		sdl.PutAudioStreamData(stream, data, int32(dataToSend))
		player.dataPos += dataToSend
	}

	// 再次检查循环
	if player.loop && player.dataPos >= player.audioLen {
		player.dataPos = 0
	}
}

// 播放控制方法（保持不变）
func (p *wavPlayer) Play() {
	if p.stream == nil {
		return
	}
	p.isPlaying = true
	p.dataPos = 0
	sdl.ResumeAudioStreamDevice(p.stream)
}

func (p *wavPlayer) Pause() {
	if p.stream == nil {
		return
	}
	p.isPlaying = false
	sdl.AudioStreamDevicePaused(p.stream)
}

func (p *wavPlayer) Stop() {
	if p.stream == nil {
		return
	}
	p.isPlaying = false
	p.dataPos = 0
	sdl.AudioStreamDevicePaused(p.stream)
	sdl.ClearAudioStream(p.stream)
}

func (p *wavPlayer) SetLoop(loop bool) {
	p.loop = loop
}

// 关闭播放器，释放资源
func (p *wavPlayer) Close() {
	if p.stream != nil {
		p.Stop()
		sdl.DestroyAudioStream(p.stream)
		p.stream = nil
	}
	if p.audioLen > 0 {
		sdl.Free(unsafe.Pointer(p.audioBuf))
		p.audioBuf = nil
		p.audioLen = 0
	}
	unregisterWavPlayer(p.id)
}
