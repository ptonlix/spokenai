package praudio

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"time"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/go-audio/audio"
	w "github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate        = 44100
	channelCount      = 1
	secondsToRecord   = 5
	bufferSize        = 2048
	bytesPerSample    = 2
	maxRecordDuration = 60 * time.Second
	maxRecordSize     = 1024 * 1024 * 10 // 10MB
)

type AudioRecorder struct {
	stream *portaudio.Stream
	dataCh chan []int32
}

func NewAudioRecorder() (*AudioRecorder, error) {
	recorder := &AudioRecorder{
		dataCh: make(chan []int32, 2),
	}
	stream, err := portaudio.OpenDefaultStream(channelCount, 0, sampleRate, bufferSize, func(in []int32) {
		data := make([]int32, len(in))
		copy(data, in)
		select {
		case recorder.dataCh <- data:
		default:
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %v", err)
	}

	recorder.stream = stream

	if err := recorder.stream.Start(); err != nil {
		return nil, fmt.Errorf("failed to start audio stream: %v", err)
	}

	return recorder, nil
}

func (ar *AudioRecorder) Read() []int32 {
	select {
	case data := <-ar.dataCh:
		return data
	default:
		return nil
	}
}

func (ar *AudioRecorder) Stop() {
	ar.stream.Stop()
	ar.stream.Close()
}

func record() ([]int32, error) {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buffer := make([]int32, sampleRate*channelCount*secondsToRecord)
	stream, err := portaudio.OpenDefaultStream(channelCount, 0, sampleRate, len(buffer), func(in []int32) {
		for i := range in {
			buffer[i] = in[i]
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %v", err)
	}

	err = stream.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start audio stream: %v", err)
	}
	defer stream.Close()

	time.Sleep(time.Duration(secondsToRecord) * time.Second)

	return buffer, nil
}

func int32SliceToIntSlice(input []int32) []int {
	output := make([]int, len(input))
	for i, v := range input {
		output[i] = int(v)
	}
	return output
}

func saveToWavFile(filename string, buffer []int) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	encoder := w.NewEncoder(outFile, sampleRate, 32, channelCount, 1)
	defer encoder.Close()

	audioBuffer := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: channelCount,
			SampleRate:  sampleRate,
		},
		Data:           buffer,
		SourceBitDepth: 32,
	}

	err = encoder.Write(audioBuffer)
	if err != nil {
		return fmt.Errorf("failed to write audio data to file: %v", err)
	}

	return nil
}

// RecordAndSaveWav 从麦克风录制音频并保存为WAV文件
func RecordAndSaveWav(filename string) error {
	buffer, err := record()
	if err != nil {
		return err
	}

	err = saveToWavFile(filename, int32SliceToIntSlice(buffer))
	if err != nil {
		return err
	}

	return nil
}

func recordWithSignal(ctx context.Context, maxDuration time.Duration) ([]int32, error) {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buffer := make([]int32, 0, sampleRate*channelCount*int(maxDuration.Seconds()))
	stream, err := portaudio.OpenDefaultStream(channelCount, 0, sampleRate, len(buffer), func(in []int32) {
		buffer = append(buffer, in...)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %v", err)
	}

	err = stream.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start audio stream: %v", err)
	}
	defer stream.Close()

	select {
	case <-ctx.Done():
	case <-time.After(maxDuration):
	}

	return buffer, nil
}

// RecordAndSaveWavWithInterrupt 从麦克风录制音频并保存为WAV文件，可以被信号中断
func RecordAndSaveWithInterrupt(filename string) error {
	ctx, cancel := context.WithCancel(context.Background())

	// 监听操作系统信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("接收到中断信号，停止录制...")
		cancel()
	}()

	buffer, err := recordWithSignal(ctx, maxRecordDuration)
	if err != nil {
		return err
	}

	err = saveToWavFile(filename, int32SliceToIntSlice(buffer))
	if err != nil {
		return err
	}

	return nil
}

func RecordAndSaveWithContext(ctx context.Context, filename string) error {
	portaudio.Initialize()
	defer portaudio.Terminate()
	done := make(chan struct{})

	// 初始化音频录制
	recorder, err := NewAudioRecorder()
	if err != nil {
		return fmt.Errorf("failed to initialize audio recorder: %v", err)
	}
	defer recorder.Stop()

	// 初始化进度条
	progressBar := pb.Full.Start(maxRecordSize)
	progressBar.SetRefreshRate(time.Millisecond * 200) // 设置刷新率
	progressBar.Set(pb.Bytes, true)                    // 显示录制音频的数据量

	// 记录开始时间
	startTime := time.Now()

	// 创建Context，用于取消录音
	ctxnew, cancel := context.WithCancel(ctx)

	// 开启协程进行录音
	samples := make([]int32, 0)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctxnew.Done():
				return
			case data := <-recorder.dataCh:
				// 将音频数据追加到samples
				samples = append(samples, data...)

				// 当达到最长录音时间时，取消录音
				if time.Since(startTime) >= maxRecordDuration {
					cancel()
					break
				}

				// 更新录制音频的数据量
				dataSize := int64(len(samples)) * int64(reflect.TypeOf(samples).Elem().Size())
				progressBar.Add(int(dataSize)) // 更新进度条
				progressBar.SetCurrent(dataSize)
			}
		}
	}()

	// 等待录音完成或接收到中断信号
	<-done

	progressBar.Finish()

	// 保存音频数据到WAV文件
	if err := saveToWavFile(filename, int32SliceToIntSlice(samples)); err != nil {
		return fmt.Errorf("failed to save audio to file: %v", err)
	}

	return nil
}

