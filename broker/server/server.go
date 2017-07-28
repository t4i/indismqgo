package main

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker"
	"github.com/t4i/indismqgo/broker/websocket"
	"log"
	"sync"
)

func main() {

	// Create A Server
	srv := broker.NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)
	debug := true
	srv.Debug(&debug)
	// Create A Handler for the /test path
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		//defer wg.Done()
		log.Println("/test message recieved")
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}

		fmt.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	})
	websocket.ListenWebSocket(srv, "/", 8080, nil)

}
