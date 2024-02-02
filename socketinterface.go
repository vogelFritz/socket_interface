package socketinterface

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Socket struct {
	Conn net.Conn
	Room string
}

type Server struct {
	listener net.Listener
	sockets  []Socket
	rooms    map[string][]net.Conn
	events   map[string]func(data string, socket Socket)
}

func (srv *Server) Init(address string) {
	var err error
	srv.sockets = []Socket{}
	srv.events = map[string]func(data string, socket Socket){}
	srv.rooms = map[string][]net.Conn{}
	fmt.Printf("Server starting at %v\n", address)
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
		conn, err := srv.listener.Accept()
		if err != nil {
			log.Fatal("Error accepting client")
		}
		log.Println("Client connected")
		newSocket := Socket{Conn: conn}
		srv.sockets = append(srv.sockets, newSocket)
		go srv.handleConnection(newSocket)
	}
}

func (srv Server) handleConnection(socket Socket) {
	for {
		buffer := make([]byte, 1000)
		mLen, err := socket.Conn.Read(buffer)
		if err != nil {
			log.Println("Error reading")
			break
		}
		srv.parseMessage(buffer[:mLen], socket)
	}
}

func (srv Server) parseMessage(msg []byte, socket Socket) {
	stringMsg := string(msg)
	for eventName := range srv.events {
		if strings.Contains(stringMsg, eventName) {
			srv.events[eventName](stringMsg[len(eventName):], socket)
		}
	}
}

func (srv Server) AddEventListener(event string, handler func(data string, socket Socket)) {
	srv.events[event] = handler
}

func (srv Server) addDefaultEventListeners() {
	srv.AddEventListener("join", func(roomName string, socket Socket) {
		srv.AddToRoom(roomName, socket)
	})
}

type EmissionParams struct {
	Room   string
	Socket net.Conn
	Event  string
	Data   string
}

func (srv Server) AddToRoom(roomName string, socket Socket) {
	socket.Room = roomName
	srv.rooms[roomName] = append(srv.rooms[roomName], socket.Conn)
}

func (srv Server) Emit(params EmissionParams) {
	if params.Room != "" && params.Socket != nil {
		log.Fatal("When using the Emit method, specify either a room or a socket, not both")
	} else if params.Room != "" {
		srv.emitToRoom(params.Room, params.Event, params.Data)
	} else if params.Socket != nil {
		srv.emitToSocket(params.Socket, params.Event, params.Data)
	} else {
		srv.emitToAllSockets(params.Event, params.Data)
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
		srv.emitToSocket(srv.sockets[i].Conn, event, data)
	}
}

func (srv Server) emitToSocket(socket net.Conn, event string, data string) {
	finalMessage := event + data
	socket.Write([]byte(finalMessage))
}
