package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"log"
	"net/url"
	"sync"
	"time"
)

func Example_websocket() {

	// Create A Server
	srv := NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Create A Handler for the /test path
	srv.Handlers.Set("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		defer wg.Done()
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}
		fmt.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	})
	go srv.ListenWebSocket("/", 8080)
	//
	//Create a Client
	client := NewBroker("client")

	//connect to the server
	u, _ := url.Parse("ws://localhost:8080")
	conn, err := client.ConnectWebsocket(u, nil, nil, client.AutoWsReconnect)
	if err != nil {
		log.Fatal(err)
	}
	if conn == nil {
		log.Fatal("empty connection")
	}

	//send the server a test message
	m, err := client.NewMsgObject("srv", indismqgo.ActionGET, "/test", []byte("Hello From Client"), nil).ToBuffer()
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Send(m)
	if err != nil {
		log.Fatal(err)
	}
	if indismqgo.WaitTimeout(&wg, 10*time.Second) {
		log.Fatal("timeout")
	}

	//Output:
	//Recieved Hello From Client

}
