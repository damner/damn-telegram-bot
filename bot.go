package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/botanio/sdk/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	template        = "$main_obozvat"
	moreButtonText  = "Давай ещё!"
	moreButtonAlert = "Готово"
	messageStart    = "Введи имя человека, которого ты хочешь обругать."
	messageF        = "После /f нужно указать имя бабы."
)

type BotanDamnData struct {
	usename string
	name    string
	gender  Gender
}

var botanLogger botan.Botan

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

	url := os.Getenv("DAMNRU_SERVICE_URL")
	log.Printf("Damn service URL: %s\n", url)
	service := NewDamnService(url)

	bot, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("DAMNRU_TELEGRAM_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalln(err)
	}

	bot.Handle("/start", func(message *tb.Message) {
		bot.Send(message.Sender, messageStart)
	})

	moreButton := tb.InlineButton{
		Unique: "more",
		Text:   moreButtonText,
	}

	moreFemaleButton := tb.InlineButton{
		Unique: "more_female",
		Text:   moreButtonText,
	}

	bot.Handle(&moreButton, func(c *tb.Callback) {
		name := strings.Trim(c.Data, " ")
		gender := GenderMale

		damn := service.Generate(name, gender)
		logGeneratedDamn(c.Message.Sender, damn)
		sendDamn(bot, c.Sender, damn, moreButton)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(&moreFemaleButton, func(c *tb.Callback) {
		name := strings.Trim(c.Data, " ")
		gender := GenderFemale

		damn := service.Generate(name, gender)
		logGeneratedDamn(c.Message.Sender, damn)
		sendDamn(bot, c.Sender, damn, moreFemaleButton)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle("/f", func(message *tb.Message) {
		if message.Payload == "" {
			bot.Send(message.Sender, messageF)
			return
		}

		name := strings.Trim(message.Payload, " ")
		gender := GenderFemale

		damn := service.Generate(name, gender)
		logGeneratedDamn(message.Sender, damn)
		sendDamn(bot, message.Sender, damn, moreFemaleButton)
	})

	bot.Handle(tb.OnText, func(message *tb.Message) {
		name := strings.Trim(message.Text, " ")
		gender := GenderMale

		damn := service.Generate(name, gender)
		logGeneratedDamn(message.Sender, damn)
		sendDamn(bot, message.Sender, damn, moreButton)
	})

	bot.Start()
}

func sendDamn(bot *tb.Bot, sender *tb.User, damn *Damn, button tb.InlineButton) {
	button.Data = damn.Name

	bot.Send(sender, damn.Result, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			[]tb.InlineButton{
				button,
				button,
			},
		},
	})
}

func logGeneratedDamn(sender *tb.User, damn *Damn) {
	log.Println(damn.Result)

	if len(botanLogger.Token) > 0 {
		data := BotanDamnData{
			usename: sender.Username,
			name:    damn.Name,
			gender:  damn.Gender,
		}

		botanLogger.TrackAsync(sender.ID, data, "test4", func(ans botan.Answer, err []error) {
			log.Printf("Lot to AppMetrica data=%+v, answer=%+v, errors=%+v\n", data, ans, err)
		})
	}
}
