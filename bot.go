package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/botanio/sdk/go"
	"github.com/opennota/morph"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	messageStart           = "Введи имя человека, которого ты хочешь обругать."
	buttonMoreText         = "Давай ещё!"
	buttonChangeGenderText = "Сменить пол"
)

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

		logEvent(message.Sender, "start", struct{}{})
	})

	bot.Handle(&buttons.more, func(c *tb.Callback) {
		name := c.Data[1:]
		gender := Gender(c.Data[0:1])

		logEvent(c.Sender, "button-more", struct {
			name   string
			gender Gender
		}{
			name:   name,
			gender: gender,
		})

		responseWithDamn(bot, c.Sender, name, gender)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(&buttons.changeGender, func(c *tb.Callback) {
		name := c.Data[1:]
		gender := Gender(c.Data[0:1])

		logEvent(c.Sender, "button-change-gender", struct {
			name   string
			gender Gender
		}{
			name:   name,
			gender: gender,
		})

		responseWithDamn(bot, c.Sender, name, gender.Another())

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(tb.OnText, func(message *tb.Message) {
		logEvent(message.Sender, "text", struct {
			name string
		}{
			name: message.Text,
		})

		gender := getGender(message.Text)

		responseWithDamn(bot, message.Sender, message.Text, gender)
	})

	bot.Start()
}

func getGender(name string) Gender {
	gender := GenderMale

	_, _, tags := morph.Parse(strings.ToLower(name))

	if len(tags) > 0 {
		if strings.Contains(tags[0], "femn") {
			gender = GenderFemale
		}
	}

	return gender
}

func responseWithDamn(bot *tb.Bot, user *tb.User, name string, gender Gender) {
	damn := service.Generate(name, gender)
	sendDamn(bot, user, damn)

	logEvent(user, "damn", struct {
		name   string
		gender Gender
	}{
		name:   damn.Name,
		gender: damn.Gender,
	})
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

func logEvent(user *tb.User, name string, values interface{}) {
	log.Printf("%v %v %+v\n", user.ID, name, values)

	if len(botanLogger.Token) > 0 {
		botanLogger.TrackAsync(user.ID, values, name, func(ans botan.Answer, err []error) {
			log.Printf("Log to AppMetrica userID=%v, name=%v, values=%+v, answer=%+v, errors=%+v\n", user.ID, name, values, ans, err)
		})
	}
}
