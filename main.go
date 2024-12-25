package main

import (
	"xuedinge/signaling"
	"xuedinge/stun"
)

func main() {
	// 启动STUN服务器（在3478端口）
	go stun.StartSTUNServer()

	// 启动WebSocket信令服务器（在60042端口）
	signaling.StartSignalingServer()
}
