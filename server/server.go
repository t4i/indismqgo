package server

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	indismq "github.com/t4i/indismqgo"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//Server ...
type Server struct {
	Imq        *indismq.Imq
	webSockets map[string]*websocket.Conn
	clients    map[*websocket.Conn]string
	upgrader   websocket.Upgrader
	sendLocks  map[*websocket.Conn]*sync.Mutex

	//GetCertificate ..
	GetCertificate func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

	//OnUpgrade ...
	OnUpgrade func(w http.ResponseWriter, r *http.Request) (success bool, message string)

	//OnMsgRecieved ...
	OnMsgRecieved func(m *indismq.Msg, w *websocket.Conn) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnBroker ...
	OnBroker func(m *indismq.Msg) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnSyn ...
	OnSyn func(client string, r *http.Request, ws *websocket.Conn) (ok bool, errMsg string, err int8)

	//OnRelay ...
	OnRelay func(m *indismq.Msg) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnUnknownClient ...
	OnUnknownClient func(m *indismq.Msg) (forward bool, to string)
	//OnDisconnected ...
	OnServerDisconnected func(address string, name string, header http.Header)
	//OnConnected ...
	OnConnected func(address string, name string, header http.Header, client string, ws *websocket.Conn)

	//OnReady ...
	OnReady func(client string, ws *websocket.Conn)
	//OnClientDisconnected ...
	OnClientDisconnected func(ws *websocket.Conn)
}

//NewServer ...
func NewServer() *Server {
	var s = new(Server)
	s.Imq = indismq.NewImq()
	s.webSockets = make(map[string]*websocket.Conn)
	s.clients = make(map[*websocket.Conn]string)
	s.upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options
	s.sendLocks = make(map[*websocket.Conn]*sync.Mutex)
	return s
}

//SetHandler ...

func (s *Server) messageRecieved(message *[]byte, w *websocket.Conn) {
	m := indismq.ParseMsg(message)
	if s.OnMsgRecieved != nil {
		ok, subM, errMsg, err := s.OnMsgRecieved(m, w)
		if !ok {
			s.SendMessage(s.Imq.Err(m, errMsg, err).Data, w)
			return
		}
		if subM != nil {
			m = subM
		}
	}
	r := s.Imq.RecieveMessage(m)
	if r != nil && r.Data != nil {
		s.SendMessage(r.Data, w)
	}

}

//SendMessage ...
func (s *Server) SendMessage(data *[]byte, ws *websocket.Conn) {

	if data != nil {
		t := schema.GetRootAsImq(*data, 0)
		if s.Imq.Debug {
			fmt.Println("Sender ID ", string(t.MsgId()), " From ", string(t.From()), " To ", string(t.To()))
		}
		s.sendLocks[ws].Lock()
		er := ws.WriteMessage(2, *data)
		s.sendLocks[ws].Unlock()
		if er != nil {
			log.Println(er)
		}
	}
}

func (s *Server) relayHandler(m *indismq.Msg) *indismq.Msg {
	if s.OnRelay != nil {
		ok, subM, errMsg, err := s.OnRelay(m)
		if !ok {
			return s.Imq.Err(m, errMsg, err)
		}
		if subM != nil {
			m = subM
		}
	}
	if w, ok := s.webSockets[string(m.Fields.To())]; ok {
		s.SendMessage(m.Data, w)
		return nil
	} else if s.OnUnknownClient != nil {
		forward, to := s.OnUnknownClient(m)
		if forward {
			if w, ok := s.webSockets[to]; ok {
				s.SendMessage(m.Data, w)
				return nil
			}
		}
	}

	return s.Imq.Err(m, "Client not found", schema.ErrINVALID)

}

func (s *Server) brokerHandler(m *indismq.Msg) *indismq.Msg {
	if s.OnBroker != nil {
		ok, subM, errMsg, err := s.OnBroker(m)
		if !ok {
			return s.Imq.Err(m, errMsg, err)
		}
		if subM != nil {
			m = subM
		}
	}
	var callback indismq.Handler
	if m.Fields.Callback() != 0 {
		callback = func(imqMessage *indismq.Msg) *indismq.Msg {
			s.SendMessage(imqMessage.Data, s.webSockets[string(m.Fields.From())])
			return nil
		}
	}
	s.Imq.BrokerReplay(m, func(client string, imqMessage *indismq.Msg) {
		if _, ok := s.webSockets[client]; ok {
			s.SendMessage(imqMessage.Data, s.webSockets[client])
		} else if s.OnUnknownClient != nil {
			forward, to := s.OnUnknownClient(m)
			if forward {
				if w, ok := s.webSockets[to]; ok {
					s.SendMessage(m.Data, w)

				}
			}
		}
	}, callback)
	return nil
}

//Listen ...
func (s *Server) Listen(path string, port int, name string) {
	s.Imq.SetName(name)
	s.Imq.SetRelayHandler(s.relayHandler)
	s.Imq.SetBrokerHandler(s.brokerHandler)
	fmt.Println("server starting")
	log.Println("starting ws")
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(path, s.upgrade)
	go log.Println(http.ListenAndServe(":"+strconv.Itoa(port), serverMux).Error())
}

