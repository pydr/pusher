package pusher

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/pydr/logger"
	"go.uber.org/zap"
)

type Client struct {
	IP       string          // 客户端ip地址
	Conn     *websocket.Conn // 连接对象
	Receiver chan []byte     // 接收队列
	Send     chan []byte     // 发送队列

	Locker      sync.Mutex
	IsClosed    bool
	CloseNotice chan bool
	Heart       int      // 心跳
	Subbed      []string // 订阅的主题
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		IP:          conn.RemoteAddr().String(),
		Conn:        conn,
		Receiver:    make(chan []byte, 1024),
		Send:        make(chan []byte, 1024),
		CloseNotice: make(chan bool, 2),
	}
}

func (c *Client) closeClient() {
	for {
		select {
		case <-c.CloseNotice:
			Logger.Info("closing websocket connection", zap.String("ip", c.IP))
			err := c.Conn.Close()
			if err != nil {
				Logger.Error("close client failed", zap.Error(err))
			}
			c.IsClosed = true
			for _, sub := range c.Subbed {
				WsCliMgr.Remove(sub, c)
			}
			close(c.Send)

			return
		}
	}
}

// 心跳
func (c *Client) heart() {
	for {
		c.Heart++
		if c.Heart > 3 {
			// 心跳超时, 关闭连接
			Logger.Info("heart message timeout, connection closing...", zap.String("ip", c.IP))
			c.CloseNotice <- true
		}

		time.Sleep(5 * time.Second)
		if c.IsClosed {
			return
		}
		msg := MakeNewSendMsg(PingAction, nil)
		jsonMsg, _ := json.Marshal(msg)
		c.Send <- jsonMsg
	}
}

type HandleFunc func(*ReceiveMsg) error

// 读取客户端请求
func (c *Client) readMsg(subHandle, unSubHandle HandleFunc) {
	for {
		if !c.IsClosed {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				Logger.Info("connection had closed", zap.String("ip", c.IP), zap.Error(err))
				return
			}

			// 处理msg
			var receiveMsg = new(ReceiveMsg)
			err = json.Unmarshal(message, receiveMsg)
			if err != nil {
				data := MakeNewSendMsg(RejectAction, nil)
				jsonData, _ := json.Marshal(data)

				c.Send <- jsonData
			}

			data := make(map[string]string)
			switch receiveMsg.Action {
			case Pong:
				c.Heart--
				continue

			case Sub:
				err = subHandle(receiveMsg)
				if err != nil {
					resp := MakeNewSendMsg(RejectAction, nil)
					jsonData, _ := json.Marshal(resp)
					c.Send <- jsonData
					continue
				}

				WsCliMgr.Add(receiveMsg.Topic, c)
				data["subbed"] = receiveMsg.Topic
				c.Subbed = append(c.Subbed, receiveMsg.Topic)

			case Unsub:
				err = unSubHandle(receiveMsg)
				if err != nil {
					data := MakeNewSendMsg(RejectAction, nil)
					jsonData, _ := json.Marshal(data)
					c.Send <- jsonData
					continue
				}

				WsCliMgr.Remove(receiveMsg.Topic, c)
				// 从订阅列表里删除
				for i, sub := range c.Subbed {
					if sub == receiveMsg.Topic {
						if i == len(c.Subbed)-1 {
							c.Subbed = c.Subbed[:i]
						} else {
							c.Subbed = append(c.Subbed[:i], c.Subbed[i+1:]...)
						}
						break
					}
				}

				data["unsubbed"] = receiveMsg.Topic

			case Close:
				c.CloseNotice <- true
			}

			resp := MakeNewSendMsg(SuccessRespAction, data)
			jsonData, _ := json.Marshal(resp)
			c.Send <- jsonData
		}
	}
}

// 推送数据到客户端
func (c *Client) writeMsg() {
	defer func() {
		c.CloseNotice <- true
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			if !c.IsClosed {
				err := c.Conn.WriteMessage(websocket.BinaryMessage, message)
				if err != nil {
					Logger.Error("write msg to "+c.IP+"failed", zap.Error(err))
					return
				}
			} else {
				return
			}
		}
	}
}

func (c *Client) Run(subHandle, unSubHandle HandleFunc) {
	go c.closeClient()
	go c.heart()
	go c.readMsg(subHandle, unSubHandle)
	go c.writeMsg()
}
