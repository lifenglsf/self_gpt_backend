package services

import "github.com/labstack/echo/v4"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatInterface interface {
	Gen(ver string) error
}
type ChatStruct struct {
	echo.Context
}
type Requests struct {
	Model    string    `json:"model" validate:"required"`
	Messages []Message `json:"messages" validate:"required"`
}
