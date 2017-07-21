package main

import (
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker"
	"github.com/t4i/indismqgo/broker/websocket"
	"log"
	"net/url"
	"sync"
)

func main() {

	// Create A Server
	wg := sync.WaitGroup{}
	wg.Add(1)
	client := broker.NewBroker("srv")
	debug := true
	client.Debug(&debug)
	u, _ := url.Parse("ws://localhost:8080")
	conn, _ := websocket.ConnectWebsocket(client, u, nil, nil, false)
	m, _ := client.NewMsgObject("", indismqgo.ActionGET, "/test", nil, func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		//defer wg.Done()

		log.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	}).ToBuffer()
	conn.Send(m)
	wg.Wait()
}
