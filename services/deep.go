package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lifenglsf/self_gpt_backend/config"
	"github.com/lifenglsf/self_gpt_backend/utils"
	"io"
	"log"
	"net/http"
	"strings"
)

type DeepChat struct {
	echo.Context
}
type deepResponse struct {
	Id      string
	Choices []choices
	Created int64
	Model   string
	Object  string
	Usage   deepUsage
}
type choices struct {
	FinishReason string      `json:"finish_reason"`
	Index        int         `json:"index"`
	Message      deepMessage `json:"message"`
	Delta        deepDelta   `json:"delta"`
}
type deepDelta struct {
	deepMessage
}
type deepUsage struct {
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
type deepMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func (sc *DeepChat) Gen(ver string) error {
	hostUrl = "https://api.deepseek.com/chat/completions"
	conf := config.GetDeepConf()
	requstData := sc.Get("requestJson").(*Requests)
	model := requstData.Model
	if model == "" {
		model = "deepseek-chat"
	}
	model = strings.ToLower(model)
	client := http.Client{}
	body := map[string]interface{}{
		"messages":   requstData.Messages,
		"model":      model,
		"stream":     true,
		"max_tokens": 2048,
	}
	jsonData, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", hostUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.StreamOut(sc.Response(), `{"error_code":"500","msg":"new request error"}`)
		return nil
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.ApiKey)
	//req.Header.Set("X-DashScope-SSE", "enable")
	resp, err := client.Do(req)
	if err != nil {
		utils.StreamOut(sc.Response(), `{"error_code":"500","msg":"get deep response failed"}`)
		return nil
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	w := sc.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	for {
		line, err := reader.ReadString('\n')
		log.Println(line)
		if len(line) > 6 && line[:5] != "data:" {
			utils.StreamOut(sc.Response(), line)
			return nil
		}
		log.Println(line, err)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading response:", err)
			utils.StreamOut(sc.Response(), utils.FormatFailedMsg("Error reading response:"+err.Error()))
			return nil
		}
		if len(line) > 6 && line[:5] == "data:" {
			eventData := line[6:] // 移除 "data: " 前缀
			if eventData[:4] == "[DON" {
				break
			}
			var br deepResponse
			err = json.Unmarshal([]byte(eventData), &br)
			if err != nil {
				utils.Stream(sc.Response(), utils.FormatFailedMsg("unmarshal deep response failed"+err.Error()+eventData))
				sc.Response().Flush()
				return nil
			}
			for _, v := range br.Choices {
				if v.Delta.Content != "" {
					utils.Stream(sc.Response(), v.Delta.Content)
					sc.Response().Flush()
				}
			}
		}
	}
	return nil
}
