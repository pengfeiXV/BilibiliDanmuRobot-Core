package handler

import (
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic"
)

// 下播输出
func (ws *wsHandler) sayGoodbyeByWs() {
	// 下播输出
	ws.client.RegisterCustomEventHandler("PREPARING", func(s string) {
		if len(ws.svc.Config.GoodbyeInfo) > 0 {
			logic.PushToBulletSender(ws.svc.Config.GoodbyeInfo)
		}
	})
}
