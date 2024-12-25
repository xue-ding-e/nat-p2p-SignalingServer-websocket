package signaling

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/rand"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求
		},
	}
	clients = make(map[string]*websocket.Conn)
	mutex   = &sync.Mutex{}
)

type Message struct {
	Type     string      `json:"type"`
	TargetID string      `json:"targetUserId,omitempty"`
	FromID   string      `json:"fromUserId,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

func generateUserID() string {
	// 生成随机用户ID
	return "user_" + randomString(8)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}

	userID := generateUserID()

	mutex.Lock()
	clients[userID] = conn
	mutex.Unlock()

	// 发送用户ID给客户端
	conn.WriteJSON(Message{
		Type: "userId",
		Data: userID,
	})

	defer func() {
		mutex.Lock()
		delete(clients, userID)
		mutex.Unlock()
		conn.Close()
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("读取消息失败:", err)
			break
		}

		// 添加消息类型日志
		log.Printf("收到消息类型: %s, 从: %s, 发送给: %s\n", msg.Type, userID, msg.TargetID)

		// 转发消息给目标用户
		if targetConn, ok := clients[msg.TargetID]; ok {
			msg.FromID = userID
			err = targetConn.WriteJSON(msg)
			if err != nil {
				log.Printf("发送消息到 %s 失败: %v\n", msg.TargetID, err)
				continue
			}
			log.Printf("成功转发消息到 %s\n", msg.TargetID)
		} else {
			log.Printf("目标用户 %s 不存在\n", msg.TargetID)
		}
	}
}

func StartSignalingServer() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("信令服务器启动在 :60042")
	log.Fatal(http.ListenAndServe(":60042", nil))
}

func randomString(length int) string {
	// 实现一个简单的随机字符串生成函数
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
