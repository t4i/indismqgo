package indismqgo

import ()

//NewContext Creates a New Imq Context with he given name
func NewContext(name string) (ctx *Context) {
	ctx = new(Context)

	ctx.Debug = false
	ctx.Handlers = &Handlers{handlers: make(map[string]Handler)}
	ctx.Messages = &Messages{messages: make(map[string]*MsgBuffer)}
	ctx.Subscribers = &Subscribers{subscribers: make(map[string]map[string]bool)}
	ctx.Connections = &Connections{connections: make(map[string]Connection)}
	ctx.Queues = &Queues{queues: make(map[string]*Queue)}
	if len(name) > 0 {
		ctx.Name = []byte(name)
	} else {
		ctx.Name = NewRandID()
	}
	return ctx
}

//Context is the scope in which Imq operates it defines the handlers, Subscribers and Connections
type Context struct {
	Handlers          *Handlers
	Debug             bool
	Name              []byte
	UnknownHandler    Handler
	ConnectionHandler Handler
	Messages          *Messages
	Subscribers       *Subscribers
	Connections       *Connections
	Queues            *Queues
}
