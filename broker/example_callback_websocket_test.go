package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker/websocket"
	"log"
	"net/url"
	"sync"
	"time"
)

func Example_callbackWebsocket() {

	// Create A Server
	srv := NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Create A Handler for the /test path
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}
		fmt.Println("Websocket Server Recieved", string(m.Fields.BodyBytes()))
		rep, err := srv.MakeReply(m, indismqgo.StatusOK, []byte("Hello Back From Server"))
		if err != nil {
			log.Fatal(err)
		} else if rep == nil {
			log.Fatal("failed to make reply")
		}
		err = c.Send(rep)
		if err != nil {
			log.Fatal(err)
			return err
		}
		time.Sleep(1 * time.Second)
		return nil
	})
	go websocket.ListenWebSocket(srv, "/", 8084, nil)
	//
	//Create a Client
	client := NewBroker("client")

	//connect to the server
	u, _ := url.Parse("ws://localhost:8084")
	conn, err := websocket.ConnectWebsocket(client, u, nil, nil, true)
	if err != nil {
		log.Fatal(err)
	}
	if conn == nil {
		log.Fatal("empty connection")
	}

	//send the server a test message with a callback from the server
	m, err := client.NewMsgObject("srv", indismqgo.ActionGET, "/test", []byte("Hello From Client"), func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		defer wg.Done()
		if string(m.Fields.From()) != "srv" {
			log.Fatal("Message Error")
		}
		fmt.Println("Websocket Client Recieved", string(m.Fields.BodyBytes()))
		return nil
	}).ToBuffer()
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
	//Websocket Server Recieved Hello From Client
	//Websocket Client Recieved Hello Back From Server

}
