package game

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/SunshineZzzz/purego-sdl3/sdl"
	"github.com/jfreymuth/oggvorbis"
)

// 全局OGG播放器句柄管理
var oggPlayerHandles = struct {
	handles map[uint32]*oggPlayer
	nextID  uint32
}{
	handles: make(map[uint32]*oggPlayer),
	nextID:  1,
}

func registerOGGPlayer(p *oggPlayer) uint32 {
	id := oggPlayerHandles.nextID
	oggPlayerHandles.handles[id] = p
	oggPlayerHandles.nextID++
	return id
}
func getOGGPlayer(id uint32) *oggPlayer {
	return oggPlayerHandles.handles[id]
}
func unregisterOGGPlayer(id uint32) {
	delete(oggPlayerHandles.handles, id)
}

type oggPlayer struct {
	// SDL音频流
	stream *sdl.AudioStream
	// PCM音频数据
	audioData []byte
	// 当前播放位置
	dataPos int
	// 正在播放
	isPlaying bool
	// 是否循环播放
	loop bool
	// id
	id uint32
	// 音频规格
	sampleRate int32
	channels   int32
}

func newOggPlayer(soundFilePath string) (*oggPlayer, error) {
	file, err := os.Open(soundFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sound file, %v, %v", soundFilePath, err)
	}
	defer file.Close()

	oggReader, err := oggvorbis.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create oggvorbis reader, %v, %v", soundFilePath, err)
	}

	pcmData := make([]float32, 1024*1024)
	totalSamples := 0
	for {
		n, err := oggReader.Read(pcmData[totalSamples:])
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("failed to read oggvorbis data, %v, %v", soundFilePath, err)
		}
		if n == 0 {
			break
		}
		totalSamples += n
	}

	spec := &sdl.AudioSpec{
		Freq:     int32(oggReader.SampleRate()),
		Channels: int32(oggReader.Channels()),
		Format:   sdl.AudioF32,
	}

	callback := sdl.NewAudioStreamCallback(audioCallback)
	player := &oggPlayer{
		audioData:  float32ToBytes(pcmData[:totalSamples]),
		dataPos:    0,
		isPlaying:  false,
		loop:       false,
		sampleRate: int32(oggReader.SampleRate()),
		channels:   int32(oggReader.Channels()),
	}

	player.id = registerOGGPlayer(player)

	player.stream = sdl.OpenAudioDeviceStream(
		sdl.AudioDeviceDefaultPlayback,
		spec,
		callback,
		unsafe.Pointer(uintptr(player.id)),
	)

	if player.stream == nil {
		return nil, fmt.Errorf("failed to open audio stream: %s", sdl.GetError())
	}

	return player, nil
}

// 音频回调函数
func audioCallback(userdata unsafe.Pointer, stream *sdl.AudioStream, additionalAmount, totalAmount int32) {
	id := uint32(uintptr(userdata))
	player := getOGGPlayer(id)

	// 安全检查
	if player == nil || player.id != id || !player.isPlaying || len(player.audioData) == 0 {
		return
	}

	// 计算剩余数据量
	remaining := len(player.audioData) - player.dataPos
	if remaining <= 0 {
		if player.loop {
			player.dataPos = 0
			remaining = len(player.audioData)
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
		data := player.audioData[player.dataPos : player.dataPos+dataToSend]
		sdl.PutAudioStreamData(stream, (*uint8)(unsafe.Pointer(&data[0])), int32(dataToSend))
		player.dataPos += dataToSend
	}

	// 再次检查循环（如果刚好发送完所有数据）
	if player.loop && player.dataPos >= len(player.audioData) {
		player.dataPos = 0
	}
}

// 播放
func (p *oggPlayer) Play() {
	if p.stream == nil || p.id == 0 {
		return
	}
	p.isPlaying = true
	p.dataPos = 0
	sdl.ResumeAudioStreamDevice(p.stream)
}

// 暂停
func (p *oggPlayer) Pause() {
	if p.stream == nil || p.id == 0 {
		return
	}
	p.isPlaying = false
	sdl.AudioStreamDevicePaused(p.stream)
}

// 停止
func (p *oggPlayer) Stop() {
	if p.stream == nil || p.id == 0 {
		return
	}
	p.isPlaying = false
	p.dataPos = 0
	sdl.AudioStreamDevicePaused(p.stream)
	sdl.ClearAudioStream(p.stream)
}

// 设置循环播放
func (p *oggPlayer) SetLoop(loop bool) {
	p.loop = loop
}

// 关闭播放器，释放资源
func (p *oggPlayer) Close() {
	if p.stream != nil {
		p.Stop()
		sdl.DestroyAudioStream(p.stream)
		p.stream = nil
	}
	unregisterOGGPlayer(p.id)
}
