package server

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	indismq "github.com/t4i/indismqgo"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWaitHTTP = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSizeHTTP = 8192

	// Time allowed to read the next pong message from the peer.
	pongWaitHTTP = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriodHTTP = (pongWaitHTTP * 9) / 10
)

func (s *Server) DefaultHttpSender(m *indismq.Msg, c *Connection) error {
	if m == nil || c == nil || len(c.Name) < 1 {
		return errors.New("Invalid Message")
	}

	s.SendHttpMessage(m, c.URL, c.Headers)
	return nil
}

//SendMessage ...
func (s *Server) SendHttpMessage(m *indismq.Msg, Url *url.URL, headers http.Header) {

	if m != nil && m.Data != nil {
		if s.Imq.Debug {
			t := schema.GetRootAsImq(*m.Data, 0)
			fmt.Println("Sender ID ", string(t.MsgId()), " From ", string(t.From()), " To ", string(t.To()))
		}
		client := http.Client{}
		buf := bytes.NewBuffer(*m.Data)
		req, err := http.NewRequest("POST", Url.String(), buf)
		if err != nil {
			log.Println(err)
		}
		req.Header = headers
		res, err := client.Do(req)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		rm := indismq.ParseMsg(&body)
		log.Println(body)
		if rm == nil {
			return
		}
		from := string(rm.Fields.From())
		resp := s.Imq.RecieveMessage(m, from)
		if resp != nil {
			defer s.SendHttpMessage(resp, Url, headers)
		}
	}
}

//Connect ...
func (s *Server) ConnectHttp(address string, header http.Header, user string) (ok bool, err error) {
	if _, ok := s.ConnectionTypes["http"]; !ok {
		return false, errors.New("Invalid Connection Type")
	}
	Url, err := url.ParseRequestURI(address)
	if err != nil {
		return false, err
	}

	syn := s.Imq.Syn("", func(m *indismq.Msg) *indismq.Msg {

		c, err := s.NewConnection(string(m.Fields.From()), "", "http", Url, make(map[string][]string))
		if err != nil {
			log.Println(err)
			return nil
		}
		s.Connections[string(m.Fields.From())] = c
		fmt.Println("ack success ", string(m.Fields.From()))
		if s.OnReady != nil {
			s.OnReady(c)
		}
		pingQuit := make(chan bool)
		go s.pingHttp(c, pingQuit)
		go func(pingQuit chan bool) {
			select {
			case <-pingQuit:
				break
			}
			if s.OnServerDisconnected != nil {
				defer s.OnServerDisconnected(c)
			}

		}(pingQuit)
		return nil
	}, user)
	s.SendHttpMessage(syn, Url, header)
	return true, nil
}

func (s *Server) pingHttp(c *Connection, quit chan bool) {
	// ticker := time.NewTicker(pingPeriod)
	// defer ticker.Stop()
	// var pingCount = 0
	// for {

	// 	select {
	// 	case <-ticker.C:
	// 		if s.Imq.Debug {
	// 			log.Println("ping")
	// 		}
	// 		if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
	// 			if s.Imq.Debug {
	// 				log.Println("ping:", err)
	// 				log.Println("ping count:", pingCount)
	// 			}
	// 			if pingCount == 2 {
	// 				quit <- true
	// 			}
	// 			pingCount++

	// 		} else {
	// 			// sendQueue()
	// 			if s.Imq.Debug {
	// 				log.Println("ping success")
	// 			}
	// 		}
	// 	case <-quit:
	// 		if s.Imq.Debug {
	// 			log.Println("ping done")
	// 		}
	// 		return
	// 	}
	// }
}

func (s *Server) receiveHttp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	if body == nil {
		return
	}
	m := indismq.ParseMsg(&body)
	from := string(m.Fields.From())
	c := s.Connections[from]
	if c == nil || len(c.User) < 1 {
		conType := s.ConnectionTypes["http"]
		if conType == nil {
			log.Println("invalid connection type")
			return
		}
		c = &Connection{User: from, ConnectionType: s.ConnectionTypes["http"]}
	}
	if c.ConnectionType.AuthHandler != nil {
		c.ConnectionType.AuthHandler(r.Header)
	}
	resp := s.Imq.RecieveMessage(m, c.User)
	if resp != nil && resp.Data != nil && w != nil {
		w.Write(*resp.Data)
		return
	}
}

//Listen ...
func (s *Server) ListenHttp(path string, port int) {
	fmt.Println("server starting")
	log.Println("starting http")
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(path, s.receiveHttp)
	go log.Println(http.ListenAndServe(":"+strconv.Itoa(port), serverMux).Error())
}

//ListenTLS ...
func (s *Server) ListenHttpTLS(path string, port int, certFile string, keyFile string, config *tls.Config) {
	if s.Imq.Debug {
		fmt.Println("server starting")
		log.Println("starting http")
	}
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(path, s.receiveHttp)
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   serverMux,
	}
	s.GetCertificate = config.GetCertificate
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if s.Imq.Debug {
			log.Printf("%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Server Error: " + err)
		}
	}()

}
