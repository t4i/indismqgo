package broker

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	indismq "github.com/t4i/indismqgo"
	"log"
	"net/http"
	"strconv"
	//"sync"
	"io/ioutil"
	"net/url"
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

type WsConn struct {
	Context *indismq.Context
	*websocket.Conn
	sync.Mutex
	queue []*indismq.MsgBuffer
}

func (b *Broker) NewWsConnection(ws *websocket.Conn) *WsConn {
	conn := &WsConn{Context: b.Context}
	conn.Conn = ws
	conn.queue = []*indismq.MsgBuffer{}
	return conn
}

func (ws *WsConn) SendQueue() {
	if len(ws.queue) > 0 {
		temp := ws.queue[:0]
		ws.Lock()
		for _, v := range ws.queue {

			err := ws.WriteMessage(2, v.Data)
			if err != nil {
				temp = append(temp, v)
			}
		}
		ws.Unlock()
		ws.queue = temp
	}

}

//SendMessage ...
func (conn *WsConn) Send(m *indismq.MsgBuffer) error {
	if conn.Context.Debug {
		log.Println(string(conn.Context.Name), "Sending", m.String())
	}

	// ws, ok := conn.Conn.(*wsConn)
	// if !ok {
	// 	return false, errors.New("Not a valid Websocket")
	// }
	if m != nil && m.Data != nil {
		if conn.Context.Debug {
			t := indismq.GetRootAsImq(m.Data, 0)
			fmt.Println("Sender ID ", string(t.MsgId()), " From ", string(t.From()), " To ", string(t.To()))
		}
		conn.Lock()
		er := conn.WriteMessage(2, m.Data)
		conn.Unlock()
		if er != nil {
			log.Println(er)
			conn.queue = append(conn.queue, m)
		}
		conn.Context.Messages.Store(m)
	} else if m != nil && m.Data != nil {
		conn.queue = append(conn.queue, m)
	}
	return nil
}

// func (s *Broker) onWsConnected(ws *wsConn) {
// 	//s.Connections[c.User] = c
// 	//s.sendLocks[ws] = &sync.Mutex{}
// 	//s.webSockets[c.User] = ws
// }

// func (s *Broker) onWsDisconnected(ws *wsConn) {
// 	// delete(s.webSockets, c.Name)
// 	// delete(s.sendLocks, ws)
// 	// s.Context.DelSubscriberAll(c.Name)
// }
var WsReconnectDelay time.Duration = time.Second * 5

func (s *Broker) AutoWsReconnect(URL *url.URL, header http.Header, ws *WsConn) {
	connected := false
	for !connected {
		log.Println("Retrying", URL.String())
		time.Sleep(WsReconnectDelay)
		_, err := s.ConnectWebsocket(URL, header, ws, s.AutoWsReconnect)
		if err != nil && s.Context.Debug {
			log.Println(err)
		} else {
			connected = true
		}
	}

}

//Connect ...
func (s *Broker) ConnectWebsocket(URL *url.URL, header http.Header, ws *WsConn, onDisconnected func(*url.URL, http.Header, *WsConn)) (*WsConn, error) {
	if ws == nil {
		ws = s.NewWsConnection(nil)
	}
	var err error
	if s.Context.Debug {
		log.Println("connecting to", URL.String())
	}
	ws.Conn, _, err = websocket.DefaultDialer.Dial(URL.String(), header)
	if err != nil {
		if onDisconnected != nil {
			go onDisconnected(URL, header, ws)
		} else {
			log.Println(err, URL)
		}
		return nil, err
	}
	ws.SetReadDeadline(time.Now().Add(pongWaitWs))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWaitWs)); return nil })

	//s.onWsConnected(ws)
	recieveQuit := make(chan bool)
	go s.recieveWs(ws, recieveQuit)
	connectMsg, err := s.Context.Connection(nil, nil, func(m *indismq.MsgBuffer, c indismq.Connection) error {
		if m.Fields.Status() >= 200 && m.Fields.Status() < 300 {
			from := string(m.Fields.From())
			if s.Context.Debug {
				log.Println("connected", from)
			}
			s.Context.Connections.Set(from, ws)
			pingQuit := make(chan bool)
			go s.ping(ws, pingQuit)
			go func(recieveQuit chan bool, pingQuit chan bool) {
				select {
				case <-recieveQuit:
					if s.Context.Debug {
						log.Println("recieveQuit")
					}
					pingQuit <- true
					break
				case <-pingQuit:
					if s.Context.Debug {
						log.Println("pingQuit")
					}
					recieveQuit <- true
					break
				}
				s.Context.Connections.Del(from)
				ws.Close()
				if s.Context.Debug {
					log.Println("Disconnected", from)
				}
				if onDisconnected != nil {
					defer onDisconnected(URL, header, ws)
				}
				// if s.OnBrokerDisconnected != nil {
				// 	defer s.OnBrokerDisconnected(c)
				// }

			}(recieveQuit, pingQuit)
		} else {
			log.Println("error connecting", indismq.StatusText(int(m.Fields.Status())))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	ws.Send(connectMsg)
	ws.SendQueue()
	return ws, nil
}

func (s *Broker) ping(ws *WsConn, quit chan bool) {
	ticker := time.NewTicker(pingPeriodWs)
	defer ticker.Stop()
	var pingCount = 0
	for {

		select {
		case <-ticker.C:
			if s.Context.Debug {
				log.Println("ping")
			}
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWaitWs)); err != nil {
				if s.Context.Debug {
					log.Println("ping:", err)
					log.Println("ping count:", pingCount)
				}
				if pingCount == 2 {
					quit <- true
				}
				pingCount++

			} else {
				// sendQueue()
				if s.Context.Debug {
					log.Println("ping success")
				}
			}
		case <-quit:
			if s.Context.Debug {
				log.Println("ping done")
			}
			return
		}
	}
}

