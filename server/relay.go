package server

import (
	indismq "github.com/t4i/indismqgo"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
)

func (s *Server) relayHandler(m *indismq.Msg) *indismq.Msg {
	if s.OnRelay != nil {
		ok, subM, errMsg, err := s.OnRelay(m)
		if !ok {
			return s.Imq.Err(m, errMsg, err, string(m.Fields.User()))
		}
		if subM != nil {
			m = subM
		}
	}
	c := s.Connections[string(m.Fields.To())]
	if c != nil && c.ConnectionType != nil && len(c.Name) < 1 {
		c.ConnectionType.DefaultSender(m, c)
		return nil
	} else if s.OnUnknownClient != nil {
		forward, to := s.OnUnknownClient(m)
		if forward {
			c = s.Connections[to]
			if c != nil && c.ConnectionType != nil && len(c.Name) < 1 {
				c.ConnectionType.DefaultSender(m, c)
				return nil
			}
		}
	}

	return s.Imq.Err(m, "Client not found", schema.ErrINVALID, string(m.Fields.User()))

}
