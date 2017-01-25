package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	indismq "github.com/t4i/indismqgo"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWaitWs = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSizeWs = 8192

	// Time allowed to read the next pong message from the peer.
	pongWaitWs = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriodWs = (pongWaitWs * 9) / 10
)

//WsMessageRecieved this is connection specific
func (s *Server) WsMessageRecieved(message *[]byte, ws *websocket.Conn) {
	m := indismq.ParseMsg(message)
	from := string(m.Fields.From())
	c := s.Connections[from]
	if c == nil || len(c.User) < 1 {
		c = &Connection{User: from, ConnectionType: s.ConnectionTypes["ws"]}
	}
	r := s.Imq.RecieveMessage(m, c.User)
	if r != nil && r.Data != nil {
		s.SendWsMessage(m, ws)
	}

}
func (s *Server) DefaultWsSender(m *indismq.Msg, c *Connection) error {
	if m == nil || c == nil || len(c.Name) < 1 {
		return errors.New("Invalid Message")
	}
	ws := s.webSockets[c.User]
	if ws == nil {
		return errors.New("No Connection")
	}
	s.SendWsMessage(m, ws)
	return nil
}

//SendMessage ...
func (s *Server) SendWsMessage(m *indismq.Msg, ws *websocket.Conn) {
	if m != nil && ws != nil && m.Data != nil {
		if s.Imq.Debug {
			t := schema.GetRootAsImq(*m.Data, 0)
			fmt.Println("Sender ID ", string(t.MsgId()), " From ", string(t.From()), " To ", string(t.To()))
		}
		s.sendLocks[ws].Lock()
		er := ws.WriteMessage(2, *m.Data)
		s.sendLocks[ws].Unlock()
		if er != nil {
			log.Println(er)
		}
	}
}

func (s *Server) onWsConnected(ws *websocket.Conn) {
	//s.Connections[c.User] = c
	s.sendLocks[ws] = &sync.Mutex{}
	//s.webSockets[c.User] = ws
}

func (s *Server) onWsDisconnected(c *Connection, ws *websocket.Conn) {
	delete(s.webSockets, c.Name)
	delete(s.sendLocks, ws)
	s.Imq.DelSubscriberAll(c.Name)
}

//Connect ...
func (s *Server) ConnectWebsocket(address string, name string, header http.Header, user string) (ok bool, err error) {
	s.Imq.SetName(name)
	s.Imq.SetRelayHandler(s.relayHandler)
	s.Imq.SetBrokerHandler(s.brokerHandler)
	if _, ok := s.ConnectionTypes["ws"]; !ok {
		return false, errors.New("Invalid Connection Type")
	}
	Url, err := url.ParseRequestURI(address)
	if err != nil {
		return false, err
	}
	var ws *websocket.Conn
	ws, _, err = websocket.DefaultDialer.Dial(address, header)
	if err != nil {
		fmt.Println("connection failed ", err)
		return
	}
	//ws.SetReadDeadline(time.Now().Add(pongWait))
	//ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	s.onWsConnected(ws)
	recieveQuit := make(chan bool)
	go s.recieveWs(ws, recieveQuit)

	syn := s.Imq.Syn("", func(m *indismq.Msg) *indismq.Msg {

		c, err := s.NewConnection(string(m.Fields.From()), "", "ws", Url, make(map[string][]string))
		if err != nil {
			delete(s.sendLocks, ws)
			ws.Close()
			return nil
		}
		s.Connections[string(m.Fields.From())] = c
		s.webSockets[string(m.Fields.From())] = ws
		fmt.Println("ack success ", string(m.Fields.From()))
		if s.OnReady != nil {
			s.OnReady(c)
		}
		pingQuit := make(chan bool)
		go s.ping(ws, pingQuit)
		go func(recieveQuit chan bool, pingQuit chan bool) {
			select {
			case <-recieveQuit:
				pingQuit <- true
				break
			case <-pingQuit:
				recieveQuit <- true
				break
			}
			s.onWsDisconnected(c, ws)
			if s.OnServerDisconnected != nil {
				defer s.OnServerDisconnected(c)
			}

		}(recieveQuit, pingQuit)
		return nil
	}, user)
	s.SendWsMessage(syn, ws)
	return true, nil
}

func (s *Server) ping(ws *websocket.Conn, quit chan bool) {
	ticker := time.NewTicker(pingPeriodWs)
	defer ticker.Stop()
	var pingCount = 0
	for {

		select {
		case <-ticker.C:
			if s.Imq.Debug {
				log.Println("ping")
			}
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWaitWs)); err != nil {
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
	if _, ok := s.ConnectionTypes["ws"]; !ok {
		log.Println("Invalid Connection Type")
		return
	}
	if s.ConnectionTypes["ws"].AuthHandler != nil {
		if ok, err := s.ConnectionTypes["ws"].AuthHandler(r.Header); err != nil || !ok {
			w.WriteHeader(http.StatusUnauthorized)
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
	recieveQuit := make(chan bool)
	go s.recieveWs(temp, recieveQuit)

	syn := s.Imq.Syn("", func(m *indismq.Msg) *indismq.Msg {
		c, err := s.NewConnection(string(m.Fields.From()), "", "ws", r.URL, make(map[string][]string))
		if err != nil {
			delete(s.sendLocks, temp)
			temp.Close()
			return nil
		}
		if s.OnSyn != nil {
			if ok, errMsg, err := s.OnSyn(r, c); !ok {
				ticker := time.NewTicker(time.Second * 1)
				go func() {
					temp.Close()
					ticker.Stop()
				}()
				return s.Imq.Err(m, errMsg, err, "")
			}
		}

		s.Connections[string(m.Fields.From())] = c
		s.webSockets[string(m.Fields.From())] = temp
		fmt.Println("ack success ", string(m.Fields.From()))
		if s.OnReady != nil {
			s.OnReady(c)
		}
		pingQuit := make(chan bool)
		go s.ping(temp, pingQuit)
		go func(recieveQuit chan bool, pingQuit chan bool) {
			select {
			case <-recieveQuit:
				pingQuit <- true
				break
			case <-pingQuit:
				recieveQuit <- true
				break
			}

			s.onWsDisconnected(c, temp)
			if s.OnClientDisconnected != nil {
				defer s.OnClientDisconnected(c)
			}
		}(recieveQuit, pingQuit)
		return nil
	}, "")
	fmt.Println(string(syn.Fields.From()))
	s.onWsConnected(temp)
	s.SendWsMessage(syn, temp)

}

func (s *Server) recieveWs(w *websocket.Conn, quit chan bool) {
	defer w.Close()
	for {
		_, message, err := w.ReadMessage()
		if err != nil {
			quit <- true
			return
		}
		s.WsMessageRecieved(&message, w)
		//}

	}
}

//Listen ...
func (s *Server) ListenWebSocket(path string, port int) {
	fmt.Println("server starting")
	log.Println("starting ws")
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(path, s.upgrade)
	go log.Println(http.ListenAndServe(":"+strconv.Itoa(port), serverMux).Error())
}

//ListenTLS ...
func (s *Server) ListenWebsocketTLS(path string, port int, certFile string, keyFile string, config *tls.Config) {
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
			log.Printf("%s-%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Server Error: " + err)
		}
	}()

}
