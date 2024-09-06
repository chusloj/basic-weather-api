package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	pollInterval = time.Second * 5
)

const (
	endpoint = "https://api.open-meteo.com/v1/forecast" // ?latitude=52.52&longitude=13.41&current=temperature_2m,wind_speed_10m&hourly=temperature_2m,relative_humidity_2m,wind_speed_10m"
)

type Sender interface {
	Send(*Data) error
}

type SMSSender struct {
	number string
}

func NewSMSSender(number string) *SMSSender {
	return &SMSSender{number: number}
}

func (s *SMSSender) Send(data *Data) error {
	fmt.Println("sending weather to number: ", s.number)
	return nil
}

type EmailSender struct {
	email string
}

func NewEmailSender(email string) *EmailSender {
	return &EmailSender{email: email}
}

func (e *EmailSender) Send(data *Data) error {
	fmt.Println("sending weather to email: ", e.email)
	return nil
}

type Data struct {
	Elevation float64        `json:"elevation"`
	Hourly    map[string]any `json:"hourly"`
}

type WPoller struct {
	closech chan struct{}
	senders []Sender
}

func NewWPoller(senders []Sender) *WPoller {
	return &WPoller{
		closech: make(chan struct{}),
		senders: senders,
	}
}

func (wp *WPoller) start() {
	fmt.Println("Starting the WPoller (Weather Poller)")

	ticker := time.NewTicker(pollInterval)

main_loop:
	for {
		select {
		case <-ticker.C:
			data, err := getWeatherResults(52.52, 13.41) // args taken from commented-out part of endpoint
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Elevation: ", data.Elevation)
			if err := wp.handleData(data); err != nil {
				log.Fatal(err)
			}
		// graceful shutdown
		case <-wp.closech:
			break main_loop
		}
	}
	fmt.Println("WPoller stopped gracefully")
}

func (wp *WPoller) close() {
	fmt.Println("closing the channel")
	close(wp.closech)
}

func (wp *WPoller) handleData(data *Data) error {
	for _, s := range wp.senders {
		if err := s.Send(data); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func getWeatherResults(lat, long float64) (*Data, error) {
	uri := fmt.Sprintf("%s?latitude=%.2f&longitude=%.2f&hourly=temperature_2m", endpoint, lat, long)
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var data Data

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}

	return &data, nil
}

func main() {
	smsSender := NewSMSSender("719")
	emailSender := NewEmailSender("test@gmail.com")
	wpoller := NewWPoller([]Sender{smsSender, emailSender})
	go wpoller.start()

	timer := time.NewTimer(time.Second * 9)
	<-timer.C
	wpoller.close()
}
