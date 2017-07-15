package main

import (
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker"
	"log"
	"net/url"
	"sync"
)

func main() {

	// Create A Server
	wg := sync.WaitGroup{}
	wg.Add(1)
	client := broker.NewBroker("srv")
	client.Context.Debug = true
	u, _ := url.Parse("ws://localhost:8080")
	conn, _ := client.ConnectWebsocket(u, nil, nil, nil)
	m, _ := client.NewMsgObject("", indismqgo.ActionGET, "/test", nil, func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		//defer wg.Done()

		log.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	}).ToBuffer()
	conn.Send(m)
	wg.Wait()
}
