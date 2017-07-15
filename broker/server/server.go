package main

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker"
	"log"
	"sync"
)

func main() {

	// Create A Server
	srv := broker.NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)
	srv.Context.Debug = true
	// Create A Handler for the /test path
	srv.Handlers.Set("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		//defer wg.Done()
		log.Println("/test message recieved")
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}

		fmt.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	})
	srv.ListenWebSocket("/", 8080)

}
