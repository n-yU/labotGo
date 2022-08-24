// イベント受信処理
package listen

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/member"
	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンドテキスト分割
func splitCmd(cmdText string) (string, []string) {
	cmdValues := strings.Split(cmdText, " ")
	cmdType := cmdValues[0]

	if len(cmdValues) > 1 {
		cmdValues = cmdValues[1:]
	} else {
		cmdValues = []string{}
	}

	return cmdType, cmdValues
}

// コマンド 受信処理
func Command(cmd slack.SlashCommand) error {
	var (
		blocks       []slack.Block
		responseType string
		ok           = false
	)

	// コマンド受信
	userId, cmdText := cmd.UserID, cmd.Text
	Logger.Printf("受信コマンド (from:%s): %s\n", userId, cmdText)
	cmdType, cmdValues := splitCmd(cmdText)

	// コマンドタイプ別 処理
	switch cmdType {
	case "hello":
		if len(cmdValues) == 0 {
			text := "*Hello, World!*"
			blocks, responseType, ok = post.CreateSingleTextBlock(text), InChannel, true
		} else {
			text := fmt.Sprintf("%s *hello* に引数を与えることはできません\n", Cmd)
			blocks, responseType = post.CreateSingleTextBlock(text), Ephemeral
		}
	case "member":
		if len(cmdValues) == 0 {
			text := fmt.Sprintf("%s *member* には1つ以上の引数を与える必要があります\n", cmd)
			blocks, responseType = post.CreateSingleTextBlock(text), Ephemeral
		} else {
			blocks, responseType, ok = member.GetBlocks(cmdValues)
		}
	default:
		text := fmt.Sprintf("コマンド %s *%s* を使用することはできません\n", Cmd, cmdType)
		blocks, responseType = post.CreateSingleTextBlock(text), Ephemeral
	}

	// コマンド処理 成功有無通知
	if ok {
		Logger.Printf("SUCCESSFUL command: %s\n", cmdText)
	} else {
		Logger.Printf("FAILED command: %s\n", cmdText)
	}

	// メッセージ投稿
	err := post.PostMessage(cmd, blocks, responseType)
	return err
}

// ブロックアクション 受信処理
func BlockAction(callback slack.InteractionCallback) error {
	var err error

	// アクションID 取得
	if len(callback.ActionCallback.BlockActions) == 0 {
		return err
	}
	actionId := callback.ActionCallback.BlockActions[0].ActionID

	// アクションID別 処理
	switch {
	case strings.HasPrefix(actionId, "member"):
		err = member.Action(actionId, callback)
	default:
		Logger.Printf("不明なアクション %s を受け取りました\n", actionId)
	}

	return err
}
