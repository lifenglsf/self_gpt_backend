package main

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lifenglsf/gpt/config"
	"github.com/lifenglsf/gpt/services"
	"log"
	"net/http"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "static")
	e.File("/", "./index.html")
	e.GET("sample", func(c echo.Context) error {
		fmt.Fprintf(c.Response(), "aaaaa")
		return nil
	})
	var chat services.ChatInterface
	spg := e.Group("/spark")
	spg.POST("/v1", func(c echo.Context) error {
		chat = &services.SparkChat{Context: c}
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		chat = &services.SparkChat{Context: c}
		return chat.Gen("v1")
	})
	spg.POST("/v2", func(c echo.Context) error {
		chat = &services.SparkChat{Context: c}
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		return chat.Gen("v2")
	})
	spg.POST("/v3", func(c echo.Context) error {
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		chat = &services.SparkChat{Context: c}
		return chat.Gen("v3")
	})
	spg.GET("/v35", func(c echo.Context) error {
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		chat = &services.SparkChat{Context: c}
		return chat.Gen("v3.5")
	})
	bag := e.Group("/api/baidu")
	bag.POST("/v1", func(c echo.Context) error {
		log.Println("cccc")
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			log.Println("validate error")
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
	//	msg.Messages = append(msg.Messages, services.Message{
	//		Role:    "user",
	//		Content: "算法",
	//	})
		c.Set("requestJson", msg)
		chat = &services.BaiduChat{Context: c}
		return chat.Gen("v1")
		return nil
	})
	ag := e.Group("/ali")
	ag.POST("/v1", func(c echo.Context) error {
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			log.Println("validate error")
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		chat = &services.AliChat{Context: c}
		return chat.Gen("v1")
	})
	dpg := e.Group("/deep")
	dpg.POST("/v1", func(c echo.Context) error {
		msg := new(services.Requests)
		if err := c.Bind(msg); err != nil {
			log.Println("validate error")
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(msg); err != nil {
			return err
		}
		c.Set("requestJson", msg)
		chat = &services.DeepChat{Context: c}
		return chat.Gen("v1")
	})
	server := config.GetServerConf()
	e.Logger.Fatal(e.Start(server.Address))
}
