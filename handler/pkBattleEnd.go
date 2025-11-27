package handler

import "github.com/pengfeiXV/BilibiliDanmuRobot-Core/svc"

func (ws *wsHandler) pkBattleEnd() {
	ws.client.RegisterCustomEventHandler("PK_BATTLE_END", func(s string) {
		cleanOtherSide(ws.svc)
	})
	ws.client.RegisterCustomEventHandler("PK_END", func(s string) {
		cleanOtherSide(ws.svc)
	})
	ws.client.RegisterCustomEventHandler("PK_BATTLE_CRIT", func(s string) {
		cleanOtherSide(ws.svc)
	})
	ws.client.RegisterCustomEventHandler("PK_BATTLE_SETTLE_NEW", func(s string) {
		cleanOtherSide(ws.svc)
	})
}
func cleanOtherSide(svcCtx *svc.ServiceContext) {
	for k := range svcCtx.OtherSideUid {
		delete(svcCtx.OtherSideUid, k)
	}
}
