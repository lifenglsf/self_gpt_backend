package utils

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type failed struct {
	Code int    `json:"error_code"`
	Msg  string `json:"msg"`
}

func Stream(resp *echo.Response, msg string) error {
	event := Event{
		Data: []byte(msg),
	}
	if err := event.MarshalTo(resp); err != nil {
		return err
	}
	return nil
}
func StreamOut(resp *echo.Response, msg string) {
	//sd := bytes.Split([]byte(msg), []byte("\n"))
	//for i := range sd {
	//	fmt.Fprintf(resp, "data: %s\n", sd[i])
	//	resp.Flush()
	//}
	resp.Header().Set(echo.HeaderContentType, "text/event-stream")
	resp.WriteHeader(http.StatusOK)
	for _, l := range msg {
		fmt.Fprintf(resp, "%c", l)
		resp.Flush()
		time.Sleep(time.Millisecond * 20)
	}
}

func FormatFailedMsg(msg string) string {
	//if code == 0 {
	//	code = 500
	//}
	f := failed{
		Code: 500,
		Msg:  msg,
	}
	marshal, err := json.Marshal(f)
	if err != nil {
		return `{"code":500,"msg":"system error"}`
	}
	return string(marshal)
}
