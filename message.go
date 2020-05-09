package pusher

import "time"

type (
	// 接收数据
	ReceiveMsg struct {
		Action    string `json:"action"`
		Pong      int64  `json:"pong,omitempty"`
		Topic     string `json:"topic,omitempty"`
		Timestamp int64  `json:"timestamp"`
	}

	// 消息动作种类
	Action struct {
		Action  string `json:"action"`
		Message string `json:"message,omitempty"`
		Status  bool   `json:"status"`
	}

	// 发送数据
	SendMsg struct {
		*Action
		Ping      int64       `json:"ping,omitempty"`
		Timestamp int64       `json:"timestamp"`
		Data      interface{} `json:"data,omitempty"`
	}
)

// 消息action
const (
	Ping   = "ping"   // 心跳 ping
	Pong   = "pong"   // 心跳 pong
	Sub    = "sub"    // 订阅
	Unsub  = "unsub"  // 取消订阅
	Close  = "close"  // 关闭连接
	Reject = "reject" // 拒绝
	Notice = "notice" // 通知
	Resp   = "resp"   // 响应
)

func MakeNewSendMsg(action *Action, data interface{}) *SendMsg {
	var ping int64
	if action.Action == Ping {
		ping = time.Now().Unix()
	}
	return &SendMsg{
		Action:    action,
		Ping:      ping,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}
}

func makeAction(action, message string, status bool) *Action {
	return &Action{
		Action:  action,
		Message: message,
		Status:  status,
	}
}

func (a *Action) SetMessage(msg string) *Action {
	a.Message = msg
	return a
}

func (a *Action) SetStatus(status bool) *Action {
	a.Status = status
	return a
}

var (
	PingAction        = makeAction(Ping, "", true)
	RejectAction      = makeAction(Reject, "无效的请求, 清检查您的参数", false)
	CloseAction       = makeAction(Close, "服务端关闭", false)
	SuccessRespAction = makeAction(Resp, "success", true)
	NoticeAction      = makeAction(Notice, "", true)
)