func RecordAndSaveWithChannel(filename string, sig <-chan struct{}) error {
	portaudio.Initialize()
	defer portaudio.Terminate()
	done := make(chan struct{})

	// 初始化音频录制
	recorder, err := NewAudioRecorder()
	if err != nil {
		return fmt.Errorf("failed to initialize audio recorder: %v", err)
	}
	defer recorder.Stop()

	// 初始化进度条
	progressBar := pb.Full.Start(maxRecordSize)
	progressBar.SetRefreshRate(time.Millisecond * 200) // 设置刷新率
	progressBar.Set(pb.Bytes, true)                    // 显示录制音频的数据量

	// 记录开始时间
	startTime := time.Now()

	// 创建Context，用于取消录音
	ctx, cancel := context.WithCancel(context.Background())

	// 开启协程进行录音
	samples := make([]int32, 0)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-recorder.dataCh:
				// 将音频数据追加到samples
				samples = append(samples, data...)

				// 当达到最长录音时间时，取消录音
				if time.Since(startTime) >= maxRecordDuration {
					cancel()
					break
				}

				// 更新录制音频的数据量
				dataSize := int64(len(samples)) * int64(reflect.TypeOf(samples).Elem().Size())
				progressBar.Add(int(dataSize)) // 更新进度条
				progressBar.SetCurrent(dataSize)
			}
		}
	}()

	// 等待录音完成或接收到中断信号
	select {
	case <-done:
	case <-sig:
		cancel()
		<-done
	}

	progressBar.Finish()

	// 保存音频数据到WAV文件
	if err := saveToWavFile(filename, int32SliceToIntSlice(samples)); err != nil {
		return fmt.Errorf("failed to save audio to file: %v", err)
	}

	return nil
}

func BackupWAVfile(filename string) error {
	return os.Rename(filename, filename+time.Now().Format("20060102150405"))
}

func RecordAndSaveWithInterruptShow(filename string) error {
	portaudio.Initialize()
	defer portaudio.Terminate()
	sig := make(chan os.Signal, 1)
	done := make(chan struct{})

	// 初始化音频录制
	recorder, err := NewAudioRecorder()
	if err != nil {
		return fmt.Errorf("failed to initialize audio recorder: %v", err)
	}
	defer recorder.Stop()

	// 初始化进度条
	progressBar := pb.Full.Start(maxRecordSize)
	progressBar.SetRefreshRate(time.Millisecond * 200) // 设置刷新率
	progressBar.Set(pb.Bytes, true)                    // 显示录制音频的数据量

	// 记录开始时间
	startTime := time.Now()

	// 捕获中断信号
	signal.Notify(sig, os.Interrupt)

	// 创建Context，用于取消录音
	ctx, cancel := context.WithCancel(context.Background())

	// 开启协程进行录音
	samples := make([]int32, 0)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-recorder.dataCh:
				// 将音频数据追加到samples
				samples = append(samples, data...)

				// 当达到最长录音时间时，取消录音
				if time.Since(startTime) >= maxRecordDuration {
					cancel()
					break
				}

				// 更新录制音频的数据量
				dataSize := int64(len(samples)) * int64(reflect.TypeOf(samples).Elem().Size())
				progressBar.Add(int(dataSize)) // 更新进度条
				progressBar.SetCurrent(dataSize)
			}
		}
	}()

	// 等待录音完成或接收到中断信号
	select {
	case <-done:
	case <-sig:
		cancel()
		<-done
	}

	progressBar.Finish()

	// 保存音频数据到WAV文件
	if err := saveToWavFile(filename, int32SliceToIntSlice(samples)); err != nil {
		return fmt.Errorf("failed to save audio to file: %v", err)
	}

	return nil
}

// PlayWavFile 播放指定的WAV文件
func PlayWavFile(filename string) error {
	// 打开WAV文件
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open WAV file: %v", err)
	}
	defer f.Close()

	// 解码WAV文件
	s, format, err := wav.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode WAV file: %v", err)
	}
	defer s.Close()

	// 初始化扬声器
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("failed to initialize speaker: %v", err)
	}

	// 播放音频
	done := make(chan struct{})
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))

	// 等待音频播放完成
	<-done

	return nil
}
