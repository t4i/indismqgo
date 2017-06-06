package broker

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	indismq "github.com/t4i/indismqgo"
)

type AuthHandlerFunc func(auth string) (bool, error)

//Server ...
type Broker struct {
	upgrader        websocket.Upgrader
	GetCertificate  func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
	HTTPAuthHandler AuthHandlerFunc
	WsAuthHandler   AuthHandlerFunc
	*indismq.Context
}

// func (s *Server) NewConnectionType(name string, defaultSender SenderFunc, authHandler AuthHandlerFunc) {
// 	s.ConnectionTypes[name] = &ConnectionType{Name: name, DefaultSender: defaultSender, AuthHandler: authHandler}
// }

//NewServer ...
func NewBroker(name string) *Broker {
	var broker = &Broker{}
	broker.Context = indismq.NewContext(name)
	return broker
}
