package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	uuid "github.com/satori/go.uuid"
	"os"
)

type Config struct {
	Token         string `yaml:"Token"`
	NotifyEnabled bool   `yaml:"notify_enabled"`
	BotToken      string `yaml:"bot_token"`
	ChatId        string `yaml:"chat_id"`
}

var Cfg = Config{
	Token:         uuid.NewV4().String(),
	NotifyEnabled: false,
	BotToken:      "",
	ChatId:        "",
}

var path = flag.String("conf", "config.yml", "")

func Init() error {
	f, err := os.Open(*path)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate default config
			f, err := os.Create(*path)
			if err != nil {
				return err
			}
			defer f.Close()
			return yaml.NewEncoder(f).Encode(&Cfg)
		}
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(&Cfg)
}