func (s *Broker) upgrade(w http.ResponseWriter, r *http.Request) {
	if s.Context.Debug {
		log.Println("upgrade req")
	}
	auth := true
	var err error
	if s.WsAuthHandler != nil {
		auth, err = s.WsAuthHandler(r.Header.Get("Authorization"))
		if err != nil {
			auth = false
		}
	}
	defer r.Body.Close()
	var m *indismq.MsgBuffer
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	if body != nil && len(body) > 0 {
		m = indismq.ParseMsg(body, s.Context)
	}
	if !auth && m != nil {
		if a := m.Fields.Authorization(); len(a) > 0 && s.WsAuthHandler != nil {
			ok, _ := s.WsAuthHandler(string(a))
			if ok {
				auth = true
			}
		}
		if !auth {
			rep, err := s.MakeReply(m, http.StatusUnauthorized, []byte("Unauthorized"))
			if err == nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(rep.Data)
			return
		}

	} else if !auth {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	ws := s.NewWsConnection(nil)
	ws.Conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
	}
	recieveQuit := make(chan bool)
	go s.recieveWs(ws, recieveQuit)
	connectMsg, err := s.Context.Connection(nil, nil, func(m *indismq.MsgBuffer, c indismq.Connection) error {
		if s.Debug {
			log.Print(m.String())
		}
		if m.Fields.Status() >= 200 && m.Fields.Status() < 300 {
			from := string(m.Fields.From())
			if s.Debug {
				log.Println("connected", from)
			}
			s.Context.Connections.Set(string(m.Fields.From()), ws)
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
				s.Context.Connections.Del(from)
				ws.Close()
				log.Println("Disconnected", from)
				// if s.OnClientDisconnected != nil {
				// 	defer s.OnClientDisconnected(c)
				// }
			}(recieveQuit, pingQuit)
		} else {
			log.Println("error connecting", indismq.StatusText(int(m.Fields.Status())))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = ws.Send(connectMsg)
	if err != nil {
		log.Println(err)
	}
}

func (s *Broker) recieveWs(conn *WsConn, quit chan bool) {

	// wsConn, ok := conn.Conn.(*wsConn)
	// if !ok {
	// 	log.Println("invalid Websocket Connection")
	// 	return
	// }
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			quit <- true
			return
		}
		go func() {
			m := indismq.ParseMsg(message, s.Context)
			err := s.Context.Recieve(m, conn)
			if err != nil {
				log.Println(err)
			}
		}()

		//}

	}
}

//Listen ...
func (s *Broker) ListenWebSocket(path string, port int) {
	if s.Debug {
		log.Println("starting ws")
	}
	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, s.upgrade)
	log.Println(http.ListenAndServe(":"+strconv.Itoa(port), BrokerMux).Error())
}

//ListenTLS ...
func (s *Broker) ListenWebsocketTLS(path string, port int, certFile string, keyFile string, config *tls.Config) {
	if s.Context.Debug {
		fmt.Println("Broker starting")
		log.Println("starting ws")
	}
	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, s.upgrade)
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   BrokerMux,
	}
	s.GetCertificate = config.GetCertificate
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if s.Context.Debug {
			log.Printf("%s-%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Broker Error: " + err)
		}
	}()

}
