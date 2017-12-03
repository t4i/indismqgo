package http

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/t4i/indismqgo"
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

var _ indismqgo.Connection = (*HttpConn)(nil)

type Events struct {
	// OnBeforeConnected func(conn *HttpConn, req *http.Request) bool
	// OnConnected       func(m *indismqgo.MsgBuffer, conn *HttpConn) bool
	// OnDisconnected    func(conn *HttpConn)
	OnMessage       func(m *indismqgo.MsgBuffer, r *http.Request) error
	OnBeforeMessage func(r *http.Request) error
}

type HttpConn struct {
	URL     *url.URL
	Req     *http.Request
	Context indismqgo.Context
	Events
}

// func (c *HttpConn) OnBeforeConnected(m *indismq.MsgBuffer, conn *HttpConn) bool {

// 	return true
// }

// func (c *HttpConn) OnUnknown(m *indismq.MsgBuffer, conn *HttpConn) bool {
// 	return true
// }
// func (c *HttpConn) OnBeforeSubscribe(m *indismq.MsgBuffer, conn *HttpConn) bool {
// 	return true
// }

// type ConnEvents struct {
// 	OnConnect    func(key string, conn Connection)
// 	OnDisconnect func(key string, conn Connection)
// }

func NewHttpConn(ctx indismqgo.Context, URL *url.URL, headers http.Header, ev *Events) *HttpConn {
	conn := &HttpConn{URL: URL, Context: ctx}
	if ev != nil {
		conn.Events = *ev
	}
	conn.Req = &http.Request{Header: headers}

	return conn
}
func (conn *HttpConn) VerifyConnectionRequest(m *indismqgo.MsgBuffer) error {
	return nil
}
func (conn *HttpConn) VerifySubscribeRequest(m *indismqgo.MsgBuffer) error {
	return nil
}

var HttpRetryDelay time.Duration = time.Second * 5

//SendMessage ...
func (conn *HttpConn) Send(m *indismqgo.MsgBuffer) error {
	// httpConn, ok := conn.Conn.(*HttpConn)
	// if !ok {
	// 	return false, errorctx.New("Invalid http conn")
	// }
	if m != nil && m.Data != nil {
		if conn.Context.Debug(nil) {
			fmt.Println(m.String())
		}
		client := http.Client{Timeout: writeWaitHTTP}
		buf := bytes.NewBuffer(m.Data)
		req, err := http.NewRequest("POST", conn.URL.String(), buf)
		if err != nil {
			log.Println(err)
		}
		req.Header = conn.Req.Header
		go func() {
			if messages, ok := conn.Context.(indismqgo.MessageStore); ok {
				messages.SetMessage(string(m.Fields.MsgId()), m)
			}
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
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if body == nil || len(body) < 5 {
				return
			}
			if res.StatusCode != http.StatusOK {
				log.Println(string(body))
				return
			}
			rm := indismqgo.ParseMsg(body, conn.Context)
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

//Connect ...
func ConnectHttp(ctx indismqgo.Context, conn *HttpConn) (ok bool, err error) {
	//conn := ctx.NewConnType(ctx.SendHttpMessage, indismq.ConnClassHalfDuplex)

	connectMsg, err := indismqgo.NewCtxConnectionMsg(ctx, nil, nil, func(m *indismqgo.MsgBuffer, c indismqgo.Connection) error {
		log.Println("connected", string(m.Fields.From()))

		return nil
	})
	conn.Send(connectMsg)
	return true, nil
}

type AsyncWait struct {
	done chan *indismqgo.MsgBuffer
}

func (s *AsyncWait) Send(m *indismqgo.MsgBuffer) error {
	// if ctx.Debug {
	// 	log.Println("waitfor response called")
	// 	// conType := reflect.TypeOf(conn.Conn)
	// 	// log.Println(conType)
	// }
	// done, ok := conn.Conn.(chan *indismq.MsgBuffer)
	// if !ok {
	// 	return false, errorctx.New("Not a valid Http Connection")
	// }
	s.done <- m
	return nil
}

func receiveHttp(ctx indismqgo.Context, w http.ResponseWriter, r *http.Request, ev *Events) {
	if ctx.Debug(nil) {
		log.Println("receive http")
	}
	if ev != nil && ev.OnBeforeMessage != nil {
		if ev.OnBeforeMessage(r) != nil {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}
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
	m := indismqgo.ParseMsg(body, ctx)
	if ev != nil && ev.OnMessage != nil {
		if ev.OnMessage(m, r) != nil {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}
	}
	// if ctx.HTTPAuthHandler != nil {
	// 	auth, err = ctx.HTTPAuthHandler(r.Header.Get("Authorization"))
	// 	if err != nil {
	// 		auth = false
	// 	}
	// }
	var resp *indismqgo.MsgBuffer

	if ctx.Debug(nil) {
		log.Println("Recieve", r.Host, r.RequestURI, r.URL.Host)
	}
	done := make(chan *indismqgo.MsgBuffer, 1)
	//ctx.NewConnType(ctx.WaitForHttpResponse, indismq.ConnClassHalfDuplex)

	err = ctx.Recieve(m, &AsyncWait{done: done})
	if err != nil {
		log.Println(err)
		return
	}
	if resp == nil {
		if ctx.Debug(nil) {
			log.Println(string(ctx.Name(nil)), "waiting for response")
		}
		resp = <-done
		if ctx.Debug(nil) {
			log.Println("response received")
		}
	}

	if resp != nil && resp.Data != nil && w != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(resp.Data)
		return
	}
	log.Println("empty response")

}

//Listen ...
func ListenHttp(ctx indismqgo.Context, path string, port int, ev *Events) {
	if ctx.Debug(nil) {
		log.Println("starting http", string(ctx.Name(nil)))
	}

	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		receiveHttp(ctx, w, r, ev)
	})
	log.Println(http.ListenAndServe(":"+strconv.Itoa(port), BrokerMux).Error())
}

func ContextHandler(ctx indismqgo.Context, ev *Events) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receiveHttp(ctx, w, r, ev)
	}
}

//ListenTLS ...
func ListenHttpTLS(ctx indismqgo.Context, path string, port int, certFile string, keyFile string, config *tls.Config, ev *Events) {
	if ctx.Debug(nil) {
		log.Println("starting https", string(ctx.Name(nil)))
	}
	BrokerMux := http.NewServeMux()
	BrokerMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		receiveHttp(ctx, w, r, ev)
	})
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		TLSConfig: config,
		Handler:   BrokerMux,
	}
	go func() {
		err := srv.ListenAndServeTLS(certFile, keyFile).Error()
		if ctx.Debug(nil) {
			log.Printf("%s-%s", path, strconv.Itoa(port))
			log.Println("TLS Broker Error: " + err)
		}
	}()

}
