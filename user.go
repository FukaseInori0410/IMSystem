package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// create a new user
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//start a goroutine to listen to the user channel
	go user.ListenMessage()
	return user
}

func (user *User) Online() {
	//add the user to the map
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//broadcast the online message
	user.server.BroadCast(user, "is online now")
}

func (user *User) Offline() {
	//move the user out of the map
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//broadcast the online message
	user.server.BroadCast(user, "is offline now")
}

func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *User) DoMessage(msg string) {
	//to inquire who is online
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, onlineUser := range user.server.OnlineMap {
			onlineMsg := "[" + onlineUser.Addr + "]" + onlineUser.Name + ":" + "is online\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//to change your user name
		newName := strings.Split(msg, "|")[1]
		//
		if _, ok := user.server.OnlineMap[newName]; ok {
			user.SendMsg("This name has been used.")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("Your new name is:" + user.Name + " now.\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		//some legality test
		if remoteName == "" {
			user.SendMsg("Incorrect format! You should send msg like \"to|zhang3|nihao\" to use this function")
			return
		}
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("This user name does not exist.")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("Blank message, please try again!")
			return
		}
		remoteUser.SendMsg(user.Name + "said to you:" + content)
	} else {
		user.server.BroadCast(user, msg)
	}
}

// listen to the user channel and write it
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
