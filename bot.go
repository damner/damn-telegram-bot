package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/botanio/sdk/go"
	"github.com/tucnak/telebot"
)

type BotanMessage struct {
	usename string
}

var serviceUrl string

func main() {
	serviceUrl = strings.TrimRight(os.Getenv("DAMNRU_SERVICE_URL"), "/")
	log.Println(serviceUrl)

	botan1 := botan.New(os.Getenv("DAMNRU_APPMETRICA_TOKEN"))
	ch := make(chan bool)

	bot, err := telebot.NewBot(os.Getenv("DAMNRU_TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 1*time.Second)

	for message := range messages {
		log.Printf("Received a message from %s [%d] with the text: %s\n", message.Sender.Username, message.Sender.ID, message.Text)

		if message.Text == "/start" {
			bot.SendMessage(message.Chat, "Введи имя человека, которого ты хочешь обругать.", nil)
			continue
		}

		if message.Text == "/f" {
			bot.SendMessage(message.Chat, "После /f нужно указать имя бабы.", nil)
			continue
		}

		originalMessage := message.Text
		isFamale := false

		if len(message.Text) >= 2 {
			if message.Text[0:2] == "/f" {
				message.Text = message.Text[3:]
				isFamale = true
			}
		}

		damn := Generate(strings.Trim(message.Text, " "), isFamale)
		log.Println(damn)

		bot.SendMessage(message.Chat, damn, &telebot.SendOptions{
			ReplyMarkup: telebot.ReplyMarkup{
				Selective:      true,
				ResizeKeyboard: true,
				CustomKeyboard: [][]string{
					[]string{originalMessage},
				},
			},
		})

		botan1.TrackAsync(message.Sender.ID, BotanMessage{message.Sender.Username}, "test3", func(ans botan.Answer, err []error) {
			log.Printf("Event [%d] %+v\n", message.Sender.ID, ans)
			ch <- true
		})
	}

	<-ch
}

func Generate(name string, isFamale bool) string {
	values := url.Values{}
	values.Set("template", "$main_obozvat")
	values.Set("name", "{NAME}")
	values.Set("sex", "m")

	if isFamale {
		values.Set("sex", "w")
	}

	resp, err := http.Get(serviceUrl + "/create?" + values.Encode())
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	damn := string(body)
	damn = strings.Replace(damn, "{NAME}", name, -1)
	damn = regexp.MustCompile("\\^.").ReplaceAllStringFunc(damn, func(m string) string {
		return strings.ToUpper(m[1:])
	})

	return damn
}
