package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
"title":"Retail Sales m\/m",
"country":"AUD",
"date":"2023-10-29T20:30:00-04:00",
"impact":"Medium","forecast":"0.3%",
"previous":"0.2%
*/
type NewsItem struct {
	Title          string `json:"title"`
	Country        string `json:"country"`
	Date           string `json:"date"`
	CurrencyImpact string `json:"impact"`
	Forecast       string `json:"forecast"`
	Previous       string `json:"previous"`
	Actual         string `json:"actual"`
}

func main() {
	url := "https://nfs.faireconomy.media/ff_calendar_thisweek.json"

	client := resty.New()

	resp, err := client.R().EnableTrace().Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch data: %v\n", err)
		return
	}
	if resp.StatusCode() == 200 {
		var newsData []NewsItem

		err := json.Unmarshal(resp.Body(), &newsData)
		if err != nil {
			fmt.Printf("Failed to parse JSON: %v\n", err)
			return
		}

		sort.SliceStable(newsData, func(i, j int) bool {
			return newsData[i].CurrencyImpact > newsData[j].CurrencyImpact
		})

		for _, news := range newsData {
			fmt.Printf("Event: %s\n", news.Title)
			fmt.Printf("Currency Impact: %s\n", news.CurrencyImpact)
			fmt.Printf("Date: %s\n", news.Date)
			fmt.Printf("Actual: %s\n", news.Actual)
			fmt.Printf("Forecast: %s\n", news.Forecast)
			fmt.Printf("Previous: %s\n\n", news.Previous)
		}

		bot, err := tgbotapi.NewBotAPI("5582287845:AAHf0r_Pi7oVYRfoPSezVrnt_tyM3F8VB-k")
		if err != nil {
			fmt.Printf("Failed To Connect to Bot: %v\n", err)
		}
		bot.Debug = true
		chatID := int64(196176954)

		for _, news := range newsData {
			message := "Top news by Currency Impact:\n\n"
			message += fmt.Sprintf("Event: %s\n", news.Title)
			message += fmt.Sprintf("Data: %s\n", news.Date)
			message += fmt.Sprintf("Currency Impact: %s\n", news.CurrencyImpact)
			message += fmt.Sprintf("Actual: %s\n", news.Actual)
			message += fmt.Sprintf("Forecast: %s\n", news.Forecast)
			message += fmt.Sprintf("Previous: %s\n\n", news.Previous)
			msg := tgbotapi.NewMessage(chatID, message)

			if _, err := bot.Send(msg); err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
			}

		}

	} else {
		fmt.Printf("Failed to fetch data. Status code: %d\n", resp.StatusCode())
	}
	time.Sleep(24 * time.Hour)

}
