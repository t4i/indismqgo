package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	//"github.com/pkg/profile"
	"log"
	"net/http"
	"sync"
	imq "t4i/IndisMQ/go"
	schema "t4i/IndisMQ/schema/IndisMQ"
	"time"
)

var webSockets = make(map[string]*websocket.Conn)
var ws []*websocket.Conn
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options
var sendLock sync.Mutex

func messageRecieved(message *[]byte, w *websocket.Conn) {
	m := imq.RecieveMessage(message)
	if m != nil && m.RawData != nil {
		sendMessage(m.RawData)
	}

}
func sendMessage(data *[]byte) {
	if data != nil {
		er := ws.WriteMessage(2, *data)
		if er != nil {
			log.Println(er)
		}
	}

}
func relayHandler(m *imq.Msg) *imq.Msg {
	if w, ok := webSockets[string(m.Fields.To())]; ok {
		sendMessage(m.RawData, w)
		return nil
	} else {
		return imq.Err(m, "Client not found", schema.ErrINVALID)
	}

}
func brokerHandler(m *imq.Msg) *imq.Msg {

	imq.BrokerReplay(m, func(client string, imqMessage *imq.Msg) {
		if _, ok := webSockets[client]; ok {
			sendMessage(imqMessage.RawData, webSockets[client])
		}

	}, func(imqMessage *imq.Msg) *imq.Msg {
		sendMessage(imqMessage.RawData, webSockets[string(m.Fields.From())])
		return nil
	})
	return nil
}
func callHandler(m *imq.Msg) *imq.Msg {
	fmt.Println("/test called")
	return imq.Rep(m, "Hello", nil)
}
func main() {
	imq.SetName("Server")
	imq.SetRelayHandler(relayHandler)
	imq.SetBrokerHandler(brokerHandler)
	imq.SetHandler("/temp", callHandler)
	fmt.Println("server starting")
	log.Println("starting ws")
	http.HandleFunc("/test", upgrade)
	log.Println(http.ListenAndServe(":7000", http.HandlerFunc(upgrade)).Error())

	var m=imq.Req("Server1", "/foo", []byte("custom message"), callback)
	sendMessage(m.Data)
}

func callback(m *imq.Msg) *imq.Msg {
	if m.Fields.Sts == schema.StsSUCCESS {
		fmt.Println("woohoo made it")
	}
	return nil
}

func upgrade(w http.ResponseWriter, r *http.Request) {
	//log.Println("upgrade request")
	var err error
	var temp *websocket.Conn
	temp, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	ws = append(ws, temp)
	syn := imq.Syn("", func(val *imq.Msg) *imq.Msg {
		webSockets[string(val.Fields.From())] = temp
		fmt.Println("ack success ", string(val.Fields.From()))
		return nil
	})
	sendMessage(syn.RawData, temp)
	receive(temp)

}

var sending bool
var count int

func send() {

	temp := []byte("Test1")

	c := 5
	for i := 0; i < c; i++ {
		time.Sleep(time.Millisecond * 2000)
		sendMessage(imq.Req("", "/what", nil, nil).RawData, webSockets["browser"])
		imq.Mult(false, "/hello", temp, func(client string, val *imq.Msg) {
			//fmt.Println("/hommy ", val)
			sendMessage(val.RawData, webSockets[client])
		}, func(val *imq.Msg) *imq.Msg {
			fmt.Println("got response")
			return nil
		})

	}

}
func receive(w *websocket.Conn) {
	defer w.Close()
	for {
		_, message, err := w.ReadMessage()
		if err != nil {
			break
		}
		messageRecieved(&message, w)

	}
}

