package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/blivedm-go/client"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/blivedm-go/message"
	_ "github.com/pengfeiXV/BilibiliDanmuRobot-Core/blivedm-go/utils"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/config"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/entity"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/http"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/logic/danmu"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/svc"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/utiles"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

type wsHandler struct {
	client *client.Client
	svc    *svc.ServiceContext
	// 机器人
	robotBulletCtx    context.Context
	robotBulletCancel context.CancelFunc
	// 弹幕发送
	sendBulletCtx    context.Context
	sendBulletCancel context.CancelFunc
	// 特效欢迎
	interactCtx    context.Context
	interactCancel context.CancelFunc
	// 礼物感谢
	thanksGiftCtx   context.Context
	thankGiftCancel context.CancelFunc
	// pk提醒
	pkCtx    context.Context
	pkCancel context.CancelFunc
	// 弹幕处理
	danmuLogicCtx    context.Context
	danmuLogicCancel context.CancelFunc
	// 定时弹幕
	cronDanmu           *cron.Cron
	mapCronDanmuSendIdx map[int]int
	userId              int
	initStart           bool
}

func NewWsHandler() WsHandler {
	ctx, err := mustloadConfig()
	if err != nil {
		return nil
	}
	ws := new(wsHandler)
	ws.initStart = false
	err = ws.starthttp()
	if err != nil {
		logx.Error(err)
		return nil
	}
	ws.client = client.NewClient(ctx.Config.RoomId)
	ws.client.SetCookie(http.CookieStr)
	ws.svc = ctx
	// 初始化定时弹幕
	ws.cronDanmu = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)))
	ws.mapCronDanmuSendIdx = make(map[int]int)

	// 设置uid作为基本配置
	strUserId, ok := http.CookieList["DedeUserID"]
	if !ok {
		logx.Infof("uid加载失败，请重新登录")
		return nil
	}
	ws.userId, err = strconv.Atoi(strUserId)
	ctx.RobotID = strUserId
	roominfo, err := http.RoomInit(ctx.Config.RoomId)
	if err != nil {
		logx.Error(err)
		return nil
	}
	ws.svc.LiveStatus = roominfo.Data.LiveStatus
	ctx.UserID = roominfo.Data.Uid
	return ws
}
func (ws *wsHandler) ReloadConfig() error {
	ctx, err := mustloadConfig()
	oldconfig := *ws.svc.Config
	if err != nil {
		return err
	}
	ws.svc.Config = ctx.Config
	if ctx.Config.RoomId != oldconfig.RoomId {
		logx.Infof("房间号更改，更换房间号 ：%v", ctx.Config.RoomId)
		ws.client.Stop()
		ws.client = client.NewClient(ctx.Config.RoomId)
		ws.client.SetCookie(http.CookieStr)
		roominfo, err := http.RoomInit(ctx.Config.RoomId)
		if err != nil {
			logx.Error(err)
			// return err
		}
		ctx.UserID = roominfo.Data.Uid
		err = ws.client.Start()
		if err != nil {
			return err
		}
		ws.registerHandler()
	}
	if ctx.Config.CronDanmu != oldconfig.CronDanmu || !areSlicesEqual(ctx.Config.CronDanmuList, oldconfig.CronDanmuList) {
		logx.Info("识别到定时弹幕配置发生变化，重新加载")
		for _, i := range ws.cronDanmu.Entries() {
			ws.cronDanmu.Remove(i.ID)
		}
		ws.corndanmuStart()
	}
	return nil
}

type WsHandler interface {
	InitStartWsClient()
	StopWsClient()
	SayGoodbye()
	StopChanel()
	StartWsClient() error
	starthttp() error
	ReloadConfig() error
	GetSvc() svc.ServiceContext
	GetUserinfo() *entity.UserinfoLite
}

