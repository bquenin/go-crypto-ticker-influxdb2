package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/bquenin/go-crypto-ticker-influxdb2/cmd/ticker/config"
	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/k0kubun/pp/v3"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	pp.Println(cfg)

	// Create new influxdb client with default option for server url authenticate by token
	influxdb := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	defer influxdb.Close()

	// User blocking write to the desired bucket
	writeAPI := influxdb.WriteAPIBlocking(cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)

	// Create a websocket to receive tick data from Coinbase
	var wsDialer websocket.Dialer
	ws, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	// Subscription message
	subscribe := coinbasepro.Message{
		Type:       "subscribe",
		ProductIds: []string{"BTC-USD", "ETH-USD"},
		Channels: []coinbasepro.MessageChannel{
			{
				Name: "ticker",
			},
		},
	}

	// Subscribe
	if err := ws.WriteJSON(subscribe); err != nil {
		log.Fatal().Err(err).Send()
	}

	// Read ticks from the websocket
	messages := make(chan coinbasepro.Message)
	go func() {
		defer close(messages)
		for {
			// Read tick data from websocket
			message := coinbasepro.Message{}
			if err := ws.ReadJSON(&message); err != nil {
				log.Error().Err(err).Send()
				break
			}
			log.Info().Str("product", message.ProductID).Str("price", message.Price).Send()

			messages <- message
		}
	}()

	// Write ticks data into InfluxDB
	for message := range messages {
		// Convert price to float
		price, err := strconv.ParseFloat(message.Price, 64)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}

		// Create point using full params constructor
		p := influxdb2.NewPoint("tick",
			map[string]string{"product": message.ProductID},
			map[string]interface{}{"price": price},
			time.Now())

		// Write point immediately
		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
			log.Error().Err(err).Send()
		}
	}
}
