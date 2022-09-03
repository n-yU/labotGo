// 機能: メンバーグルーピング
package group

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType, subValues := cmdValues[0], cmdValues[1:]; subType {
	case "team":
		blocks, ok = getBlocksTeam(), true
	case "custom":
		blocks, ok = getBlocksCustom(), true
	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s group *%s* %s を使用することはできません", util.Cmd, subType, strings.Join(subValues, " ")))
		blocks, ok = post.SingleTextBlock(text), false
	}

	responseType = util.Ephemeral
	return blocks, responseType, ok
}

// 指定アクション実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	var (
		blocks       []slack.Block
		responseType string
		actionUserID = callback.User.ID
	)

	switch actionID {
	case aid.GroupTeam:
		blocks, responseType = GroupTeam(actionUserID, callback.BlockActionState.Values)
	case aid.GroupCustom:
		blocks, responseType = GroupCustom(actionUserID, callback.BlockActionState.Values)
	default:
	}

	if len(blocks) > 0 {
		err = post.PostMessage(callback, blocks, responseType)
	}
	return err
}

// グルーピング結果セクション 取得
func getGroupResultSections(groupUserIDs []string, groupNo int, md data.MembersData) (resultSections []*slack.ContextBlock) {
	groupNameElements := []slack.MixedElement{post.TxtBlockObj(util.Markdown, fmt.Sprintf("*グループ%02d*", groupNo))}
	resultSections = append(resultSections, slack.NewContextBlock("", groupNameElements...))

	elements := []slack.MixedElement{}
	for _, userID := range groupUserIDs {
		elements = append(elements, slack.NewImageBlockElement(md[userID].Image24, userID))
		elements = append(elements, post.TxtBlockObj(util.Markdown, fmt.Sprintf("*<@%s>*", userID)))
		// Context の element の上限は10個のためセクション化してリセット
		if len(elements) >= 10 {
			resultSections = append(resultSections, slack.NewContextBlock("", elements...))
			elements = []slack.MixedElement{}
		}
	}

	if len(elements) > 0 {
		resultSections = append(resultSections, slack.NewContextBlock("", elements...))
	}

	return resultSections
}
