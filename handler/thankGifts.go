package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/entity"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic/danmu"
)

// 礼物感谢
func (ws *wsHandler) thankGifts() {
	ws.client.RegisterCustomEventHandler("SEND_GIFT", func(s string) {
		send := &entity.SendGiftText{}
		_ = json.Unmarshal([]byte(s), send)
		if ws.svc.Config.ThanksGift {
			logic.PushToGiftChan(send)
		}
		danmu.SaveBlindBoxStat(send, ws.svc)
	})
	ws.client.RegisterCustomEventHandler("GUARD_BUY", func(s string) {
		if ws.svc.Config.ThanksGift {
			send := &entity.GuardBuyText{}
			_ = json.Unmarshal([]byte(s), send)
			if ws.svc.Config.ThanksGiftUseAt {
				logic.PushToGuardChan(send, &entity.DanmuMsgTextReplyInfo{
					ReplyUid: strconv.Itoa(send.Data.Uid),
				})
			} else {
				logic.PushToGuardChan(send)
			}
		}
	})

	ws.client.RegisterCustomEventHandler("COMMON_NOTICE_DANMAKU", func(s string) {
		if ws.svc.Config.ThanksGift {
			data := &entity.CommonNoticeDanmaku{}
			_ = json.Unmarshal([]byte(s), data)
			if len(data.Data.ContentSegments) == 5 &&
				data.Data.ContentSegments[1].Text == "投喂" &&
				data.Data.ContentSegments[2].Text == "大航海盲盒" {

				logic.PushToBulletSender(fmt.Sprintf("感谢 %s 的 %s", data.Data.ContentSegments[0].Text, data.Data.ContentSegments[4].Text))
			} else if len(data.Data.ContentSegments) == 6 &&
				data.Data.ContentSegments[2].Text == "投喂" &&
				data.Data.ContentSegments[3].Text == "大航海盲盒" {

				logic.PushToBulletSender(fmt.Sprintf("感谢 %s 的 %s", data.Data.ContentSegments[1].Text, data.Data.ContentSegments[5].Text))
			}
		}
	})
}
