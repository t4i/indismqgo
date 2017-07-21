package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/broker/http"
	"log"
	"net/url"
	"sync"
	"time"
)

func Example_callbackHTTP() {
	debug := false
	if debug {
		log.SetFlags(log.Llongfile)
	}
	// Create A Server
	srv := NewBroker("srv")
	srv.Debug(&debug)
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Create A Handler for the /test path
	srv.Handlers.SetHandler("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		if string(m.Fields.From()) != "client" {
			log.Fatal("Client Message Error")
		}
		fmt.Println("Server Recieved", string(m.Fields.BodyBytes()))
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

		return nil
	})
	go http.ListenHttp(srv, "/", 8082, nil)
	//
	//Create a Client
	client := NewBroker("client")
	client.Debug(&debug)
	//connect to the server
	u, _ := url.Parse("http://localhost:8082")
	conn := http.NewHttpConn(client, u, nil, nil)
	if conn == nil {
		log.Fatal("empty http Connection")
	}

	//send the server a test message
	m, err := client.NewMsgObject("srv", indismqgo.ActionGET, "/test", []byte("Hello From Client"), func(m *indismqgo.MsgBuffer, c indismqgo.Sender) error {
		defer wg.Done()
		if string(m.Fields.From()) != "srv" {
			log.Fatal("Server Message Error")
		}
		fmt.Println("Client Recieved", string(m.Fields.BodyBytes()))
		return nil
	}).ToBuffer()

	if err != nil {
		log.Fatal(err)
	}
	err = conn.Send(m)
	if err != nil {
		log.Fatal(err)
	}
	if indismqgo.WaitTimeout(&wg, 5*time.Second) {
		log.Fatal("timeout")
	}

	//Output:
	//Server Recieved Hello From Client
	//Client Recieved Hello Back From Server

}
