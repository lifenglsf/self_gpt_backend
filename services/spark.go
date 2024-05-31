package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/lifenglsf/gpt/config"
	"github.com/lifenglsf/gpt/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SparkChat struct {
	echo.Context
}

func (sc *SparkChat) Gen(ver string) error {
	conf := config.GetSparkConf()
	appid := conf.Appid
	apiKey := conf.ApiKey
	apiSecret := conf.ApiSecret
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(assembleAuthUrl1(ver, apiKey, apiSecret), nil)
	if err != nil {
		return fmt.Errorf("%s%s", "ws dial failed", err)
	} else if resp.StatusCode != 101 {
		return fmt.Errorf("%s%s", "ws status code !=101", err)
	}

	go func() {
		requestData := sc.Get("requestJson").(*Requests)
		data := genParams1(ver, appid, requestData.Messages)
		conn.WriteJSON(data)

	}()
	var answer = ""
	//获取返回的数据
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			utils.StreamOut(sc.Response(), utils.FormatFailedMsg("read message error:"+err.Error()))
			return nil
		}

		var data map[string]interface{}
		err1 := json.Unmarshal(msg, &data)
		if err1 != nil {
			utils.StreamOut(sc.Response(), utils.FormatFailedMsg("parse json error:%s"+err1.Error()))
			return nil
		}
		//fmt.Println(string(msg))
		//解析数据
		payload := data["payload"].(map[string]interface{})
		choices := payload["choices"].(map[string]interface{})
		header := data["header"].(map[string]interface{})
		code := header["code"].(float64)

		if code != 0 {
			//fmt.Println(data["payload"])
			return fmt.Errorf("%s%s", "code不是0", data["payload"])
		}
		status := choices["status"].(float64)
		//fmt.Println(status)
		text := choices["text"].([]interface{})
		content := text[0].(map[string]interface{})["content"].(string)
		if status != 2 {
			log.Println(content)
			utils.StreamOut(sc.Response(), content)
			answer += content
		} else {
			fmt.Println("收到最终结果")
			utils.StreamOut(sc.Response(), content)
			answer += content
			usage := payload["usage"].(map[string]interface{})
			temp := usage["text"].(map[string]interface{})
			totalTokens := temp["total_tokens"].(float64)
			fmt.Println("total_tokens:", totalTokens)
			conn.Close()
			break
		}
	}
	return nil
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl1(ver, apiKey, apiSecret string) string {
	hostUrl := "wss://spark-api.xf-yun.com/v1.1/chat"
	if ver == "v2" {
		hostUrl = "wss://spark-api.xf-yun.com/v2.1/chat"
	} else if ver == "v3" {
		hostUrl = "wss://spark-api.xf-yun.com/v3.1/chat"
	} else if ver == "v3.5" {
		hostUrl = "wss://spark-api.xf-yun.com/v3.5/chat"
	}
	log.Println(ver, hostUrl)
	ul, err := url.Parse(hostUrl)
	if err != nil {
		fmt.Println("parse hosturl failed", err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	// fmt.Println(sgin)
	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	// fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hostUrl + "?" + v.Encode()
	return callurl
}

func HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func readResp(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("code=%d,body=%s", resp.StatusCode, string(b))
}
func genParams1(ver, appid string, msg []Message) map[string]interface{} { // 根据实际情况修改返回的数据结构和字段名
	domain := "general"
	if ver == "v2" {
		domain = "generalv2"
	} else if ver == "v3" {
		domain = "generalv3"
	} else if ver == "v3.5" {
		domain = "generalv3.5"
	}
	data := map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
		"header": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"app_id": appid, // 根据实际情况修改返回的数据结构和字段名
		},
		"parameter": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"chat": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
				"domain":      domain,     // 根据实际情况修改返回的数据结构和字段名
				"temperature": 0.5,        // 根据实际情况修改返回的数据结构和字段名
				"top_k":       int64(4),   // 根据实际情况修改返回的数据结构和字段名
				"max_tokens":  int64(150), // 根据实际情况修改返回的数据结构和字段名
				"auditing":    "default",  // 根据实际情况修改返回的数据结构和字段名
			},
		},
		"payload": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"message": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
				"text": msg, // 根据实际情况修改返回的数据结构和字段名
			},
		},
	}

	return data // 根据实际情况修改返回的数据结构和字段名
}
