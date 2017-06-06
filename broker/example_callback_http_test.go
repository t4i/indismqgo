package broker

import (
	"fmt"
	"github.com/t4i/indismqgo"
	"log"
	"net/url"
	"sync"
	"time"
)

func Example_callbackHTTP() {

	// Create A Server
	srv := NewBroker("srv")
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Create A Handler for the /test path
	srv.Handlers.Set("/test", func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
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
	go srv.ListenHttp("/", 8082)
	//
	//Create a Client
	client := NewBroker("client")
	//connect to the server
	u, _ := url.Parse("http://localhost:8082")
	conn := client.NewHttpConn(u, nil)
	if conn == nil {
		log.Fatal("empty http Connection")
	}

	//send the server a test message
	m, err := client.NewMsgObject("srv", indismqgo.ActionGET, "/test", []byte("Hello From Client"), func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
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
	if indismqgo.WaitTimeout(&wg, 10*time.Second) {
		log.Fatal("timeout")
	}

	//Output:
	//Server Recieved Hello From Client
	//Client Recieved Hello Back From Server

}
