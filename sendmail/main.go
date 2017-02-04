package main

import (
	"log"
	"os"

	"encoding/json"
	"io/ioutil"

	gomail "gopkg.in/gomail.v2"
)

type Config struct {
	SMTP string `json:"smtp_addr"`
	Port int    `json:"smtp_port"`
	From string `json:"from"`
	Auth bool   `json:"auth"`
	ID   string `json:"ID"` // for authentication
	Pass string `json:"pass"`
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal(os.Args[0], "[config]", "[to]", "[subject]", "[body]")
	}

	configPath, to, subject, body := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	b, err := ioutil.ReadFile(configPath)

	if err != nil {
		b, _ := json.Marshal(Config{})
		log.Fatalf("Config not found(template: %s)", string(b))
	}

	var config Config
	err = json.Unmarshal(b, &config)

	if err != nil {
		b, _ := json.Marshal(Config{})
		log.Fatalf("Json unmarshal failed(template: %s)", string(b))
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	var dialer *gomail.Dialer

	if config.Auth {
		dialer = gomail.NewDialer(config.SMTP, config.Port, config.ID, config.Pass)
	} else {
		dialer = &gomail.Dialer{Host: config.SMTP, Port: config.Port}
	}

	if err := dialer.DialAndSend(m); err != nil {
		panic(err)
	}
}
