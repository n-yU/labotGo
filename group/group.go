// 機能: メンバーグルーピング
package group

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/shuffle"
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

// 定型セクション: グルーピング結果
func GroupResultSections(groupUserIDs []string, groupNo int, md data.MembersData) (resultSections []*slack.ContextBlock) {
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

// 定型セクション: グルーピングタイプ選択
func TypeSelectSection(actionID string) *slack.SectionBlock {
	typeOptions := post.OptionBlockObjectList([]string{util.GroupTypeOptionNum, util.GroupTypeOptionSize}, false)
	typeSelectOptionText := post.TxtBlockObj(util.PlainText, "グルーピングタイプを選択")
	typeSelectOption := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic, typeSelectOptionText, actionID, typeOptions...,
	)
	typeSelectText := post.TxtBlockObj(util.Markdown, "*グルーピングタイプ*")
	typeSelectSection := slack.NewSectionBlock(typeSelectText, nil, slack.NewAccessory(typeSelectOption))
	return typeSelectSection
}

// 定型セクション: グルーピングバリュー入力
func ValueInputSection(actionID string) *slack.InputBlock {
	valueInputSectionText := post.TxtBlockObj(util.PlainText, "グループ数・グループサイズ")
	valueInputSectionHint := post.TxtBlockObj(
		util.PlainText, "指定グループのメンバー数以下の自然数を入力してください\nグループサイズを指定する場合は末尾に +/- を付けてください",
	)
	valueInputText := post.TxtBlockObj(util.PlainText, "グループ数・グループサイズを入力")
	valueInput := slack.NewPlainTextInputBlockElement(valueInputText, actionID)
	valueInputSection := slack.NewInputBlock("", valueInputSectionText, valueInputSectionHint, valueInput)
	return valueInputSection
}

// グルーピングリクエスト欠損値 チェック
func CheckMissingValue(elements []string, groupType string, groupValue string, isCustom bool) (blocks []slack.Block) {
	isEmptyTeamNames := (len(elements) == 0)
	isEmptyGroupType, isEmptyGroupValue := (groupType == ""), (groupValue == "")

	if isEmptyTeamNames || isEmptyGroupType || isEmptyGroupValue {
		emptyElements := []string{}
		if isEmptyTeamNames {
			if isCustom {
				emptyElements = append(emptyElements, "メンバー")
			} else {
				emptyElements = append(emptyElements, "チーム")
			}
		}
		if isEmptyGroupType {
			emptyElements = append(emptyElements, "タイプ")
		}
		if isEmptyGroupValue {
			emptyElements = append(emptyElements, "グループ数・グループサイズ")
		}
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"%s が指定されていません", strings.Join(emptyElements, "／"),
		)))
	}
	return blocks
}

// 定型ブロック: グルーピング結果
func GroupBlocks(
	memberUserIDs []string, groupType string, groupValueInt int, groupValueOption string,
	teamNamesString string, md data.MembersData,
) (blocks []slack.Block, ok bool) {
	var headerText string
	shuffledMemberUIDs := shuffle.ShuffleMemberUserIDs(memberUserIDs)

	switch groupType {
	case util.GroupTypeOptionNum:
		// タイプ: グループ数指定
		if len(teamNamesString) > 0 {
			headerText = post.ScsText(fmt.Sprintf(
				"指定チーム *%s* を *グループ数=%d* でグルーピングしました", teamNamesString, groupValueInt,
			))
		} else {
			headerText = post.ScsText(fmt.Sprintf("指定メンバーを *グループ数=%d* でグルーピングしました", groupValueInt))
		}
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		blocks = []slack.Block{headerSection, util.Divider()}

		// 各メンバー グループ割当
		groupsUserIDs := make([][]string, groupValueInt, groupValueInt)
		for i, userID := range shuffledMemberUIDs {
			groupNo := i % groupValueInt
			groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], userID)
		}

		// グルーピング結果セクション 追加
		for i, groupUserIDs := range groupsUserIDs {
			if len(groupUserIDs) == 0 {
				continue
			}
			groupResultSections := GroupResultSections(groupUserIDs, i+1, md)
			for _, groupInfoSec := range groupResultSections {
				blocks = append(blocks, groupInfoSec)
			}
			blocks = append(blocks, util.Divider())
		}
		ok = true
	case util.GroupTypeOptionSize:
		// タイプ: グループサイズ指定
		if len(teamNamesString) > 0 {
			headerText = post.ScsText(fmt.Sprintf(
				"指定チーム *%s* を *グループサイズ=%d* でグルーピングしました", teamNamesString, groupValueInt,
			))
		} else {
			headerText = post.ScsText(fmt.Sprintf("指定メンバーを *グループサイズ=%d* でグルーピングしました", groupValueInt))
		}
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		blocks = []slack.Block{headerSection, util.Divider()}

		// 各メンバー グループ割当
		groupNum := len(shuffledMemberUIDs) / groupValueInt
		groupsUserIDs := make([][]string, groupNum, groupNum+1)
		for i, userID := range shuffledMemberUIDs[:(groupNum * groupValueInt)] {
			groupNo := i % groupNum
			groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], userID)
		}

		// 残余メンバー グループ割当
		if groupValueOption == "+" {
			// チーム数を維持して，残余メンバーを各チームに割当
			for i := groupNum * groupValueInt; i < len(shuffledMemberUIDs); i++ {
				groupNo := i % groupNum
				groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], shuffledMemberUIDs[i])
			}
		} else if groupValueOption == "-" {
			// チームを1つ増やし，そのチームに残余メンバーを全員割当
			groupsUserIDs = append(groupsUserIDs, shuffledMemberUIDs[(groupNum*groupValueInt):])
		} else {
		}

		// グルーピング結果セクション 追加
		for i, groupUserIDs := range groupsUserIDs {
			if len(groupUserIDs) == 0 {
				continue
			}
			groupResultSections := GroupResultSections(groupUserIDs, i+1, md)
			for _, groupInfoSec := range groupResultSections {
				blocks = append(blocks, groupInfoSec)
			}
			blocks = append(blocks, util.Divider())
		}
		ok = true
	default:
	}

	return blocks, ok
}
