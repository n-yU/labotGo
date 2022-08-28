// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType, subValues := cmdValues[0], cmdValues[1:]; subType {
	case "add":
		blocks, ok = getBlockAdd(), true
	case "edit":
		blocks, ok = getBlockEditMemberSelect(), true
	case "delete":
		blocks, ok = getBlockDeleteMemberSelect(), true
	case "list":
		blocks, ok = getBlockListMember(), true
	default:
		text := post.ErrText(fmt.Sprintf(
			"コマンド %s member *%s* %s を使用することはできません", util.Cmd, subType, strings.Join(subValues, " ")),
		)
		blocks, ok = post.SingleTextBlock(text), false
	}

	responseType = util.Ephemeral
	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	var blocks []slack.Block
	switch {
	case actionID == aid.AddMember:
		blocks = AddMember(callback.BlockActionState.Values)
	case actionID == aid.EditMemberSelectMember:
		blocks = getBlockEditTeamsSelect(callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.EditMember+"_"):
		userID := strings.Split(actionID, "_")[1]
		blocks = EditMember(callback.BlockActionState.Values, userID)
	case actionID == aid.DeleteMemberSelectMember:
		blocks = DeleteMemberConfirm(callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.DeleteMember+"_"):
		userID := strings.Split(actionID, "_")[1]
		blocks = DeleteMember(callback.BlockActionState.Values, userID)
	}

	if len(blocks) > 0 {
		err = post.PostMessage(callback, blocks, util.Ephemeral)
	}
	return err
}
