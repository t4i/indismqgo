package websocket

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/t4i/indismqgo"
	"log"
	"net/http"
	"strconv"
	//"sync"
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

var _ indismqgo.Connection = (*WsConn)(nil)

type Events struct {
	OnBeforeConnected func(ws *WsConn, req *http.Request) bool
	OnConnected       func(m *indismqgo.MsgBuffer, ws *WsConn) bool
	OnDisconnected    func(ws *WsConn)
	OnMessage         func(m *indismqgo.MsgBuffer, ws *WsConn) bool
}

type WsConn struct {
	Name    string
	Context indismqgo.Context
	queue   []*indismqgo.MsgBuffer
	Claims  map[string]string
	*websocket.Conn
	sync.Mutex
	Events
}

func NewWsConnection(s indismqgo.Context, name string, ws *websocket.Conn, ev *Events) *WsConn {
	conn := &WsConn{Context: s}
	conn.Conn = ws
	if ev != nil {
		conn.Events = *ev
	}
	conn.queue = []*indismqgo.MsgBuffer{}
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
func (conn *WsConn) Send(m *indismqgo.MsgBuffer) error {
	if conn.Context.Debug(nil) {
		log.Println(string(conn.Context.Name(nil)), "Sending", m.String())
	}

	if m != nil && m.Data != nil {
		if conn.Context.Debug(nil) {
			fmt.Println(m.String())
		}
		conn.Lock()
		er := conn.WriteMessage(2, m.Data)
		conn.Unlock()
		if er != nil {
			log.Println(er)
			conn.queue = append(conn.queue, m)
		}
		if messages, ok := conn.Context.(indismqgo.MessageStore); ok {
			messages.SetMessage(string(m.Fields.MsgId()), m)
		}

	} else if m != nil && m.Data != nil {
		conn.queue = append(conn.queue, m)
	}
	return nil
}

var WsReconnectDelay time.Duration = time.Second * 5

func (ws *WsConn) dialWebsocket(u string, header http.Header) {
	connected := false
	var err error
	for !connected {
		ws.Conn, _, err = websocket.DefaultDialer.Dial(u, header)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 1)
		} else {
			connected = true
		}
	}
	return
}

func (ws *WsConn) finishConnection() {
	recieveQuit := make(chan bool)
	go ws.recieveWs(recieveQuit)
	connectMsg, err := indismqgo.NewCtxConnectionMsg(ws.Context, nil, nil, func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		if ws.Context.Debug(nil) {
			log.Print(m.String())
		}
		if m.Fields.Status() >= 200 && m.Fields.Status() < 300 {
			from := string(m.Fields.From())
			if ws.Context.Debug(nil) {
				log.Println("connected", from)
			}
			ws.Name = string(m.Fields.From())
			if ws.OnConnected != nil {
				ws.OnConnected(m, ws)
			}

		} else {
			log.Println("error connecting", indismqgo.StatusText(int(m.Fields.Status())))
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

//Connect ...
func ConnectWebsocket(s indismqgo.Context, URL *url.URL, header http.Header, ev *Events, reconnect bool) (*WsConn, error) {
	ws := NewWsConnection(s, "", nil, ev)
	if reconnect {
		oldOnDis := ws.OnDisconnected
		ws.OnDisconnected = func(ws *WsConn) {
			defer ConnectWebsocket(s, URL, header, ev, reconnect)
			if oldOnDis != nil {
				oldOnDis(ws)
			}

		}
	}
	if s.Debug(nil) {
		log.Println("connecting to", URL.String())
	}
	ws.dialWebsocket(URL.String(), header)
	ws.finishConnection()
	return ws, nil
}

var upgrader = websocket.Upgrader{}

func upgrade(s indismqgo.Context, w http.ResponseWriter, r *http.Request, ev *Events) {
	if s.Debug(nil) {
		log.Println("upgrade req")
	}
	ws := NewWsConnection(s, "", nil, ev)
	if ws.OnBeforeConnected != nil {
		if !ws.OnBeforeConnected(ws, r) {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
		}
	}
	var err error
	ws.Conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	ws.finishConnection()

}

func (conn *WsConn) recieveWs(quit chan bool) {
	ticker := time.NewTicker(pingPeriodWs)
	defer func() {
		conn.Close()
		ticker.Stop()
		if conn.OnDisconnected != nil {
			conn.OnDisconnected(conn)
		}

	}()

	readerr := make(chan error)
	var pingCount = 0
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				readerr <- err
				return
			}

			go func() {
				m := indismqgo.ParseMsg(message, conn.Context)
				if conn.OnMessage != nil {
					if !conn.OnMessage(m, conn) {
						return
					}
				}
				err := conn.Context.Recieve(m, conn)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}()
	select {
	case err := <-readerr:
		if conn.Context.Debug(nil) {
			log.Println(err)
		}
		break
	case <-quit:
		break
	case <-ticker.C:
		if conn.Context.Debug(nil) {
			log.Println("ping")
		}
		conn.Lock()
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWaitWs)); err != nil {
			conn.Unlock()
			if conn.Context.Debug(nil) {
				log.Println("ping:", err)
				log.Println("ping count:", pingCount)
			}
			if pingCount == 2 {
				if conn.Context.Debug(nil) {
					log.Println("Ping Error, closing")
				}
				break
			}
			pingCount++

		} else {
			conn.Unlock()
			if conn.Context.Debug(nil) {
				log.Println("ping success")
			}
		}
	}
}

//Listen ...
func ListenWebSocket(s indismqgo.Context, path string, port int, ev *Events) {
	if s.Debug(nil) {
		log.Println("starting ws", string(s.Name(nil)))
	}
	ContextMux := http.NewServeMux()
	ContextMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		upgrade(s, w, r, ev)
	})
	log.Println(http.ListenAndServe(":"+strconv.Itoa(port), ContextMux).Error())
}

//ListenTLS ...
func ListenWebsocketTLS(s indismqgo.Context, path string, port int, certFile string, keyFile string, config *tls.Config, ev *Events) {
	if s.Debug(nil) {
		log.Println("starting wss", string(s.Name(nil)))
	}
	ContextMux := http.NewServeMux()
	ContextMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		upgrade(s, w, r, ev)
	})
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   ContextMux,
	}
	//s.GetCertificate = config.GetCertificate
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if s.Debug(nil) {
			log.Printf("%s-%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Context Error: " + err)
		}
	}()

}