func (ws *wsHandler) InitStartWsClient() {
	ws.startLogic()
}
func (ws *wsHandler) StartWsClient() error {
	if ws.svc.Config.EntryMsg != "off" {
		err := http.Send(ws.svc.Config.EntryMsg, ws.svc)
		if err != nil {
			logx.Error(err)
		}
	}
	ws.cronDanmu.Start()
	ws.client = client.NewClient(ws.svc.Config.RoomId)
	ws.client.SetCookie(http.CookieStr)
	ws.registerHandler()

	ws.client.OnLiveStart(func(start *message.LiveStart) {
		ws.svc.LiveStatus = entity.Live
	})
	ws.client.OnLiveStop(func(end *message.LiveStop) {
		ws.svc.LiveStatus = entity.NotStarted
	})

	return ws.client.Start()
}
func (ws *wsHandler) GetUserinfo() *entity.UserinfoLite {
	return http.GetUserInfo()
}
func (ws *wsHandler) GetSvc() svc.ServiceContext {
	return *ws.svc
}
func (ws *wsHandler) StopWsClient() {
	ws.cronDanmu.Stop()
	ws.client.Stop()
	// ws.svc.Db.Db.Close()
}
func (ws *wsHandler) StopChanel() {
	if ws.sendBulletCancel != nil {
		ws.sendBulletCancel()
	}
	if ws.robotBulletCancel != nil {
		ws.robotBulletCancel()
	}
	if ws.thankGiftCancel != nil {
		ws.thankGiftCancel()
	}
	if ws.interactCancel != nil {
		ws.interactCancel() // 关闭弹幕姬goroutine
	}
	if ws.pkCancel != nil {
		ws.pkCancel()
	}
	if ws.danmuLogicCancel != nil {
		ws.danmuLogicCancel()
	}
	for _, i := range ws.cronDanmu.Entries() {
		ws.cronDanmu.Remove(i.ID)
	}
}
func (ws *wsHandler) SayGoodbye() {
	if len(ws.svc.Config.GoodbyeInfo) > 0 {

		var danmuLen = ws.svc.Config.DanmuLen
		var msgdata []string
		msgrun := []rune(ws.svc.Config.GoodbyeInfo)
		msgLen := len(msgrun)
		msgcount := msgLen / danmuLen
		tmpmsgcount := msgLen % danmuLen
		if tmpmsgcount != 0 {
			msgcount += 1
		}
		for m := 1; m <= msgcount; m++ {
			if msgLen < m*danmuLen {
				msgdata = append(msgdata, string(msgrun[(m-1)*danmuLen:msgLen]))
				continue
			}
			msgdata = append(msgdata, string(msgrun[(m-1)*danmuLen:danmuLen*m]))
		}
		for _, msgs := range msgdata {
			err := http.Send(msgs, ws.svc)
			if err != nil {
				logx.Errorf("下播弹幕发送失败：%s msg: %s", err, msgs)
			}
			time.Sleep(1 * time.Second) // 防止弹幕发送过快
			// logx.Info(">>>>>>>>>", msgs)
		}
	}
}
func (ws *wsHandler) startLogic() {
	ws.sendBulletCtx, ws.sendBulletCancel = context.WithCancel(context.Background())
	go logic.StartSendBullet(ws.sendBulletCtx, ws.svc)
	logx.Info("弹幕推送已开启...")
	// 机器人
	ws.robotBulletCtx, ws.robotBulletCancel = context.WithCancel(context.Background())
	go logic.StartBulletRobot(ws.robotBulletCtx, ws.svc)
	// 弹幕逻辑
	ws.danmuLogicCtx, ws.danmuLogicCancel = context.WithCancel(context.Background())
	go danmu.StartDanmuLogic(ws.danmuLogicCtx, ws.svc)

	logx.Info("弹幕机器人已开启")
	// 特效欢迎
	ws.interactCtx, ws.interactCancel = context.WithCancel(context.Background())
	go logic.Interact(ws.interactCtx, ws.svc)

	logx.Info("欢迎模块已开启")

	// 礼物感谢
	ws.thanksGiftCtx, ws.thankGiftCancel = context.WithCancel(context.Background())
	go logic.ThanksGift(ws.thanksGiftCtx, ws.svc)

	logx.Info("礼物感谢已开启")
	// pk提醒
	ws.pkCtx, ws.pkCancel = context.WithCancel(context.Background())
	go logic.PK(ws.pkCtx, ws.svc)

	// 下播提醒
	// ws.sayGoodbyeByWs()

	// 定时弹幕
	ws.corndanmuStart()

	// ws.registerHandler()
}
func (ws *wsHandler) registerHandler() {
	ws.welcomeEntryEffect()
	ws.welcomeInteractWord()
	logx.Info("弹幕处理已开启")
	ws.receiveDanmu()
	// 天选自动关闭欢迎
	ws.anchorLot()
	logx.Info("pk提醒已开启")
	ws.pkBattleStart()
	ws.pkBattleEnd()
	// 禁言用户提醒
	ws.blockUser()
	ws.thankGifts()
	// 红包
	ws.redPocket()
}
func (ws *wsHandler) starthttp() error {
	var err error
	http.InitHttpClient()
	// 判断是否存在历史cookie
	if http.FileExists("token/bili_token.txt") && http.FileExists("token/bili_token.json") {
		err = http.SetHistoryCookie()
		if err != nil {
			logx.Error("用户登录失败")
			return err
		}
		logx.Info("用户登录成功")
	} else {
		// if err = ws.userlogin(); err != nil {
		//	logx.Errorf("用户登录失败：%v", err)
		//	return
		// }
		// logx.Info("用户登录成功")
		logx.Error("用户登录失败")
		return errors.New("用户登录失败")
	}
	return nil
}
func (ws *wsHandler) userlogin() error {
	var err error
	http.InitHttpClient()
	var loginUrl *entity.LoginUrl
	if loginUrl, err = http.GetLoginUrl(); err != nil {
		logx.Error(err)
		return err
	}

	if err = utiles.GenerateQr(loginUrl.Data.Url); err != nil {
		logx.Error(err)
		return err
	}

	if _, err = http.GetLoginInfo(loginUrl.Data.OauthKey); err != nil {
		logx.Error(err)
		return err
	}

	return err
}
func (ws *wsHandler) corndanmuStart() {
	if ws.svc.Config.CronDanmu == false {
		return
	}
	for n, danmux := range ws.svc.Config.CronDanmuList {
		if danmux.Danmu != nil {
			i := n
			danmus := danmux
			_, err := ws.cronDanmu.AddFunc(danmus.Cron, func() {
				roomInfo, err := http.RoomInit(int(ws.svc.UserID))
				if err != nil {
					logx.Error(err)
					return
				}
				ws.svc.LiveStatus = roomInfo.Data.LiveStatus
				if ws.svc.LiveStatus != entity.Live {
					return
				}
				if len(danmus.Danmu) > 0 {
					if danmus.Random {
						logic.PushToBulletSender(danmus.Danmu[rand.Intn(len(danmus.Danmu))])
					} else {
						_, ok := ws.mapCronDanmuSendIdx[i]
						if !ok {
							ws.mapCronDanmuSendIdx[i] = 0
						}
						ws.mapCronDanmuSendIdx[i] = ws.mapCronDanmuSendIdx[i] + 1
						logic.PushToBulletSender(danmus.Danmu[ws.mapCronDanmuSendIdx[i]%len(danmus.Danmu)])
					}
				}
			})
			if err != nil {
				logx.Errorf("第%d条定时弹幕配置出现错误: %v", i+1, err)
			}
		}
	}
	ws.cronDanmu.Start()
}
func mustloadConfig() (*svc.ServiceContext, error) {
	dir := "./token"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory does not exist, create it
		err = os.Mkdir(dir, 0755)
		if err != nil {
			panic(fmt.Sprintf("无法创建token文件夹 请手动创建:%s", err))
		}
	}

	var c config.Config
	conf.MustLoad("etc/bilidanmaku-api.yaml", &c, conf.UseEnv())
	logx.MustSetup(c.Log)
	logx.DisableStat()
	// 配置数据库文件夹
	info, err := os.Stat(c.DBPath)
	if os.IsNotExist(err) || !info.IsDir() {
		err = os.MkdirAll(c.DBPath, 0777)
		if err != nil {
			logx.Errorf("文件夹创建失败：%s", c.DBPath)
			return nil, err
		}
	}
	ctx := svc.NewServiceContext(c)
	return ctx, err
}

// 比较两个 Person 切片是否相同
func areSlicesEqual(a, b []config.CronDanmuList) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}
