package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"log"
	"net/url"
	"sync"
	"time"
)

func Example_brokerWebsocketHttp() {
	log.SetFlags(log.Llongfile)
	// Create A Server
	debug := false
	srv := NewBroker("srv")
	srv.Context.Debug = debug
	srv.Debug = debug
	wg := sync.WaitGroup{}
	wg.Add(1)
	go srv.ListenWebSocket("/", 8085)
	go srv.ListenHttp("/", 8086)
	//
	//Create a Client
	client := NewBroker("client1")
	client.Debug = debug
	client.Context.Debug = debug
	//connect to the server
	u, _ := url.Parse("ws://localhost:8085")
	client.ConnectWebsocket(u, nil, nil, client.AutoWsReconnect)
	client.Handlers.Set("/test", func(m *indismqgo.MsgBuffer, conn indismqgo.Connection) error {
		if string(m.Fields.From()) != "client2" {
			log.Fatal("Client Message Error")
		}
		fmt.Println("Client1 Recieved", string(m.Fields.BodyBytes()))
		rep, err := client.MakeReply(m, indismqgo.StatusOK, []byte("Hello Back From Client1"))
		if err != nil {
			log.Fatal(err)
		} else if rep == nil {
			log.Fatal("failed to make reply")
		}
		err = conn.Send(rep)
		if err != nil {
			log.Fatal(err)
			return err
		}

		return nil
	})
	//
	//Create a Second Client on HTTP
	client2 := NewBroker("client2")
	client2.Debug = debug
	client2.Context.Debug = debug
	u2, _ := url.Parse("http://localhost:8086")
	conn2 := client2.NewHttpConn(u2, nil)

	//send the server a test message to client 1 and expect a reply
	m, _ := client2.NewMsgObject("client1", indismqgo.ActionGET, "/test", []byte("Hello From Client 2"), func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		defer wg.Done()
		if string(m.Fields.From()) != "client1" {
			log.Fatal("Message Error")
		}
		fmt.Println("Client2 Client Recieved", string(m.Fields.BodyBytes()))
		return nil
	}).ToBuffer()
	conn2.Send(m)
	if indismqgo.WaitTimeout(&wg, 10*time.Second) {
		log.Fatal("timeout")
	}
	//Output:
	// Client1 Recieved Hello From Client 2
	// Client2 Client Recieved Hello Back From Client1

}
