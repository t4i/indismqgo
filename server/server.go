package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	imq "github.com/t4i/indismqgo"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
	"log"
	"net/http"
	"sync"
)

var webSockets = make(map[string]*websocket.Conn)
var ws []*websocket.Conn
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options
var sendLock sync.Mutex

func messageRecieved(message *[]byte, w *websocket.Conn) {
	m := imq.RecieveMessage(message)
	if m != nil && m.Data != nil {
		sendMessage(m.Data, w)
	}

}
func sendMessage(data *[]byte, ws *websocket.Conn) {
	if data != nil {
		er := ws.WriteMessage(2, *data)
		if er != nil {
			log.Println(er)
		}
	}
}
func relayHandler(m *imq.Msg) *imq.Msg {
	if w, ok := webSockets[string(m.Fields.To())]; ok {
		sendMessage(m.Data, w)
		return nil
	}
	return imq.Err(m, "Client not found", schema.ErrINVALID)

}
func brokerHandler(m *imq.Msg) *imq.Msg {

	imq.BrokerReplay(m, func(client string, imqMessage *imq.Msg) {
		if _, ok := webSockets[client]; ok {
			sendMessage(imqMessage.Data, webSockets[client])
		}

	}, func(imqMessage *imq.Msg) *imq.Msg {
		sendMessage(imqMessage.Data, webSockets[string(m.Fields.From())])
		return nil
	})
	return nil
}

//Listen ...
func Listen(path string, port int, name string) {
	imq.SetName(name)
	imq.SetRelayHandler(relayHandler)
	imq.SetBrokerHandler(brokerHandler)
	fmt.Println("server starting")
	log.Println("starting ws")
	http.HandleFunc(path, upgrade)
	log.Println(http.ListenAndServe(":"+string(port), http.HandlerFunc(upgrade)).Error())
}

func callback(m *imq.Msg) *imq.Msg {
	if m.Fields.Sts() == schema.StsSUCCESS {
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
	sendMessage(syn.Data, temp)
	receive(temp)

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
