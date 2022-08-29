// チーム管理
package team

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
		blocks, ok = getBlockEditTeamSelect(), true
	case "delete":
		blocks, ok = getBlockDeleteTeamSelect(), true
	case "list":
		blocks, ok = getBlockListTeam(subValues), true
	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s team *%s* %s を使用することはできません", util.Cmd, subType, strings.Join(subValues, " ")))
		blocks, ok = post.SingleTextBlock(text), false
	}

	responseType = util.Ephemeral
	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	var (
		blocks       = []slack.Block{}
		actionUserID = callback.User.ID
	)

	switch {
	case actionID == aid.AddTeam:
		blocks = AddMember(actionUserID, callback.BlockActionState.Values)
	case actionID == aid.EditTeamSelectName:
		blocks = getBlockEditTeamInfo(actionUserID, callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.EditTeam+"_"):
		teamName := strings.Split(actionID, "_")[1]
		blocks = EditTeam(actionUserID, callback.BlockActionState.Values, teamName)
	case actionID == aid.DeleteTeamSelectTeam:
		blocks = DeleteTeamConfirm(actionUserID, callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.DeleteTeam+"_"):
		teamName := strings.Split(actionID, "_")[1]
		blocks = DeleteTeam(actionUserID, callback.BlockActionState.Values, teamName)
	}

	if len(blocks) > 0 {
		err = post.PostMessage(callback, blocks, util.Ephemeral)
	}
	return err
}
