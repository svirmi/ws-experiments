package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	endpoints = []string{
		"wss://testnet.binance.vision:9443/ws",
		"wss://stream.binance.com:9443/ws/btcusdt@aggTrade",
		// Add more endpoints as needed
	}
	numConnectionsPerEndpoint = 190 // Number of connections per endpoint
)

type SubscriptionMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}

type EndpointResponse struct {
	Endpoint string
	Message  []byte
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	responseChannel := make(chan EndpointResponse)

	wg := &sync.WaitGroup{}
	wg.Add(len(endpoints) * numConnectionsPerEndpoint)

	for _, endpoint := range endpoints {
		for i := 0; i < numConnectionsPerEndpoint; i++ {
			go connectToWebSocket(endpoint, responseChannel, wg)
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

func connectToWebSocket(endpoint string, responseChan chan<- EndpointResponse, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure WaitGroup is decremented when the function exits

	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	defer c.Close()

	subscriptionMsg := SubscriptionMessage{
		Method: "SUBSCRIBE",
		Params: []string{"btcusdt@aggTrade", "btcusdt@depth"},
		ID:     1,
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
			Endpoint: endpoint,
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
