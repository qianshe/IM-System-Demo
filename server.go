package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// NewServer 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessage 监听Message
func (ser *Server) ListenMessage() {
	for {
		msg := <-ser.Message

		//将msg发送给所有在线user
		ser.mapLock.Lock()
		for _, cli := range ser.OnlineMap {
			cli.C <- msg
		}
		ser.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (ser *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	ser.Message <- sendMsg
}

func (ser *Server) Handler(conn net.Conn) {
	// ... 当前链接的业务
	fmt.Println("链接建立成功!")

	user := NewUser(conn)

	//用户上限，将用户加入onlineMap中
	ser.mapLock.Lock()
	ser.OnlineMap[user.Name] = user
	ser.mapLock.Unlock()

	//广播当前用户上线消息
	ser.BroadCast(user, "已上线")

	//当前handler阻塞
	select {}
}

// Start 启动服务器的接口
func (ser *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ser.Ip, ser.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go ser.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go ser.Handler(conn)
	}

}
