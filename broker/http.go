package broker

import (
	"bytes"
	"crypto/tls"
	"fmt"
	indismq "github.com/t4i/indismqgo"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWaitHTTP = 15 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSizeHTTP = 8192

	// Time allowed to read the next pong message from the peer.
	pongWaitHTTP = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriodHTTP = (pongWaitHTTP * 9) / 10
)

type HttpConn struct {
	URL     *url.URL
	Headers http.Header
	Context *indismq.Context
	Broker  *Broker
	*indismq.ConnEvents
}

// type ConnEvents struct {
// 	OnConnect    func(key string, conn Connection)
// 	OnDisconnect func(key string, conn Connection)
// }
var DefaultHttpConnEvents *indismq.ConnEvents

func (s *Broker) NewHttpConn(URL *url.URL, headers http.Header) *HttpConn {
	conn := &HttpConn{URL: URL, Headers: headers, Context: s.Context, Broker: s, ConnEvents: DefaultHttpConnEvents}
	return conn
}
func (conn *HttpConn) VerifyConnectionRequest(m *indismq.MsgBuffer) error {
	return nil
}
func (conn *HttpConn) VerifySubscribeRequest(m *indismq.MsgBuffer) error {
	return nil
}

var HttpRetryDelay time.Duration = time.Second * 5

//SendMessage ...
func (conn *HttpConn) Send(m *indismq.MsgBuffer) error {
	// httpConn, ok := conn.Conn.(*HttpConn)
	// if !ok {
	// 	return false, errors.New("Invalid http conn")
	// }
	if m != nil && m.Data != nil {
		if conn.Context.Debug {
			t := indismq.GetRootAsImq(m.Data, 0)
			fmt.Println("Sender ID ", string(t.MsgId()), " From ", string(t.From()), " To ", string(t.To()))
		}
		client := http.Client{Timeout: writeWaitHTTP}
		buf := bytes.NewBuffer(m.Data)
		req, err := http.NewRequest("POST", conn.URL.String(), buf)
		if err != nil {
			log.Println(err)
		}
		req.Header = conn.Headers
		go func() {
			success := false
			var res *http.Response
			for !success {
				res, err = client.Do(req)
				if err != nil {
					log.Println("http retry")
					time.Sleep(HttpRetryDelay)
				} else {
					success = true
				}

			}

			conn.Context.Messages.Store(m)
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if body == nil || len(body) < 5 {
				return
			}
			if res.StatusCode != http.StatusOK {
				log.Println(string(body))
				return
			}
			rm := indismq.ParseMsg(body, conn.Context)
			if rm == nil {
				return
			}
			err = conn.Context.RecieveRaw(body, conn)
			if err != nil {
				log.Println(err)
				return
			}
		}()

	}
	return nil
}

func (conn *HttpConn) Events() *indismq.ConnEvents {
	return conn.ConnEvents

}

// func (conn *HttpConn) OnConnect(key string) {
// 	if conn.Broker.OnHttpConnect != nil {
// 		conn.Broker.OnHttpConnect(key, conn)
// 	}
// }

// func (conn *HttpConn) OnDisconnect(key string) {
// 	if conn.Broker.OnHttpDisconnect != nil {
// 		conn.Broker.OnHttpDisconnect(key, conn)
// 	}
// }

//var httpConnType = &indismq.ConnType{}

//Connect ...
func (s *Broker) ConnectHttp(conn *HttpConn) (ok bool, err error) {
	//conn := s.Context.NewConnType(s.SendHttpMessage, indismq.ConnClassHalfDuplex)

	connectMsg, err := s.Context.NewConnectionMsg(nil, nil, func(m *indismq.MsgBuffer, c indismq.Connection) error {

		if m.Fields.Status() >= 200 && m.Fields.Status() <= 300 {
			s.Context.Connections.Set(string(m.Fields.From()), conn)
		}
		log.Println("connected", string(m.Fields.From()))
		// pingQuit := make(chan bool)
		// go s.pingHttp(c, pingQuit)
		// go func(pingQuit chan bool) {
		// 	select {d
		// 	case <-pingQuit:
		// 		break
		// 	}
		// 	if s.OnBrokerDisconnected != nil {
		// 		defer s.OnBrokerDisconnected(c)
		// 	}

		// }(pingQuit)
		return nil
	})
	conn.Send(connectMsg)
	return true, nil
}

