package main

import (
	"log"
	"net/url"
	"time"

	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker"
	"github.com/t4i/indismqgo/broker/websocket"
)

func main() {
	log.Println("running")
	// Create A Server
	srv := broker.NewBroker("srv")
	// Create A Handler for the /test path
	done := make(chan bool)
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		//defer wg.Done()
		done <- true
		return nil
	})
	go websocket.NewImqWsServer(srv, "/", 8085, nil, nil).ListenAndServe()
	//
	//Create a Client
	client := broker.NewBroker("client")

	//connect to the server

	u, _ := url.Parse("ws://localhost:8085")
	conn, err := websocket.ConnectWebsocket(client, u, nil, nil, false)

	if err != nil {
		log.Fatal(err)
	}
	if conn == nil {
		log.Fatal("empty connection")
	}
	quant := 100000
	start := time.Now()
	for i := 0; i < quant; i++ {
		m, _ := client.NewMsgObject("srv", indismqgo.ActionGET, "/test", []byte("Hello From Client Benchmark"), nil).ToBuffer()
		conn.Send(m)
		<-done
	}
	elapsed := time.Since(start)
	log.Println(elapsed.Nanoseconds() / 1000000000)
	log.Println(elapsed)
	//persecond := int64(quant) / (elapsed.Nanoseconds() / 1e6 / 1000)
	log.Println(float64(quant) / elapsed.Seconds())
}
