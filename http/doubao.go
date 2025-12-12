package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/entity"
	"github.com/pengfeiXV/BilibiliDanmuRobot-Core/svc"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/zeromicro/go-zero/core/logx"
)

func RequestDoubaoRobot(danmuV2 entity.DanMuV2, svcCtx *svc.ServiceContext) (string, error) {
	var options []arkruntime.ConfigOption
	if svcCtx.Config.Doubao.APIUrl != "" {
		options = append(options, arkruntime.WithBaseUrl(svcCtx.Config.Doubao.APIUrl))
	}
	client := arkruntime.NewClientWithApiKey(svcCtx.Config.Doubao.APIToken, options...)
	ctx := context.Background()

	// 以system身份发送的预设，可以自定义，每句话都要以句号结尾，防止与下一句话粘连。
	b := strings.Builder{}
	b.WriteString(svcCtx.Config.Doubao.Prompt)
	b.WriteString(fmt.Sprintf("当前对话用户ID：%s，昵称：%s，身份：%s。",
		danmuV2.UserID, danmuV2.Username, danmuV2.Role))
	b.WriteString(fmt.Sprintf("主播的昵称是: %s。", svcCtx.Config.UpNickname))
	b.WriteString(fmt.Sprintf("你的昵称是: %s。", svcCtx.Config.TalkRobotCmd))
	if svcCtx.Config.Doubao.Limit {
		b.WriteString(fmt.Sprintf("回答不多于%v个字符。", svcCtx.Config.DanmuLen))
	}
	req := model.CreateChatCompletionRequest{
		Model: "doubao-seed-1-6-251015",
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(b.String()),
				},
			},
			{
				Role: "user",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(danmuV2.Content),
				},
			},
		},
		MaxTokens: volcengine.Int(100),
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		logx.Errorf("请求doubao失败：%v", err)
		return "", err
	}
	logx.Infof("请求doubao成功：%v", resp)
	return *resp.Choices[0].Message.Content.StringValue, nil
}
