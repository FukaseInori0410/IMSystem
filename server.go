package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	//map of online users
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//channel in order to broadcast
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, client := range server.OnlineMap {
			client.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	// current conn
	fmt.Println("链接建立成功！")

	user := NewUser(conn, server)

	user.Online()

	isLive := make(chan bool)
	//receive message and broadcast
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
			}
			msg := string(buf[:n-1])

			user.DoMessage(msg)
		}
	}()
	for {
		select {
		case <-isLive:
			//do nothing, but can reset the time
		case <-time.After(time.Minute * 5):
			//kick this user
			user.SendMsg("You are kicked!")
			close(user.C)
			conn.Close()
			return //runtime.Goexit()
		}
	}
}

// Start the Server
func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	go server.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accrpt err:", err)
			continue
		}
		//do handler
		go server.Handler(conn)
	}
}