// func (s *Broker) pingHttp(c *Connection, quit chan bool) {
// 	// ticker := time.NewTicker(pingPeriod)
// 	// defer ticker.Stop()
// 	// var pingCount = 0
// 	// for {

// 	// 	select {
// 	// 	case <-ticker.C:
// 	// 		if s.Context.Debug {
// 	// 			log.Println("ping")
// 	// 		}
// 	// 		if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
// 	// 			if s.Context.Debug {d
// 	// 				log.Println("ping:", err)
// 	// 				log.Println("ping count:", pingCount)
// 	// 			}
// 	// 			if pingCount == 2 {
// 	// 				quit <- true
// 	// 			}
// 	// 			pingCount++

// 	// 		} else {
// 	// 			// sendQueue()
// 	// 			if s.Context.Debug {
// 	// 				log.Println("ping success")
// 	// 			}
// 	// 		}
// 	// 	case <-quit:
// 	// 		if s.Context.Debug {
// 	// 			log.Println("ping done")
// 	// 		}
// 	// 		return
// 	// 	}
// 	// }
// }
// func (s *Broker) onConnect(m *indismq.MsgBuffer, conn *indismq.Connection, minConnType uint8) (reply *indismq.MsgBuffer, connStatus uint8, err error) {
// 	return nil, 3, nil
// }

type AsyncWait struct {
	indismq.Connection
	done chan *indismq.MsgBuffer
}

func (s *AsyncWait) Send(m *indismq.MsgBuffer) error {
	// if s.Context.Debug {
	// 	log.Println("waitfor response called")
	// 	// conType := reflect.TypeOf(conn.Conn)
	// 	// log.Println(conType)
	// }
	// done, ok := conn.Conn.(chan *indismq.MsgBuffer)
	// if !ok {
	// 	return false, errors.New("Not a valid Http Connection")
	// }
	s.done <- m
	return nil
}

func (s *Broker) receiveHttp(w http.ResponseWriter, r *http.Request) {
	if s.Context.Debug {
		log.Println("receive http")
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(nil)
		return
	}
	if body == nil || len(body) < 1 {
		log.Println("No Body Found or Empty")
		return
	}
	m := indismq.ParseMsg(body, s.Context)
	auth := true
	if s.HTTPAuthHandler != nil {
		auth, err = s.HTTPAuthHandler(r.Header.Get("Authorization"))
		if err != nil {
			auth = false
		}
	}
	var resp *indismq.MsgBuffer
	if auth {
		if s.Context.Debug {
			log.Println("Recieve", r.Host, r.RequestURI, r.URL.Host)
		}
		done := make(chan *indismq.MsgBuffer, 1)
		//s.NewConnType(s.WaitForHttpResponse, indismq.ConnClassHalfDuplex)

		err = s.Context.Recieve(m, &AsyncWait{done: done})
		if err != nil {
			log.Println(err)
			return
		}
		if resp == nil {
			if s.Context.Debug {
				log.Println(string(s.Context.Name), "waiting for response")
			}
			resp = <-done
			if s.Context.Debug {
				log.Println("response received")
			}
		}
	} else {
		resp, err = s.MakeReply(m, http.StatusUnauthorized, []byte("Unauthorized"))
	}

	if resp != nil && resp.Data != nil && w != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(resp.Data)
		return
	}
	log.Println("empty response")

}

//Listen ...
func (s *Broker) ListenHttp(path string, port int) {
	if s.Debug {
		fmt.Println("Broker starting")
		log.Println("starting http")
	}

	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, s.receiveHttp)
	log.Println(http.ListenAndServe(":"+strconv.Itoa(port), BrokerMux).Error())
}

//ListenTLS ...
func (s *Broker) ListenHttpTLS(path string, port int, certFile string, keyFile string, config *tls.Config) {
	if s.Context.Debug {
		fmt.Println("Broker starting")
		log.Println("starting http")
	}
	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, s.receiveHttp)
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   BrokerMux,
	}
	s.GetCertificate = config.GetCertificate
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if s.Context.Debug {
			log.Printf("%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Broker Error: " + err)
		}
	}()

}
