package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const template string = "$main_obozvat"

type Gender string

const (
	GenderMale   Gender = "m"
	GenderFemale Gender = "w"
)

func (g Gender) Another() Gender {
	if string(g) == string(GenderMale) {
		return GenderFemale
	}

	return GenderMale
}

// Damn service.
type Service struct {
	URL string
}

// Generated damn.
type Damn struct {
	Name   string
	Gender Gender
	// Generated text
	Result string
}

var damnRegexp = regexp.MustCompile("\\^.")

func NewDamnService(URL string) *Service {
	return &Service{
		URL: strings.TrimRight(URL, "/"),
	}
}

func (s *Service) Generate(name string, gender Gender) *Damn {
	damn := &Damn{
		Name:   name,
		Gender: gender,
	}

	result := s.request("{NAME}", gender)
	result = strings.Replace(result, "{NAME}", name, -1)
	result = damnRegexp.ReplaceAllStringFunc(result, func(m string) string {
		return strings.ToUpper(m[1:])
	})

	damn.Result = result

	return damn
}

func (s *Service) request(name string, gender Gender) string {
	values := url.Values{}
	values.Set("template", template)
	values.Set("name", strings.TrimSpace(name))
	values.Set("sex", string(gender))

	log.Println(template)

	resp, err := http.Get(s.URL + "/create?" + values.Encode())
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}
