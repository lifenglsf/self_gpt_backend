package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type config struct {
	Database map[string]databaseConf `json:"database" mapstructure:"database"`
	Server   serverConf              `json:"server" mapstructure:"server"`
	GptConf  map[string]gpt          `json:"gpt_conf" mapstructure:"gpt_conf"`
}
type gpt struct {
	Appid     string `json:"appid" `
	ApiSecret string `json:"api_secret"`
	ApiKey    string `json:"api_key"`
}
type serverConf struct {
	Address string `json:"address"`
}
type databaseConf struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Extra       string `json:"extra"`
	Debug       string `json:"debug"`
	Charset     string `json:"charset"`
	MaxIdle     string `json:"max_idle"`
	MaxOpen     string `json:"max_open"`
	MaxLifeTime string `json:"max_life_time"`
}

var conf config

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("Unable to decode into struct: %s \n", err))
	}
}

func ReadDatabase() map[string]databaseConf {
	return conf.Database
}
func GetSparkConf() gpt {
	return conf.GptConf["spark"]
}
func GetBaiduConf() gpt {
	return conf.GptConf["baidu"]
}
func GetAliConf() gpt {
	return conf.GptConf["ali"]
}
func GetDeepConf() gpt {
	return conf.GptConf["deep"]
}
func GetServerConf() serverConf {
	return conf.Server
}

var aliModel = []string{
	"qwen1.5-0.5b-chat",
	"qwen-1.8b-chat",
	"Sambert系列模型",
	"paraformer-v1",
	"paraformer-8k-v1",
	"paraformer-mtl-v1",
	"paraformer-realtime-v1",
	"paraformer-realtime-8k-v1",
	"facechain-facedetect",
	"dolly-12b-v2",
}
