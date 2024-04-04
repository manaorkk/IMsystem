package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
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
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}

	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var flag int

	fmt.Printf("1. 公聊模式\n")
	fmt.Printf("2. 私聊模式\n")
	fmt.Printf("3. 更改用户名\n")
	fmt.Printf("0. 退出\n")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法范围内的数字<<<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			fmt.Println("公聊模式选择...")
			client.PublicChat()
		case 2:
			fmt.Println("私聊模式选择...")
			client.PrivateChat()
		case 3:
			fmt.Println("更改用户名选择...")
			client.UpdateName()
		}
	}
}

func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) PublicChat() {
	chatMsg := ""
	fmt.Println(">>>>>>请输入聊天内容,exit退出")
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
		fmt.Println(">>>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) PrivateChat() {
	sendMsg := "Online\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
	}

	toClient := ""
	fmt.Println(">>>>>>请输入聊天对象,exit退出")
	fmt.Scanln(&toClient)

	for toClient != "exit" {
		chatMsg := ""
		fmt.Println(">>>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + toClient + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>请输入聊天内容,exit退出")
			fmt.Scanln(&chatMsg)
		}

		toClient = ""
		fmt.Println(">>>>>>请输入聊天对象,exit退出")
		fmt.Scanln(&toClient)

	}

}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}

	return true

}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> failure to link server...")
		return
	}

	fmt.Println(">>>>>> success to link server...")

	go client.DealResponse()
	//启动客户端业务
	client.Run()
}
