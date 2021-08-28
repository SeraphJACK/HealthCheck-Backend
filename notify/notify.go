package notify

import (
	"github.com/SeraphJACK/HealthCheck/config"
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

var bot *telebot.Bot
var chat *telebot.Chat

func Init() error {
	if !config.Cfg.NotifyEnabled {
		return nil
	}
	var err error
	bot, err = telebot.NewBot(telebot.Settings{
		Poller: &telebot.LongPoller{Timeout: time.Second * 10},
		Token:  config.Cfg.BotToken,
	})
	if err != nil {
		return err
	}
	go bot.Start()
	chat, err = bot.ChatByID(config.Cfg.ChatId)
	return err
}

func Notify(msg string, sound bool) {
	var err error
	if sound {
		_, err = bot.Send(chat, msg)
	} else {
		_, err = bot.Send(chat, msg, telebot.Silent)
	}
	if err != nil {
		log.Printf("Error sending notify message: %v, message:\n", err)
		log.Printf("%s\n", msg)
	}
}