//ListenTLS ...
func (s *Server) ListenTLS(path string, port int, name string, certFile string, keyFile string, config *tls.Config) {
	s.Imq.SetName(name)
	s.Imq.SetRelayHandler(s.relayHandler)
	s.Imq.SetBrokerHandler(s.brokerHandler)
	if s.Imq.Debug {
		fmt.Println("server starting")
		log.Println("starting ws")
	}
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(path, s.upgrade)
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   serverMux,
	}
	s.GetCertificate = config.GetCertificate
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if s.Imq.Debug {
			log.Printf("%s-%s-%s", path, strconv.Itoa(port), name)
			log.Println("TLS Server Error: " + err)
		}
	}()

}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func (s *Server) onConnected(ws *websocket.Conn) {
	s.clients[ws] = ""
	s.sendLocks[ws] = &sync.Mutex{}
}

func (s *Server) onDisconnected(ws *websocket.Conn) {
	client := s.clients[ws]
	delete(s.webSockets, client)
	delete(s.clients, ws)
	delete(s.sendLocks, ws)
	s.Imq.DelSubscriberAll(client)
}

//Connect ...
func (s *Server) Connect(address string, name string, header http.Header) (ok bool, err error) {
	s.Imq.SetName(name)
	s.Imq.SetRelayHandler(s.relayHandler)
	s.Imq.SetBrokerHandler(s.brokerHandler)
	var ws *websocket.Conn
	ws, _, err = websocket.DefaultDialer.Dial(address, header)
	if err != nil {
		fmt.Println("connection failed ", err)
		return
	}
	//ws.SetReadDeadline(time.Now().Add(pongWait))
	//ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	s.onConnected(ws)
	recieveQuit := make(chan bool)
	go s.receive(ws, recieveQuit)
	pingQuit := make(chan bool)
	go s.ping(ws, pingQuit)
	go func(reieveQuit chan bool, pingQuit chan bool) {
		select {
		case <-recieveQuit:
			pingQuit <- true
			break
		case <-pingQuit:
			recieveQuit <- true
			break
		}
		s.onDisconnected(ws)
		if s.OnServerDisconnected != nil {
			defer s.OnServerDisconnected(address, name, header)
		}

	}(recieveQuit, pingQuit)
	syn := s.Imq.Syn("", func(m *indismq.Msg) *indismq.Msg {

		s.webSockets[string(m.Fields.From())] = ws
		s.clients[ws] = string(m.Fields.From())
		fmt.Println("ack success ", string(m.Fields.From()))
		if s.OnReady != nil {
			s.OnReady(string(m.Fields.From()), ws)
		}
		return nil
	})
	s.SendMessage(syn.Data, ws)
	return true, nil
}

func (s *Server) ping(ws *websocket.Conn, quit chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	var pingCount = 0
	for {

		select {
		case <-ticker.C:
			if s.Imq.Debug {
				log.Println("ping")
			}
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				if s.Imq.Debug {
					log.Println("ping:", err)
					log.Println("ping count:", pingCount)
				}
				if pingCount == 2 {
					quit <- true
				}
				pingCount++

			} else {
				// sendQueue()
				if s.Imq.Debug {
					log.Println("ping success")
				}
			}
		case <-quit:
			if s.Imq.Debug {
				log.Println("ping done")
			}
			return
		}
	}
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	log.Println("upgrade request")
	if s.OnUpgrade != nil {
		if ok, message := s.OnUpgrade(w, r); !ok {
			w.Write([]byte(message))
			return
		}
	}
	var err error
	var temp *websocket.Conn
	temp, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	syn := s.Imq.Syn("", func(m *indismq.Msg) *indismq.Msg {

		if s.OnSyn != nil {
			if ok, errMsg, err := s.OnSyn(string(m.Fields.From()), r, temp); !ok {
				ticker := time.NewTicker(time.Second * 1)
				go func() {
					temp.Close()
					ticker.Stop()
				}()
				return s.Imq.Err(m, errMsg, err)
			}
		}
		s.webSockets[string(m.Fields.From())] = temp
		s.clients[temp] = string(m.Fields.From())
		fmt.Println("ack success ", string(m.Fields.From()))
		if s.OnReady != nil {
			s.OnReady(string(m.Fields.From()), temp)
		}
		return nil
	})
	fmt.Println(string(syn.Fields.From()))
	s.onConnected(temp)
	s.SendMessage(syn.Data, temp)
	recieveQuit := make(chan bool)
	go s.receive(temp, recieveQuit)
	pingQuit := make(chan bool)
	go s.ping(temp, pingQuit)
	go func(reieveQuit chan bool, pingQuit chan bool) {
		select {
		case <-recieveQuit:
			pingQuit <- true
			break
		case <-pingQuit:
			recieveQuit <- true
			break
		}

		s.onDisconnected(temp)
		if s.OnClientDisconnected != nil {
			defer s.OnClientDisconnected(temp)
		}
	}(recieveQuit, pingQuit)

}

func (s *Server) receive(w *websocket.Conn, quit chan bool) {
	defer w.Close()
	for {
		_, message, err := w.ReadMessage()
		if err != nil {
			quit <- true
			return
		}
		s.messageRecieved(&message, w)
		//}

	}
}
