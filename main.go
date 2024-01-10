package main

import (
	"chatgpt/utils"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	endpoints = []string{
		"wss://testnet.binance.vision/ws/%s@aggTrade",
		// "wss://stream.binance.com:9443/stream/%s@aggTrade",
	}
	numConnectionsPerEndpoint = 1 // Number of connections per endpoint
)

type SubscriptionMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

type EndpointResponse struct {
	Endpoint string
	Message  []byte
}

func main() {

	quoteAsset := "USDT"

	tPairs := utils.TradablePairs(quoteAsset)

	var symbols []string

	for _, tPair := range tPairs {
		symbols = append(symbols, strings.ToLower(tPair.Symbol))
	}

	// https://dev.binance.vision/t/how-to-create-binance-websocket-connections-to-all-coins-of-futures-market/16025/2

	symbols = symbols[:120]

	numConnections := len(symbols)

	fmt.Printf("Subscribed to %d pairs", numConnections)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	responseChannel := make(chan EndpointResponse)

	wg := &sync.WaitGroup{}
	wg.Add(len(endpoints) * numConnectionsPerEndpoint)

	for _, symbol := range symbols {
		for i := 0; i < numConnections; i++ {
			go connectToWebSocket(symbol, responseChannel, wg)
		}
	}

	go func() {
		<-interrupt
		fmt.Println("Shutting down...")

		wg.Wait() // Wait for all connections to close before exiting
		close(responseChannel)
	}()

	go aggregateResponses(responseChannel)

	select {}
}

func connectToWebSocket(symbol string, responseChan chan<- EndpointResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	stream := fmt.Sprintf(endpoints[0], symbol)

	fmt.Println(stream)

	c, _, err := websocket.DefaultDialer.Dial(stream, nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	defer c.Close()

	subscriptionMsg := SubscriptionMessage{
		Method: "SUBSCRIBE",
		// Params: []string{"btcusdt@aggTrade", "atomusdt@depth"},
		ID: rand.Intn(10000),
	}

	subscriptionJSON, err := json.Marshal(subscriptionMsg)
	if err != nil {
		log.Println("json marshal:", err)
		return
	}

	err = c.WriteMessage(websocket.TextMessage, subscriptionJSON)
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		responseChan <- EndpointResponse{
			Endpoint: symbol,
			Message:  message,
		}
	}
}

func aggregateResponses(responseChan <-chan EndpointResponse) {
	for response := range responseChan {
		fmt.Printf("Received message from %s: %s\n", response.Endpoint, string(response.Message))
		// Handle aggregated responses here
	}
}
