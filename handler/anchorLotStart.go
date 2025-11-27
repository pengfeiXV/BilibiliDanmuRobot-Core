package handler

import "github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic"

// 天选
func (ws *wsHandler) anchorLot() {
	// 天选启动
	ws.client.RegisterCustomEventHandler("ANCHOR_LOT_START", func(s string) {
		if ws.svc.Config.InteractWord || ws.svc.Config.EntryEffect || ws.svc.Config.WelcomeHighWealthy {
			ws.svc.Config.InteractWord = false
			ws.svc.Config.EntryEffect = false
			ws.svc.Config.WelcomeHighWealthy = false
		}
		logic.PushToBulletSender("识别到天选，欢迎弹幕已临时关闭")
	})
	// 天选中奖
	ws.client.RegisterCustomEventHandler("ANCHOR_LOT_AWARD", func(s string) {
		if ws.svc.Config.InteractWord != ws.svc.AutoInteract.InteractWord {
			ws.svc.Config.InteractWord = ws.svc.AutoInteract.InteractWord
		}
		if ws.svc.Config.EntryEffect != ws.svc.AutoInteract.EntryEffect {
			ws.svc.Config.EntryEffect = ws.svc.AutoInteract.EntryEffect
		}
		if ws.svc.Config.WelcomeHighWealthy != ws.svc.AutoInteract.WelcomeHighWealthy {
			ws.svc.Config.WelcomeHighWealthy = ws.svc.AutoInteract.WelcomeHighWealthy
		}
		logic.PushToBulletSender("天选结束，欢迎弹幕已恢复默认")
	})
}
