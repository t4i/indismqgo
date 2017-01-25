package server

import (
	indismq "github.com/t4i/indismqgo"
)

func (s *Server) brokerHandler(m *indismq.Msg) *indismq.Msg {
	if s.OnBroker != nil {
		ok, subM, errMsg, err := s.OnBroker(m)
		if !ok {
			return s.Imq.Err(m, errMsg, err, string(m.Fields.User()))
		}
		if subM != nil {
			m = subM
		}
	}

	var callback indismq.Handler
	if m.Fields.Callback() != 0 {
		callback = func(imqMessage *indismq.Msg) *indismq.Msg {
			c := s.Connections[string(m.Fields.From())]
			if c != nil && c.ConnectionType != nil && len(c.Name) < 1 {
				c.ConnectionType.DefaultSender(imqMessage, c)
				return nil
			}
			return nil
		}

	}
	s.Imq.BrokerReplay(m, func(client string, imqMessage *indismq.Msg) {
		c := s.Connections[client]
		if c != nil && c.ConnectionType != nil && len(c.Name) < 1 {
			c.ConnectionType.DefaultSender(imqMessage, c)
		} else if s.OnUnknownClient != nil {
			forward, to := s.OnUnknownClient(m)
			if forward {
				c = s.Connections[to]
				if c != nil && c.ConnectionType != nil && len(c.Name) < 1 {
					c.ConnectionType.DefaultSender(m, c)
				}
			}
		}
	}, callback)
	return nil
}
