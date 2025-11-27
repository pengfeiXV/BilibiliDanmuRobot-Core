package handler

import (
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic/danmu"
	"sync"
)

// 弹幕处理
func (ws *wsHandler) receiveDanmu() {
	// 弹幕处理的功能类接口
	locked := new(sync.Mutex)
	ws.client.RegisterCustomEventHandler("DANMU_MSG", func(s string) {
		locked.Lock()
		danmu.PushToBDanmuLogic(s)
		locked.Unlock()
	})
}
