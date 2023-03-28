package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type TTSRequest struct {
	Text string `json:"text"`
}

func main() {
	// 设置请求体
	ttsRequest := TTSRequest{
		Text: "There are many ways to make money, such as starting a business, investing in stocks, participating in online freelancing, or selling products or services. However, making money requires hard work, dedication, and skill development. It is important to identify your strengths, interests, and resources to determine the best way for you to make money. ",
	}
	requestBody, err := json.Marshal(ttsRequest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 发送请求
	resp, err := http.Post("http://127.0.0.1:5000/api/tts", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile("output.wav", body, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
