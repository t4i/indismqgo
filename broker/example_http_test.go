package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker/http"
	"log"
	h "net/http"
	"net/url"
	"sync"
	"time"
)

func Example_http() {
	debug := false
	// Create A Server
	srv := NewBroker("srv")
	srv.Debug(&debug)
	wg := sync.WaitGroup{}
	wg.Add(1)
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		defer wg.Done()
		if string(m.Fields.From()) != "client" {
			log.Fatal("Message Error")
		}
		fmt.Println("Recieved", string(m.Fields.BodyBytes()))
		return nil
	})
	ev := &http.Events{}
	ev.OnBeforeMessage = func(r *h.Request) bool {
		//log.Println(r)
		return true
	}
	ev.OnMessage = func(m *indismqgo.MsgBuffer, r *h.Request) bool {
		//log.Println(m.String())
		return true
	}
	go http.ListenHttp(srv, "/", 8081, ev)
	//
	//Create a Client
	client := NewBroker("client")
	client.Debug(&debug)
	//connect to the server
	u, _ := url.Parse("http://localhost:8081")
	conn := http.NewHttpConn(client, u, nil, nil)
	if conn == nil {
		log.Fatal("empty http Connection")
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
	if indismqgo.WaitTimeout(&wg, 2*time.Second) {
		log.Fatal("timeout")
	}

	//Output:
	//Recieved Hello From Client

}
