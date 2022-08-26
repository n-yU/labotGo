// メンバー管理
package member

import (
	"fmt"
	"strings"

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
		blocks, responseType, ok = getBlockEditMemberSelect(), Ephemeral, true
	case "delete":

	case "list":

	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s member *%s* を使用することはできません\n", Cmd, subType))
		blocks, responseType, ok = post.SingleTextBlock(text), Ephemeral, false
	}

	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	switch {
	case actionID == aid.AddMember:
		blocks := AddMember(callback.BlockActionState.Values)
		err = post.PostMessage(callback, blocks, Ephemeral)
	case actionID == aid.EditMemberSelectMember:
		blocks := getBlockEditTeamsSelect(callback.BlockActionState.Values)
		err = post.PostMessage(callback, blocks, Ephemeral)
	case strings.HasPrefix(actionID, aid.EditMember+"_"):
		userID := strings.Split(actionID, "_")[1]
		blocks := EditMember(callback.BlockActionState.Values, userID)
		err = post.PostMessage(callback, blocks, Ephemeral)
	}
	return err
}
