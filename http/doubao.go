package http

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
)

func RequestDoubaoRobot(msg string, svcCtx *svc.ServiceContext) (string, error) {
	client := arkruntime.NewClientWithApiKey(os.Getenv("DOUBAO_API_KEY"))
	ctx := context.Background()
	req := model.CreateChatCompletionRequest{
		Model: "doubao-seed-1-6-251015",
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(fmt.Sprintf("当前对话用户ID：%s，昵称：%s，身份：%s。请根据用户身份调整回复风格：对主播用管家对待主人的态度表达，对用户用可爱俏皮的风格表达。",
						"", "", "用户")),
				},
			},
			{
				Role: "user",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("鸭鸭是猪"),
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
