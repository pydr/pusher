package pusher

import (
	"fmt"
	"sync"
	"time"

	. "github.com/pydr/logger"
	"github.com/pydr/utils"
)

type WsClientManager struct {
	Clients map[string][]*Client // 客户端们
	Locker  sync.RWMutex
	//Broadcast chan []byte // 广播数据
}

var WsCliMgr *WsClientManager

func init() {
	WsCliMgr = NewClientManager()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			l := WsCliMgr.Length()
			Logger.Debug(fmt.Sprintf("[Websocket status] -- > Current %d connections", l))
		}
	}()
}

func NewClientManager() *WsClientManager {
	return &WsClientManager{
		Clients: make(map[string][]*Client),
		//Broadcast: make(chan []byte, 1024),
	}
}

// 获取所有连接数
func (m *WsClientManager) Length() int {
	count := 0
	for _, cls := range m.Clients {
		count += len(cls)
	}

	return count
}

// 添加客户端
func (m *WsClientManager) Add(sub string, c *Client) {
	m.Locker.Lock()
	defer m.Locker.Unlock()
	if !utils.Contains(m.Clients[sub], c) {
		m.Clients[sub] = append(m.Clients[sub], c)
	}
}

// 移除客户端
func (m *WsClientManager) Remove(sub string, c *Client) {
	m.Locker.Lock()
	defer m.Locker.Unlock()

	for i, cli := range m.Clients[sub] {
		if cli == c {
			m.Clients[sub] = append(m.Clients[sub][:i], m.Clients[sub][i+1:]...)
		}
	}
}
