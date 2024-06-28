package services

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/lifenglsf/self_gpt_backend/config"
	"github.com/lifenglsf/self_gpt_backend/utils"
	"io"
	"log"
	"net/http"
	"strings"
)

type AliChat struct {
	echo.Context
}
type aliResponse struct {
	Output    output   `json:"output"`
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Usage     aliUsage `json:"usage"`
	RequestId string   `json:"request_id"`
	Msg       string   `json:"msg"`
}
type output struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}
type aliUsage struct {
	TotalTokens  int `json:"total_tokens"`
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (sc *AliChat) Gen(ver string) error {
	hostUrl = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	conf := config.GetAliConf()
	requstData := sc.Get("requestJson").(*Requests)
	model := requstData.Model
	if model == "" {
		model = "qwen1.5-0.5b-chat"
	}
	model = strings.ToLower(model)
	client := http.Client{}
	body := map[string]interface{}{
		"input": map[string]interface{}{
			"messages": requstData.Messages,
		},
		"model": model,
	}
	jsonData, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", hostUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.StreamOut(sc.Response(), `{"error_code":"500","msg":"new request error"}`)
		return nil
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.ApiSecret)
	//req.Header.Set("X-DashScope-SSE", "enable")
	resp, err := client.Do(req)
	if err != nil {
		utils.StreamOut(sc.Response(), `{"error_code":"500","msg":"get ali response failed"}`)
		return nil
	}
	answer := ""
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.StreamOut(sc.Response(), utils.FormatFailedMsg("read ali response failed"))
		return nil
	}
	w := sc.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	//log.Fatalln(string(respBody))
	var br aliResponse
	err = json.Unmarshal(respBody, &br)
	log.Println(string(respBody))
	if err != nil {
		utils.Stream(sc.Response(), utils.FormatFailedMsg("unmarshal ali response failed"))
		sc.Response().Flush()
		return nil
	}
	if br.Code != "" {
		utils.Stream(sc.Response(), utils.FormatFailedMsg("get ali response failed,"+br.Msg))
		sc.Response().Flush()
		return nil
	}
	answer += br.Output.Text
	log.Println(answer)
	utils.Stream(sc.Response(), answer)
	sc.Response().Flush()
	return nil
}
