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

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	go user.ListenMessage()

	return user
}

//用户上线
func (user *User) Online() {
	//用户上线，将用户加入onlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//广播当前用户上线消息
	user.server.BroadCast(user, "已上线")
}

//用户下线
func (user *User) Offline() {
	//用户下线，将用户从onlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//广播当前用户下线消息
	user.server.BroadCast(user, "已下线")
}

//给当前用户发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

//处理消息
func (user *User) DoMessage(msg string) {
	if msg == "list" {
		//查询当前用户列表
		user.server.mapLock.Lock()
		for _, user1 := range user.server.OnlineMap {
			user.SendMsg("[" + user1.Addr + "]" + user1.Name + ":" + "在线...\n")
		}
		user.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式  rename|张三
		newName := strings.Split(msg, "|")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("当前用户名已存在！\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("用户名更新成功！：" + newName + "\n")
		}
	} else {
		user.server.BroadCast(user, msg)
	}
}

// ListenMessage 监听当前User channel的方法，有消息就发给客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C

		user.conn.Write([]byte(msg + "\n"))
	}
}
