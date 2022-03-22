package financialbot

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"net/http"
	"strings"
)

var financialBot bot

type bot struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

type Payload struct {
	Message string
	Room    string
}

const (
	botName = "financial-bot"

	defaultURL         = "https://stooq.com/q/l/"
	defaultFormatQuery = "sd2t2ohlcv"
	defaultExportQuery = "csv"
	recordNotFound     = "N/D"
)

var (
	botErrMessage = errors.New("could not retrieve stock data")
)

// GetStockData gets the data of the given stock, if valid, and publishes the message to be consumed
func GetStockData(stock, room string) error {
	stockPrice, err := getStockPrice(stock)
	if err != nil {
		return botErrMessage
	}

	payload := generatePayload(stock, stockPrice, room)
	bPayload, err := json.Marshal(payload)
	if err != nil {
		return botErrMessage
	}

	err = financialBot.ch.Publish(
		"",                      // exchange
		financialBot.queue.Name, // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bPayload,
		})
	if err != nil {
		return botErrMessage
	}

	return nil
}

func getStockPrice(stock string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, defaultURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("s", stock)
	q.Set("f", defaultFormatQuery)
	q.Set("h", "")
	q.Set("e", defaultExportQuery)
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	r := csv.NewReader(res.Body)

	stockPrice, err := parseCSVRecords(r)
	if err != nil {
		return "", err
	}

	return stockPrice, nil
}

func parseCSVRecords(r *csv.Reader) (string, error) {
	records, err := r.ReadAll()
	if err != nil || len(records) != 2 || len(records[1]) < 4 {
		return "", fmt.Errorf("parse CSV: invalid format received")
	}

	return records[1][3], nil
}

func generatePayload(stock, stockPrice, room string) Payload {
	var messageToPublish string
	if stockPrice == recordNotFound {
		messageToPublish = fmt.Sprintf("%s: could not find data for stock %s", botName, strings.ToUpper(stock))
	} else {
		messageToPublish = fmt.Sprintf("%s: %s quote is $%s per share", botName, strings.ToUpper(stock), stockPrice)
	}

	return Payload{
		Message: messageToPublish,
		Room:    room,
	}
}
