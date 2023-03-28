package tts

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// TTSClient 是一个用于调用TensorFlowTTS REST API的客户端
type TTSClient struct {
	BaseURL string
}

// NewTTSClient 创建一个新的TTSClient
func NewTTSClient(baseURL string) *TTSClient {
	return &TTSClient{
		BaseURL: baseURL,
	}
}

// TTSRequest 是TTS请求的结构体
type TTSRequest struct {
	Text string `json:"text"`
}

// TTSResponse 是TTS响应的结构体
type TTSResponse struct {
	Audio []byte `json:"audio"`
}

// TextToSpeech 将给定的文本转换为音频
func (c *TTSClient) TextToSpeech(text string) ([]byte, error) {
	reqBody, err := json.Marshal(TTSRequest{
		Text: text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.BaseURL+"/api/tts", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// TextToSpeech 将给定的文本转换为音频并保存为音频文件
func (c *TTSClient) TextToSpeechSaveWav(text, filename string) error {
	audioData, err := c.TextToSpeech(text)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, audioData, 0644); err != nil {
		return err
	}
	return nil
}
