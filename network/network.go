// 工具方法

// 首先我们先定义三个需要使用的工具方法，还需要定义两个消息编码常量，后面会用到
// 1.监听一个地址对应的 TCP 请求 CreateTCPListener
// 2.连接一个 TCP 地址 CreateTCPConn
// 3.将一个 TCP-A 连接的数据写入另一个 TCP-B 连接，将 TCP-B 连接返回的数据写入 TCP-A 的连接中 Join2Conn （别看这短短 10 几行代码，这就是核心了）

package network

import (
	"io"
	"log"
	"net"
)

const (
	KeepAlive     = "KEEP_ALIVE"
	NewConnection = "NEW_CONNECTION"
)

func CreateTCPListener(addr string) (*net.TCPListener, error) {
	//将addr作为TCP地址解析并返回
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	//建立tcp监听
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return lis, nil
}

func CreateTcpConn(addr string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return tcpConn, nil
}

func Join2Conn(local *net.TCPConn, remote *net.TCPConn) {
	go JoinConn(local, remote)
	// go JoinConn(remote, local)
}

// Copy copies from remote to local
func JoinConn(local *net.TCPConn, remote *net.TCPConn) {
	defer local.Close()
	defer remote.Close()
	_, err := io.Copy(local, remote)
	if err != nil {
		log.Println(err)
		log.Println("Copy local remote failed.")
		return
	}
}
