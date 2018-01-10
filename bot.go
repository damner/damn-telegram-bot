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
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	template        = "$main_obozvat"
	moreButtonText  = "Давай ещё!"
	moreButtonAlert = "Готово"
	messageStart    = "Введи имя человека, которого ты хочешь обругать."
	messageF        = "После /f нужно указать имя бабы."
)

type BotanMessage struct {
	usename string
}

var serviceUrl = strings.TrimRight(os.Getenv("DAMNRU_SERVICE_URL"), "/")

var damnRegexp = regexp.MustCompile("\\^.")

func main() {
	log.Printf("Damn service URL: %s\n", serviceUrl)

	botan1 := botan.New(os.Getenv("DAMNRU_APPMETRICA_TOKEN"))

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
		damn := Generate(strings.Trim(c.Data, " "), false)

		bot.Edit(c.Message, c.Message.Text)

		logGeneratedDamn(c.Message, damn, &botan1)
		sendDamn(bot, c.Sender, damn, moreButton, c.Data)
		bot.Respond(c, &tb.CallbackResponse{})
	})

	bot.Handle(&moreFemaleButton, func(c *tb.Callback) {
		damn := Generate(strings.Trim(c.Data, " "), true)

		bot.Edit(c.Message, c.Message.Text)

		logGeneratedDamn(c.Message, damn, &botan1)
		sendDamn(bot, c.Sender, damn, moreFemaleButton, c.Data)
		bot.Respond(c, &tb.CallbackResponse{})
	})

	bot.Handle("/f", func(message *tb.Message) {
		if message.Payload == "" {
			bot.Send(message.Sender, messageF)
			return
		}

		damn := Generate(strings.Trim(message.Payload, " "), true)

		logGeneratedDamn(message, damn, &botan1)
		sendDamn(bot, message.Sender, damn, moreFemaleButton, message.Payload)
	})

	bot.Handle(tb.OnText, func(message *tb.Message) {
		damn := Generate(strings.Trim(message.Text, " "), false)

		logGeneratedDamn(message, damn, &botan1)
		sendDamn(bot, message.Sender, damn, moreButton, message.Text)
	})

	bot.Start()
}

func sendDamn(bot *tb.Bot, sender *tb.User, damn string, button tb.InlineButton, name string) {
	button.Data = name

	bot.Send(sender, damn, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			[]tb.InlineButton{button},
		},
	})
}

func logGeneratedDamn(message *tb.Message, damn string, botan1 *botan.Botan) {
	log.Println(damn)

	botan1.TrackAsync(message.Sender.ID, BotanMessage{message.Sender.Username}, "test3", func(ans botan.Answer, err []error) {
		log.Printf("Event [%d] %+v\n", message.Sender.ID, ans)
	})
}

func Generate(name string, isFamale bool) string {
	values := url.Values{}
	values.Set("template", template)
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
	damn = damnRegexp.ReplaceAllStringFunc(damn, func(m string) string {
		return strings.ToUpper(m[1:])
	})

	return damn
}
