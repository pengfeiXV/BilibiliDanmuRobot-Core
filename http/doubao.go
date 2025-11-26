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
)

func RequestDoubaoRobot(danmuV2 entity.DanMuV2, svcCtx *svc.ServiceContext) (string, error) {
	var options []arkruntime.ConfigOption
	if svcCtx.Config.Doubao.APIUrl != "" {
		options = append(options, arkruntime.WithBaseUrl(svcCtx.Config.Doubao.APIUrl))
	}
	client := arkruntime.NewClientWithApiKey(svcCtx.Config.Doubao.APIToken, options...)
	ctx := context.Background()
	b := strings.Builder{}
	b.WriteString(svcCtx.Config.Doubao.Prompt)
	b.WriteString(fmt.Sprintf("当前对话用户ID：%s，昵称：%s，身份：%s。请根据用户身份调整回复风格：对主播用管家对待主人的态度表达，对用户用可爱俏皮的风格表达。",
		danmuV2.UserID, danmuV2.Username, danmuV2.Role))
	b.WriteString(fmt.Sprintf("主播的昵称是: %s。", svcCtx.Config.UpNickname))
	b.WriteString(fmt.Sprintf("你的昵称是: %s。", svcCtx.Config.TalkRobotCmd))
	if svcCtx.Config.Doubao.Limit {
		b.WriteString(fmt.Sprintf("请在%v个字内回答。", svcCtx.Config.DanmuLen))
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
		MaxTokens:   volcengine.Int(40),
		ServiceTier: volcengine.String("auto"),
		Thinking: &model.Thinking{
			Type: "disabled",
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return *resp.Choices[0].Message.Content.StringValue, nil
}
