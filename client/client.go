// 客户端

// 我们先来实现相对简单的客户端，客户端主要做的事情是 3 件：
// 1.连接服务端的控制通道
// 2.等待服务端从控制通道中发来建立连接的消息
// 3.收到建立连接的消息时，将本地服务和远端隧道建立连接（这里就要用到我们的工具方法了）

package main

import (
	"bufio"
	"io"
	"log"
	"myCode/project/IntranetThrough/network"
	"net"
	"time"
)

var (
	//本地暴露的服务端口
	localServerAddr = "127.0.0.1:3389"

	remoteIp = "47.103.216.249"
	//远端的服务控制通道，用来传递控制信息，如出现新连接和心跳
	remoteControlAddr = remoteIp + ":8809"
	//远端服务端口，用来建立隧道
	remoteServerAddr = remoteIp + ":8808"
)

func main() {
	tcpConn, err := network.CreateTcpConn(remoteControlAddr)
	if err != nil {
		log.Println("[连接失败]" + remoteControlAddr + err.Error())
		return
	}
	log.Println("[已连接]" + remoteControlAddr)

	reader := bufio.NewReader(tcpConn)
	time.Sleep(2)
	for {
		s, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		if s == network.NewConnection+"\n" {
			go connectionLocalAndRemote()
		}
	}

	log.Println("[已断开]" + remoteControlAddr)
}

func connectionLocalAndRemote() {
	local := connectionLocal()
	remote := connectionRemote()

	if local != nil && remote != nil {
		network.Join2Conn(local, remote)
	} else {
		if local != nil {
			_ = local.Close()
		}
		if remote != nil {
			_ = remote.Close()
		}
	}
}

func connectionLocal() *net.TCPConn {
	tcpConn, err := network.CreateTcpConn(localServerAddr)
	if err != nil {
		log.Println("[连接本地服务失败]" + err.Error())
		return nil
	}
	return tcpConn
}

func connectionRemote() *net.TCPConn {
	tcpConn, err := network.CreateTcpConn(remoteServerAddr)
	if err != nil {
		log.Println("[连接远程服务失败]" + err.Error())
		return nil
	}
	return tcpConn
}
