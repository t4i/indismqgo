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

func Example_websocket() {

	// Create A Server
	srv := NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)
	// DefaultWsConnEvents.OnMessage = func(m *indismqgo.MsgBuffer, ws *WsConn) bool {
	// 	log.Println("OnMessage", string(m.Fields.From()), string(m.Fields.BodyBytes()))
	// 	return true
	// }
	// DefaultWsConnEvents.OnConnected = func(m *indismqgo.MsgBuffer, ws *WsConn) bool {
	// 	log.Println("OnConnected", string(m.Fields.From()), string(m.Fields.BodyBytes()))
	// 	return true
	// }
	// Create A Handler for the /test path
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10)
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}
		fmt.Println("Test Called")
		return nil
	})
	go websocket.ListenWebSocket(srv, "/", 8080, nil)
	time.Sleep(time.Millisecond * 100)
	//
	//Create a Client
	client := NewBroker("client")

	//connect to the server
	u, _ := url.Parse("ws://localhost:8080")
	conn, err := websocket.ConnectWebsocket(client, u, nil, nil, true)
	if err != nil {
		log.Fatal(err)
	}
	if conn == nil {

		log.Fatal("empty connection")
	}
	time.Sleep(time.Millisecond * 100)
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
	//Test Called

}
