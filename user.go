package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// constructing function
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

func (u *User) Online() {

	//将用户加入表中，广播当前用户上线消息
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}

func (u *User) Offline() {
	u.server.BroadCast(u, "下线")
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//u=nil 立即释放内存的方法；通常当一个对象不再被引用时，它将会被垃圾回收器回收
}

func (user *User) DoMessage(msg string) {
	if msg == "Online" {
		// 查询当前在线用户
		// ? 也可以创立专门的sendmessage方法（示例），所以为什么不用管道呢，功能完全相同啊
		user.server.mapLock.Lock()
		for _, cli := range user.server.OnlineMap {
			sendMsg := "[" + cli.Addr + "]" + cli.Name + ": Online..."
			user.SendMessage(sendMsg)
		}

		user.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		user.server.mapLock.Lock()
		_, ok := user.server.OnlineMap[newName]
		if ok {
			sendMsg := "The new name is already existing..."

			user.SendMessage(sendMsg)
		} else {
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			// 所有消息的显示实际上都没有取map的key，而是取user结构体中的属性，不能忘记改变结构体中的属性
			user.Name = newName
			sendMsg := "Your name has changed to " + newName
			user.SendMessage(sendMsg)
		}
		user.server.mapLock.Unlock()

	} else if len(msg) > 3 && msg[:3] == "to|" {
		toName := strings.Split(msg, "|")[1]

		user.server.mapLock.Lock()
		toUser, ok := user.server.OnlineMap[toName]
		if ok {
			sendMsg := strings.Split(msg, "|")[2]
			if sendMsg == "" {
				user.SendMessage("there is no message sent")
			} else {
				sendMsg = user.Name + " tell you: " + sendMsg
				toUser.SendMessage(sendMsg)
			}
		} else {
			user.SendMessage("The target is offline!")
		}
		user.server.mapLock.Unlock()

	} else {
		user.server.BroadCast(user, msg)
	}

}

func (user *User) SendMessage(msg string) {
	user.conn.Write([]byte(msg + "\n"))
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C

		user.conn.Write([]byte(msg + "\n"))

	}
}
