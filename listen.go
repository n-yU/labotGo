package main

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

func getPayload(blocks []slack.Block, response_type string) map[string]interface{} {
	// payload 取得
	payload := map[string]interface{}{
		"blocks": blocks, "response_type": response_type,
	}
	return payload
}

func splitCmd(cmd_text string) (string, []string) {
	// コマンドテキスト分割
	var cmd_values []string

	cmd_list := strings.SplitN(cmd_text, " ", 2)
	cmd_type := cmd_list[0]
	if len(cmd_list) > 1 {
		cmd_values = strings.Split(cmd_list[1], " ")
	}

	return cmd_type, cmd_values
}

func getSingleTextBlock(text string) []slack.Block {
	// シングルテキストブロック 取得
	blocks := []slack.Block{slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", text, false, false), nil, nil,
	)}
	return blocks
}

func listen(cmd slack.SlashCommand) map[string]interface{} {
	// コマンド受信 & メッセージ送信
	user_name, cmd_text := cmd.UserName, cmd.Text
	logger.Printf("受信コマンド (from:%s): %s\n", user_name, cmd_text)

	// コマンドタイプ・値 格納
	cmd_type, cmd_values := splitCmd(cmd_text)

	var (
		payload map[string]interface{}
		ok      = false
	)

	switch cmd_type {
	case "hello":
		if len(cmd_values) == 0 {
			blocks := getSingleTextBlock("*Hello, World!*")
			payload, ok = getPayload(blocks, slack.ResponseTypeInChannel), true
		} else {
			blocks := getSingleTextBlock("/lg *hello* に引数を与えることはできません")
			payload, ok = getPayload(blocks, slack.ResponseTypeEphemeral), false
		}
	default:
		blocks := getSingleTextBlock(fmt.Sprintf("コマンド /lg *%s* を使用することはできません\n", cmd_type))
		payload, ok = getPayload(blocks, slack.ResponseTypeEphemeral), false
	}

	// コマンド処理成功有無 通知
	if ok {
		logger.Printf("SUCCESSFUL command: %s\n", cmd_text)
	} else {
		logger.Printf("FAILED command: %s\n", cmd_text)
	}

	client.Debugf("%+v\n", payload)
	return payload
}
