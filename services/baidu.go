package services

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lifenglsf/gpt/config"
	"github.com/lifenglsf/gpt/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type BaiduChat struct {
	echo.Context
}
type baiduResponse struct {
	Id               string     `json:"id"`
	Object           string     `json:"object"`
	Created          int        `json:"created"`
	SentenceId       int        `json:"sentence_id"`
	IsEnd            bool       `json:"is_end"`
	IsTruncated      bool       `json:"is_truncated"`
	Result           string     `json:"result"`
	MeedClearHistory bool       `json:"meed_clear_history"`
	BanRound         int        `json:"ban_round"`
	Usage            baiduUsage `json:"usage"`
	ErrorCode        int        `json:"error_code"`
	ErrorMsg         string     `json:"error_msg"`
}
type baiduUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

var hostUrl = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/"

func generateSignature(ak, sk string, req *http.Request) string {
	// 获取当前时间戳
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	// 构造签名字符串
	var sb strings.Builder
	sb.WriteString(req.Method)
	sb.WriteString("\n")
	sb.WriteString(req.URL.Path)
	sb.WriteString("\n")
	sb.WriteString(timestamp)
	sb.WriteString("\n")
	sb.WriteString("image=BASE64_ENCODED_IMAGE_DATA") // 注意将 BASE64_ENCODED_IMAGE_DATA 替换为实际的图片数据

	// 计算签名
	h := hmac.New(sha256.New, []byte(sk))
	h.Write([]byte(sb.String()))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构造Authorization头部内容
	var auth strings.Builder
	auth.WriteString("bce-auth-v1/")
	auth.WriteString(ak)
	auth.WriteString("/")
	auth.WriteString(timestamp)
	auth.WriteString("/1800") // 设置有效期，单位秒，这里设置为30分钟

	auth.WriteString("/")
	auth.WriteString(signature)

	return auth.String()
}
func (sc *BaiduChat) getAccessToken(ak, sk string) (error, string) {
	url := "https://aip.baidubce.com/oauth/2.0/token?client_id=" + ak + "&client_secret=" + sk + "&grant_type=client_credentials"
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	var r map[string]interface{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err, ""
	}
	return nil, r["access_token"].(string)
}

//	func sign() {
//		conf := config.GetBaiduConf()
//		authStringPrefix := "bce-auth-v1/" + conf.ApiKey + "/" + time.Now().Format(time.RFC3339) + "/7200"
//		header:=map[string]string{
//			"Host":"bj.bcebos.com",
//			"Content"
//		}
//	}
func (sc *BaiduChat) Gen(ver string) error {
	conf := config.GetBaiduConf()
	err, token := sc.getAccessToken(conf.ApiKey, conf.ApiSecret)
	if err != nil {
		return fmt.Errorf("%s%s", "get access token failed:", err)
	}
	requestData := sc.Get("requestJson").(*Requests)
	model := requestData.Model
	if model == "" {
		model = "ERNIE-Lite-8K"
	}
	model = strings.ToLower(model)
	urls := hostUrl + model + "?access_token=" + token
	log.Println(urls)
	messages := requestData.Messages
	if messages[0].Role == "system" {
		messages = messages[1:]
	}
	log.Println(messages)
	client := http.Client{}
	body := map[string]interface{}{
		//"messages": requstData.Messages,
		"messages": messages,
		"stream":   true,
	}
	jsonData, err := json.Marshal(body)
	log.Println(1)
	if err != nil {
		utils.StreamOut(sc.Response(), utils.FormatFailedMsg("new request marshal error"))
		return nil
	}
	req, err := http.NewRequest("POST", urls, bytes.NewBuffer(jsonData))
	log.Println(2)
	if err != nil {
		utils.StreamOut(sc.Response(), utils.FormatFailedMsg("new request error"))
		return nil
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-bce-date", time.Now().Format(time.RFC1123Z))
	req.Header.Set("Authorization", generateSignature(conf.ApiKey, conf.ApiSecret, req))
	resp, err := client.Do(req)
	if err != nil {
		utils.StreamOut(sc.Response(), utils.FormatFailedMsg("get baidu response failed:"+err.Error()))
		return nil
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	w := sc.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	//w.WriteHeader(201)
	//	w.Header().Set("Connection", "keep-alive")
	for {
		line, err := reader.ReadString('\n')
		log.Println(line)
		//if line[0] == '{' {
		//	utils.StreamOut(sc.Response(), line)
		//	return nil
		//}
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
			//log.Println(eventData)
			//if eventData[:4] == "[DON" {
			//	break
			//}
			var br baiduResponse
			err = json.Unmarshal([]byte(eventData), &br)
			if err != nil {
				utils.StreamOut(sc.Response(), utils.FormatFailedMsg("unmarshal baidu response failed"+err.Error()+eventData))
				return nil
			}
			if br.IsEnd {
				utils.Stream(sc.Response(), br.Result+"[DONE]")
			} else {
				utils.Stream(sc.Response(), br.Result)
			}
			//event := Event{
			//	Data: []byte(br.Result),
			//}
			//if err := event.MarshalTo(sc.Response()); err != nil {
			//	return err
			//}
			sc.Response().Flush()
			//utils.StreamOut(sc.Response(), br.Result)
		}
	}
	return nil
}
