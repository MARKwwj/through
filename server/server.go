// 服务端

// 服务端的实现就相对复杂一些了：
// 1.监听控制通道，接收客户端的连接请求
// 2.监听访问端口，接收来自用户的 http 请求
// 3.第二步接收到请求之后需要存放一下这个连接并同时发消息给客户端，告诉客户端有用户访问了，赶紧建立隧道进行通信

package main

import (
	"log"
	"myCode/project/IntranetThrough/network"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	controlAddr = "0.0.0.0:8809"
	tunnelAddr  = "0.0.0.0:8808"
	visitAddr   = "0.0.0.0:8899"
)

var (
	clientConn         *net.TCPConn
	connectionPool     map[string]*ConnMatch
	connectionPoolLock sync.Mutex
)

type ConnMatch struct {
	addTime time.Time
	accept  *net.TCPConn
}

func main() {
	connectionPool = make(map[string]*ConnMatch, 32)
	go createControlChannel()
	go acceptUserRequest()
	go acceptClientRequest()
	cleanConnectionPool()
}

//创建一个控制通道，用于传递控制消息，如：心跳，创建新连接
func createControlChannel() {
	tcpListener, err := network.CreateTCPListener(controlAddr)
	if err != nil {
		panic(err)
	}

	log.Println("[已监听]" + controlAddr)
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("[新链接]" + tcpConn.RemoteAddr().String())
		//如果当前已经有一个客户端存在，则丢弃这个连接
		if clientConn != nil {
			tcpConn.Close()
		} else {
			clientConn = tcpConn
			go keepAlive()
		}
	}
}

//和客户端保持一个心跳连接
func keepAlive() {
	go func() {
		for {
			if clientConn == nil {
				return
			}

			_, err := clientConn.Write(([]byte)(network.KeepAlive + "\n"))
			if err != nil {
				log.Println("已断开客户端连接", clientConn.RemoteAddr())
				clientConn = nil
				return
			}

			time.Sleep(time.Second * 3)
		}
	}()
}

//监听来自用户的请求
func acceptUserRequest() {
	tcpListener, err := network.CreateTCPListener(visitAddr)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}

		addConn2Pool(tcpConn)
		sendMessage(network.NewConnection + "\n")
	}
}

//将用户来的连接放入连接池中
func addConn2Pool(accept *net.TCPConn) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()

	now := time.Now()
	connectionPool[strconv.FormatInt(now.UnixNano(), 10)] = &ConnMatch{now, accept}
}

//发送给客户端新消息
func sendMessage(message string) {
	if clientConn == nil {
		log.Println("[无已连接的客户端]")
		return
	}

	_, err := clientConn.Write([]byte(message))
	if err != nil {
		log.Panicln("[发送消息异常]: message:" + message)
	}
}

//接受客户端的请求并建立隧道
func acceptClientRequest() {
	tcpListener, err := network.CreateTCPListener(tunnelAddr)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		go establishTunnel(tcpConn)
	}
}

func establishTunnel(tunnel *net.TCPConn) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()

	for key, connMatch := range connectionPool {
		if connMatch.accept != nil {
			go network.Join2Conn(tunnel, connMatch.accept)
			delete(connectionPool, key)
			return
		}
	}
	_ = tunnel.Close()
}

func cleanConnectionPool() {
	for {
		connectionPoolLock.Lock()
		for key, connMatch := range connectionPool {
			if time.Since(connMatch.addTime) > time.Second*10 {
				_ = connMatch.accept.Close()
				delete(connectionPool, key)
			}
		}
		connectionPoolLock.Unlock()
		time.Sleep(time.Second * 5)
	}
}
