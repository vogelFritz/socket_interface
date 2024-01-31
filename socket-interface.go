package socketinterface

import (
	"log"
	"net"
	"strings"
)

type Server struct {
	listener net.Listener
	sockets  []net.Conn
	rooms    map[string]net.Conn
	events   map[string]func()
}

func (srv *Server) Init(address string) {
	var err error
	srv.sockets = []net.Conn{}
	srv.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Couldn't start")
	}
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
			srv.events[eventName]()
		}
	}
}

func (srv Server) AddEventListener(event string, handler func()) {
	srv.events[event] = handler
}
