package broker

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker/http"
	"github.com/t4i/indismqgo/broker/websocket"
)

func Example_brokerWebsocketHttp() {

	// Create A Server
	debug := false
	if debug {
		log.SetFlags(log.Llongfile)
	}
	srv := NewBroker("srv")
	srv.Debug(&debug)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go websocket.NewImqWsServer(srv, "/", 8085, nil, nil).ListenAndServe()
	go http.ListenHttp(srv, "/", 8086, nil)
	//
	//Create a Clientd
	client := NewBroker("client1")
	client.Debug(&debug)
	//connect to the server
	u, _ := url.Parse("ws://localhost:8085")
	client1Conn, _ := websocket.ConnectWebsocket(client, u, nil, nil, true)

	if m, err := client.NewConnectionMsg(nil, nil, nil); err == nil {
		client1Conn.Send(m)
	}

	client.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, conn indismqgo.Connection) error {
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
	client2.Debug(&debug)
	u2, _ := url.Parse("http://localhost:8086")
	conn2 := http.NewHttpConn(client2, u2, nil, nil)
	if m, err := client2.NewConnectionMsg(nil, nil, nil); err == nil {
		if err := conn2.Send(m); err != nil {
			log.Println(err)
		}
	}
	//http.ConnectHttp(client2, conn2)
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
	if indismqgo.WaitTimeout(&wg, 2*time.Second) {
		log.Fatal("timeout")
	}
	//Output:
	// Client1 Recieved Hello From Client 2
	// Client2 Client Recieved Hello Back From Client1

}
