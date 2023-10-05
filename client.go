package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       9,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

func (client *Client) Menu() bool {
	var flag int
	fmt.Println("1.广播模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println("请输入消息内容，exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) ShowUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.ShowUsers()
	time.Sleep(time.Second)
	fmt.Println("请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println("请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
				chatMsg = ""
				fmt.Println("请输入消息内容，exit退出")
				fmt.Scanln(&chatMsg)
			}
		}
		remoteName = ""
		client.ShowUsers()
		time.Sleep(time.Second)
		fmt.Println("请输入聊天对象[用户名]，exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) DealResponse() {
	//conn一旦有数据，就直接copy到stdout，永久阻塞
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.Menu() != true {

		}
		switch client.flag {
		case 1:
			fmt.Println("广播模式已选择")
			client.PublicChat()
		case 2:
			fmt.Println("私聊模式已选择")
			client.PrivateChat()
		case 3:
			fmt.Println("更新用户名已选择")
			client.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

// client.exe -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是8888）")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("链接服务器失败...")
		return
	}
	go client.DealResponse()
	fmt.Println("链接服务器成功！")
	client.Run()
}
