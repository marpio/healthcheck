package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type attempt struct {
	ResponseTime time.Duration
	Error        string
}

type attempts []attempt

func main() {
	f, err := os.OpenFile("healthcheck.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(f)
	defer f.Close()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	const attemptsNumber = 5
	req, err := http.NewRequest("GET", os.Getenv("ENDPOINT_URL"), nil)
	checks := make(attempts, attemptsNumber)
	for i := 0; i < attemptsNumber; i++ {
		start := time.Now()
		resp, err := defaultHTTPClient.Do(req)
		checks[i].ResponseTime = time.Since(start)
		if err != nil {
			checks[i].Error = err.Error()
			continue
		}

		if resp.StatusCode != http.StatusAccepted {
			checks[i].Error = fmt.Errorf("response status %s", resp.Status).Error()
		}

		resp.Body.Close()
	}

	var numberOfErrors = 0
	for i := range checks {
		if checks[i].Error != "" || checks[i].ResponseTime > 10*time.Second {
			numberOfErrors++
			if numberOfErrors > 1 {
				log.Println("HappyHours funktioniert nicht! Aaaaaaa!!!")
				notify(`{"text":"HappyHours funktioniert nicht! Aaaaaaa!!!...."}`)
				break
			}
		}
	}
}

func notify(jsonStr string) {
	var incommingWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")
	var jsonReader = strings.NewReader(jsonStr)
	req, err := http.NewRequest("POST", incommingWebhookURL, jsonReader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		log.Print(err.Error())
	}
	defer resp.Body.Close()
}

var defaultHTTPClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   1,
		DisableCompression:    true,
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 5 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 10 * time.Second,
}
