package main

import (
	"flag"
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

type BotanDamnData struct {
	usename string
	female  bool
	name    string
}

var serviceUrl = strings.TrimRight(os.Getenv("DAMNRU_SERVICE_URL"), "/")

var damnRegexp = regexp.MustCompile("\\^.")

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

	log.Printf("Damn service URL: %s\n", serviceUrl)

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
		isFemale := false

		damn := Generate(name, isFemale)
		logGeneratedDamn(c.Message.Sender, name, isFemale, damn)
		sendDamn(bot, c.Sender, damn, moreButton, c.Data)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle(&moreFemaleButton, func(c *tb.Callback) {
		name := strings.Trim(c.Data, " ")
		isFemale := true

		damn := Generate(name, isFemale)
		logGeneratedDamn(c.Message.Sender, name, isFemale, damn)
		sendDamn(bot, c.Sender, damn, moreFemaleButton, c.Data)

		bot.Respond(c, &tb.CallbackResponse{})
		bot.Edit(c.Message, c.Message.Text)
	})

	bot.Handle("/f", func(message *tb.Message) {
		if message.Payload == "" {
			bot.Send(message.Sender, messageF)
			return
		}

		name := strings.Trim(message.Payload, " ")
		isFemale := true

		damn := Generate(name, isFemale)
		logGeneratedDamn(message.Sender, name, isFemale, damn)
		sendDamn(bot, message.Sender, damn, moreFemaleButton, message.Payload)
	})

	bot.Handle(tb.OnText, func(message *tb.Message) {
		name := strings.Trim(message.Text, " ")
		isFemale := false

		damn := Generate(name, isFemale)
		logGeneratedDamn(message.Sender, name, isFemale, damn)
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

func logGeneratedDamn(sender *tb.User, name string, isFemale bool, damn string) {
	log.Println(damn)

	if len(botanLogger.Token) > 0 {
		data := BotanDamnData{
			usename: sender.Username,
			female:  isFemale,
			name:    name,
		}

		botanLogger.TrackAsync(sender.ID, data, "test4", func(ans botan.Answer, err []error) {
			log.Printf("Lot to AppMetrica data=%+v, answer=%+v, errors=%+v\n", data, ans, err)
		})
	}
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
