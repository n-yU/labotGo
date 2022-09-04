// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型セクション: チーム選択
func SelectTeamsSection(teamNames []string, actionID string, initTeamNames []string, isMulti bool) *slack.SectionBlock {
	var selectOptionType string
	if isMulti {
		selectOptionType = slack.MultiOptTypeStatic
	} else {
		selectOptionType = slack.OptTypeStatic
	}

	options, initOptions := OptionBlockObjectList(teamNames, false), OptionBlockObjectList(initTeamNames, false)
	selectOptionText := TxtBlockObj(util.PlainText, "チームを選択")
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(util.Markdown, "*チーム*")
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 定型セクション: チーム情報
func InfoTeamSections(
	newTeamName, oldTeamName string, profImages map[string]string, newUserIDs, oldUserIDs []string, createdInfo *data.CreatedInfo,
) (infoSections []*slack.ContextBlock) {
	var nameObj, userIDsInfo *slack.TextBlockObject

	nameInfoObj := TxtBlockObj(util.Markdown, "*チーム名*:")
	if oldTeamName == newTeamName {
		nameObj = TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", newTeamName))
	} else {
		nameObj = TxtBlockObj(util.Markdown, fmt.Sprintf("~%s~ → *%s*", oldTeamName, newTeamName))
	}
	elements := []slack.MixedElement{nameInfoObj, nameObj}

	// 詳細版（listコマンド使用時）
	if createdInfo != nil {
		createdImageObj := slack.NewImageBlockElement(createdInfo.Image, createdInfo.UserID)
		createdUserIDObj := TxtBlockObj(util.Markdown, fmt.Sprintf("<@%s>", createdInfo.UserID))
		createdAtObj := TxtBlockObj(util.Markdown, util.FormatTime(createdInfo.At))
		elements = append(elements, []slack.MixedElement{
			TxtBlockObj(util.Markdown, "*作成*"), createdImageObj, createdUserIDObj, createdAtObj,
		}...)
	}
	infoSections = append(infoSections, slack.NewContextBlock("", elements...))

	userIDsInfoObj := TxtBlockObj(util.Markdown, "*メンバー*:")
	elements = []slack.MixedElement{userIDsInfoObj}
	if len(oldUserIDs)+len(newUserIDs) > 0 {
		for _, userID := range util.UniqueConcatSlice(oldUserIDs, newUserIDs) {
			var userIDText string
			if isOld, isNew := util.ListContains(oldUserIDs, userID), util.ListContains(newUserIDs, userID); isOld && isNew {
				userIDText = fmt.Sprintf("<@%s>", userID)
			} else if isOld {
				userIDText = fmt.Sprintf("~<@%s>~", userID)
			} else if isNew {
				userIDText = fmt.Sprintf("*<@%s>*", userID)
			} else {
			}
			elements = append(elements, slack.NewImageBlockElement(profImages[userID], userID))
			elements = append(elements, TxtBlockObj(util.Markdown, userIDText))

			// Context の element の上限は10個のためセクション化してリセット
			if len(elements) >= 9 {
				infoSections = append(infoSections, slack.NewContextBlock("", elements...))
				elements = []slack.MixedElement{TxtBlockObj(util.Markdown, "　　　　 ")}
			}
		}
	} else {
		userIDsInfo = TxtBlockObj(util.Markdown, "所属メンバーなし")
		elements = append(elements, userIDsInfo)
	}

	if len(elements) > 1 {
		infoSections = append(infoSections, slack.NewContextBlock("", elements...))
	}

	return infoSections
}

// 定型セクション: チーム名入力
func InputTeamNameSection(actionID string, initTeamName string) *slack.InputBlock {
	inputSectionText := TxtBlockObj(util.PlainText, "チーム名")
	inputSectionHint := TxtBlockObj(util.PlainText, "1〜20文字で入力してください ／ 半角スペース・半角カンマは使用できません")
	inputText := TxtBlockObj(util.PlainText, "チーム名を入力")
	input := &slack.PlainTextInputBlockElement{
		Type: slack.METPlainTextInput, ActionID: actionID, Placeholder: inputText, InitialValue: initTeamName,
	}
	input.MinLength, input.MaxLength = 1, 20
	inputSection := slack.NewInputBlock("", inputSectionText, inputSectionHint, input)
	return inputSection
}

// 定型ブロック: チームデータエラー
func ErrBlocksTeamsData(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case util.DataLoadErr:
		text = "チームデータの読み込みに失敗しました"
	case util.DataReloadErr:
		text = "チームデータの更新に失敗しました"
	default:
		util.Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := SingleTextSectionBlock(util.PlainText, ErrText(text))
	tipsSection := TipsSection(TipsDataError(util.TeamDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	util.Logger.Println(text)
	util.Logger.Println(err)
	return blocks
}

// 定型ブロック: 未定義チーム指定エラー
func ErrBlocksUnknownTeam(teamName string) []slack.Block {
	text := ErrText(fmt.Sprintf("指定したチーム `%s` は存在しません", teamName))
	textSection := SingleTextSectionBlock(util.Markdown, text)
	tipsText := []string{util.TipsTeamList()}
	tipsSection := TipsSection(tipsText)

	blocks := []slack.Block{textSection, tipsSection}
	return blocks
}
