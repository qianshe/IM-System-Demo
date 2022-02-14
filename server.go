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
	//fmt.Println("链接建立成功!")

	user := NewUser(conn, ser)

	user.Online()
	//监听用户是否活跃
	isLive := make(chan bool)

	//接受客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])
			//将消息广播
			user.DoMessage(msg)

			//用户的消息刷新活跃度
			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			//用户活跃，为了激活select，重置超时时间
		case <-time.After(time.Second * 20):
			//已经超时
			//将当前User强制关闭
			user.SendMsg("你被T了")

			user.Offline()

			//销毁资源
			close(user.C)
			//关闭链接
			conn.Close()
			//退出当前Handler
			return //runtime.Goexit()
		}
	}
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
