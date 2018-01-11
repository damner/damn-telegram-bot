package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/botanio/sdk/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	messageStart           = "Введи имя человека, которого ты хочешь обругать."
	buttonMoreText         = "Давай ещё!"
	buttonChangeGenderText = "Сменить пол"
)

type BotanDamnData struct {
	usename string
	name    string
	gender  Gender
}

var botanLogger botan.Botan

var buttons struct {
	more         tb.InlineButton
	changeGender tb.InlineButton
}

var service *Service

func main() {
	var noBotan bool
	flag.BoolVar(&noBotan, "no-botan", false, "disable logging to AppMetrica")
	flag.Parse()

	var botanToken string
	if noBotan == false {
		botanToken = os.Getenv("DAMNRU_APPMETRICA_TOKEN")
		if len(botanToken) == 0 {
			log.Fatalln("Env variable DAMNRU_APPMETRICA_TOKEN is not set")
		}
	}

	botanLogger = botan.New(botanToken)

	service = NewDamnService(os.Getenv("DAMNRU_SERVICE_URL"))
	log.Printf("Damn service URL: %s\n", service.URL)

	bot, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("DAMNRU_TELEGRAM_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalln(err)
	}

	buttons.more = tb.InlineButton{
		Unique: "more",
		Text:   buttonMoreText,
	}

	buttons.changeGender = tb.InlineButton{
		Unique: "change_gender",
		Text:   buttonChangeGenderText,
	}

	bot.Handle("/start", func(message *tb.Message) {
		bot.Send(message.Sender, messageStart)
	})

	bot.Handle(&buttons.more, func(c *tb.Callback) {
		name := c.Data[1:]
		gender := Gender(c.Data[0:1])

		responseWithDamn(bot, c.Sender, name, gender)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(&buttons.changeGender, func(c *tb.Callback) {
		name := c.Data[1:]
		gender := Gender(c.Data[0:1])

		responseWithDamn(bot, c.Sender, name, gender.Another())

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(tb.OnText, func(message *tb.Message) {
		responseWithDamn(bot, message.Sender, message.Text, GenderMale)
	})

	bot.Start()
}

func responseWithDamn(bot *tb.Bot, user *tb.User, name string, gender Gender) {
	damn := service.Generate(name, gender)
	sendDamn(bot, user, damn)
	logDamn(user, damn)
}

func sendDamn(bot *tb.Bot, user *tb.User, damn *Damn) {
	buttons.more.Data = string(damn.Gender) + damn.Name
	buttons.changeGender.Data = string(damn.Gender) + damn.Name

	bot.Send(user, damn.Result, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			[]tb.InlineButton{
				buttons.more,
				buttons.changeGender,
			},
		},
	})
}

func logDamn(user *tb.User, damn *Damn) {
	log.Printf("UID: %d, Username: %s, Result: %s\n", user.ID, user.Username, damn.Result)

	if len(botanLogger.Token) > 0 {
		data := BotanDamnData{
			usename: user.Username,
			name:    damn.Name,
			gender:  damn.Gender,
		}

		botanLogger.TrackAsync(user.ID, data, "test4", func(ans botan.Answer, err []error) {
			log.Printf("Lot to AppMetrica data=%+v, answer=%+v, errors=%+v\n", data, ans, err)
		})
	}
}
