package server

import (
	"crypto/tls"
	"errors"
	"github.com/gorilla/websocket"
	indismq "github.com/t4i/indismqgo"
	"net/http"
	"net/url"
	"sync"
)

//Connection ...
type Connection struct {
	Name           string
	User           string
	ConnectionType *ConnectionType
	URL            *url.URL
	Headers        map[string][]string
}

//NewConnection ...
func (s *Server) NewConnection(name string, user string, connectionType string, address *url.URL, headers map[string][]string) (*Connection, error) {
	t, ok := s.ConnectionTypes[connectionType]
	if !ok {
		return nil, errors.New("Invalid Connection Type:" + connectionType)
	}
	return &Connection{Name: name, User: user, ConnectionType: t, URL: address, Headers: headers}, nil
}

type ConnectionType struct {
	Name          string
	DefaultSender SenderFunc
	AuthHandler   AuthHandlerFunc
}
type AuthHandlerFunc func(headers http.Header) (bool, error)
type SenderFunc func(m *indismq.Msg, c *Connection) error

//Server ...
type Server struct {
	Imq             *indismq.Imq
	webSockets      map[string]*websocket.Conn
	ConnectionTypes map[string]*ConnectionType
	Connections     map[string]*Connection
	upgrader        websocket.Upgrader
	sendLocks       map[*websocket.Conn]*sync.Mutex

	//GetCertificate ..
	GetCertificate func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

	//OnMsgRecieved ...
	OnMsgRecieved func(m *indismq.Msg, c *Connection) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnBroker ...
	OnBroker func(m *indismq.Msg) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnSyn ...
	OnSyn func(r *http.Request, c *Connection) (ok bool, errMsg string, err int8)

	//OnRelay ...
	OnRelay func(m *indismq.Msg) (ok bool, subM *indismq.Msg, errMsg string, err int8)

	//OnUnknownClient ...
	OnUnknownClient func(m *indismq.Msg) (forward bool, to string)
	//OnDisconnected ...
	OnServerDisconnected func(c *Connection)
	//OnConnected ...
	OnConnected func(c *Connection)

	//OnReady ...
	OnReady func(c *Connection)
	//OnClientDisconnected ...
	OnClientDisconnected func(c *Connection)
}

func (s *Server) NewConnectionType(name string, defaultSender SenderFunc, authHandler AuthHandlerFunc) {
	s.ConnectionTypes[name] = &ConnectionType{Name: name, DefaultSender: defaultSender, AuthHandler: authHandler}
}

//NewServer ...
func NewServer(name string) *Server {
	var s = new(Server)
	s.Imq = indismq.NewImq()
	s.webSockets = make(map[string]*websocket.Conn)
	s.ConnectionTypes = make(map[string]*ConnectionType)
	s.Connections = make(map[string]*Connection)
	s.upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options
	s.sendLocks = make(map[*websocket.Conn]*sync.Mutex)
	s.Imq.SetName(name)
	s.Imq.SetRelayHandler(s.relayHandler)
	s.Imq.SetBrokerHandler(s.brokerHandler)
	return s
}

func (s *Server) ConnectHTTP(address string, name string, header http.Header, user string) (ok bool, err error) {

	return
}
