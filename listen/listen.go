// イベント受信処理
package listen

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/member"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/team"
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
	isEmptyValues := (len(cmdValues) == 0)

	// コマンドタイプ別 処理
	switch cmdType {
	case "hello":
		if isEmptyValues {
			text := "*Hello, World!*"
			blocks, responseType, ok = post.SingleTextBlock(text), InChannel, true
		} else {
			text := post.ErrText(fmt.Sprintf("%s *%s* に引数を与えることはできません\n", Cmd, cmdType))
			blocks, responseType = post.SingleTextBlock(text), Ephemeral
		}
	case "member":
		if isEmptyValues {
			text := post.ErrText(fmt.Sprintf("%s *%s* には1つ以上の引数を与える必要があります\n", Cmd, cmdType))
			blocks, responseType = post.SingleTextBlock(text), Ephemeral
		} else {
			blocks, responseType, ok = member.GetBlocks(cmdValues)
		}
	case "team":
		if isEmptyValues {
			text := post.ErrText(fmt.Sprintf("%s *%s* には1つ以上の引数を与える必要があります\n", Cmd, cmdType))
			blocks, responseType = post.SingleTextBlock(text), Ephemeral
		} else {
			blocks, responseType, ok = team.GetBlocks(cmdValues)
		}
	case "shuffle":

	case "group":

	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s *%s* を使用することはできません\n", Cmd, cmdType))
		blocks, responseType = post.SingleTextBlock(text), Ephemeral
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
func BlockAction(callback slack.InteractionCallback) (err error) {
	// アクションID 取得
	if len(callback.ActionCallback.BlockActions) == 0 {
		return err
	}
	actionID := callback.ActionCallback.BlockActions[0].ActionID

	// アクションID別 処理
	switch {
	case strings.HasPrefix(actionID, aid.BaseMember):
		err = member.Action(actionID, callback)
	case strings.HasPrefix(actionID, aid.BaseTeam):
		err = team.Action(actionID, callback)
	default:
		Logger.Printf("不明なアクション %s を受け取りました\n", actionID)
	}

	return err
}
