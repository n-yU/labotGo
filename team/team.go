// チーム管理
package team

import (
	"fmt"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType := cmdValues[0]; subType {
	case "add":
		blocks, responseType, ok = getBlockAdd(), Ephemeral, true
	case "edit":

	case "delete":

	case "list":

	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s team *%s* を使用することはできません\n", Cmd, subType))
		blocks, responseType, ok = post.SingleTextBlock(text), Ephemeral, false
	}

	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	switch actionID {
	case aid.AddTeam:
		blocks := AddMember(callback.BlockActionState.Values)
		err = post.PostMessage(callback, blocks, Ephemeral)
	}
	return err
}
