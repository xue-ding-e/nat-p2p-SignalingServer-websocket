package stun

import (
	"log"
	"net"

	"github.com/pion/stun"
)

func StartSTUNServer() {
	addr := "0.0.0.0:3478"
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		log.Fatal("无法解析地址:", err)
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatal("监听失败:", err)
	}
	defer conn.Close()

	log.Printf("STUN 服务器启动在 %s\n", addr)

	for {
		buf := make([]byte, 1024)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("读取数据失败:", err)
			continue
		}

		// 打印接收到的请求信息
		log.Printf("收到来自 %s 的请求\n", remoteAddr.String())

		message := &stun.Message{
			Raw: buf[:n],
		}

		if err := message.Decode(); err != nil {
			log.Println("解码STUN消息失败:", err)
			continue
		}

		// 打印消息类型
		log.Printf("收到STUN消息类型: %s\n", message.Type.String())

		if message.Type == stun.BindingRequest {
			// 创建响应消息
			response := &stun.Message{
				Type: stun.BindingSuccess,
			}

			// 创建 XOR-MAPPED-ADDRESS 属性
			// 这是STUN协议的核心部分，用于告诉客户端它的公网地址
			xorAddr := &stun.XORMappedAddress{
				IP:   remoteAddr.IP,
				Port: remoteAddr.Port,
			}

			// 构建响应，包含必要的属性
			if err := response.Build(
				stun.BindingSuccess,
				xorAddr,                    // XOR-MAPPED-ADDRESS
				stun.NewSoftware("GoSTUN"), // SOFTWARE
				stun.Fingerprint,           // FINGERPRINT
			); err != nil {
				log.Println("构建响应失败:", err)
				continue
			}

			log.Printf("发送响应到 %s，包含XOR地址: %s:%d\n",
				remoteAddr.String(),
				remoteAddr.IP.String(),
				remoteAddr.Port)

			// 发送响应
			if _, err := conn.WriteToUDP(response.Raw, remoteAddr); err != nil {
				log.Println("发送响应失败:", err)
				continue
			}

			log.Println("响应发送成功")
		}
	}
}
