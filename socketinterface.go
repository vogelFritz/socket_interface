package socketinterface

import (
	"log"
	"net"
	"strings"
)

type Server struct {
	listener net.Listener
	sockets  []net.Conn
	rooms    map[string][]net.Conn
	events   map[string]func(data string, socket net.Conn)
}

func (srv *Server) Init(address string) {
	var err error
	srv.sockets = []net.Conn{}
	srv.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Couldn't start")
	}
	srv.addDefaultEventListeners()
}

func (srv *Server) WaitForClients() {
	if srv.listener == nil {
		log.Fatal("Server must be initialized")
	}
	for {
		socket, err := srv.listener.Accept()
		if err != nil {
			log.Fatal("Error accepting client")
		}
		srv.sockets = append(srv.sockets, socket)
		srv.handleConnection(socket)
	}
}

func (srv Server) handleConnection(socket net.Conn) {
	for {
		buffer := make([]byte, 1000)
		mLen, err := socket.Read(buffer)
		if err != nil {
			log.Println("Error reading")
		}
		srv.parseMessage(buffer[:mLen], socket)
	}
}

func (srv Server) parseMessage(msg []byte, socket net.Conn) {
	stringMsg := string(msg)
	for eventName := range srv.events {
		if strings.Contains(stringMsg, eventName) {
			srv.events[eventName](stringMsg[len(eventName):], socket)
		}
	}
}

func (srv Server) AddEventListener(event string, handler func(data string, socket net.Conn)) {
	srv.events[event] = handler
}

func (srv Server) addDefaultEventListeners() {
	srv.AddEventListener("join", func(roomName string, socket net.Conn) {
		srv.rooms[roomName] = append(srv.rooms[roomName], socket)
	})
}

type EmissionParams struct {
	room   string
	socket net.Conn
	event  string
	data   string
}

func (srv Server) Emit(params EmissionParams) {
	if params.room != "" {
		srv.emitToRoom(params.room, params.event, params.data)
	} else if params.socket != nil {
		srv.emitToSocket(params.socket, params.event, params.data)
	} else {
		srv.emitToAllSockets(params.event, params.data)
	}
}

func (srv Server) emitToRoom(room string, event string, data string) {
	sockets := srv.rooms[room]
	for i := range sockets {
		srv.emitToSocket(sockets[i], event, data)
	}
}

func (srv Server) emitToAllSockets(event string, data string) {
	for i := range srv.sockets {
		srv.emitToSocket(srv.sockets[i], event, data)
	}
}

func (srv Server) emitToSocket(socket net.Conn, event string, data string) {
	finalMessage := event + data
	socket.Write([]byte(finalMessage))
}