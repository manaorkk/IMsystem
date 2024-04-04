package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// v0.2
	OnlineMap map[string]*User
	mapLock   sync.RWMutex //有关同步的机制都在sync

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听并分发消息
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播登陆信息
// 为什么这里不将broadCast和ListenMessage放在一起呢
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	server.Message <- sendMsg
}

// 处理上线信息
func (server *Server) Handler(conn net.Conn) {
	//fmt.Println("链接建立成功！")

	user := NewUser(conn, server)
	user.Online()

	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		// ？这一句是什么意思呢
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			//ctrl+c 返回0
			// ？结束服务期也会结束客户端是为什么
			if n == 0 {
				user.Offline()
				conn.Close()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的信息
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			//不做任何处理
		case <-time.After(time.Second * 300): //添加定时器功能，超时就强踢
			user.SendMessage("You are kicked!")
			user.Offline()
			conn.Close()
			return // runtime.Goexit()

		}
	}

}

func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))

	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	defer listener.Close()
	//启动监听消息
	go server.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue

		}
		// do handler

		go server.Handler(conn)

	}

	// close listen socket

}
