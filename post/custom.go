// メッセージ投稿
package post

import (
	"fmt"
	"strings"

	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 頻用テキスト: エラー
func ErrText(text string) string {
	return fmt.Sprintf(":x: %s", text)
}

// 頻用テキスト: 成功
func ScsText(text string) string {
	return fmt.Sprintf(":white_check_mark: %s", text)
}

// 頻用テキスト: 情報
func InfoText(text string) string {
	return fmt.Sprintf(":diamond_shape_with_a_dot_inside: %s", text)
}

// 頻用テキスト: データファイルエラー Tips
func TipsDataError(dataPath string) []string {
	return []string{
		fmt.Sprintf("データファイル `%s` が存在しないか，ファイル／データ形式が不適切です", dataPath),
		"データファイルを削除した上で，botを再起動すると解消されます（但しデータはリセットされます）",
	}
}

// 頻用ブロック要素: スタイル指定有ボタン
func NewButtonBlockElementWithStyle(actionID, value string, text *slack.TextBlockObject, style slack.Style) *slack.ButtonBlockElement {
	// ref.) https://github.com/slack-go/slack/blob/03f86be11aa50ac65d66f3917e250d3257389028/block_element.go#L176
	return &slack.ButtonBlockElement{
		Type:     slack.METButton,
		ActionID: actionID,
		Text:     text,
		Value:    value,
		Style:    style,
	}
}

// 頻用ブロックオブジェクト: slack.NewTextBlockObject()
func TxtBlockObj(elementType string, text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(elementType, text, false, false)
}

// 頻用ブロックオブジェクト: Option オブジェクトリスト
func OptionBlockObjectList(options []string, isUser bool) []*slack.OptionBlockObject {
	// ref.) https://github.com/slack-go/slack/blob/03f86be11aa50ac65d66f3917e250d3257389028/examples/modal_users/users.go#L92
	optionBlockObjects := []*slack.OptionBlockObject{}

	for _, opt := range options {
		var text string
		if isUser {
			text = fmt.Sprintf("<@%s>", opt)
		} else {
			text = opt
		}
		optionText := slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(text, optionText, nil))
	}

	return optionBlockObjects
}

// 頻用セクション: slack.NewSectionBlock(TxtBlockObj(), nil, nil)
func SingleTextSectionBlock(elementType string, text string) *slack.SectionBlock {
	return slack.NewSectionBlock(TxtBlockObj(elementType, text), nil, nil)
}

// 頻用セクション: Tips
func TipsSection(tips []string) *slack.ContextBlock {
	elements := make([]slack.MixedElement, 0, len(tips))
	for _, str := range tips {
		textBlockObject := slack.NewTextBlockObject(Markdown, fmt.Sprintf(":bulb: %s", str), false, false)
		elements = append(elements, textBlockObject)
	}

	tipsSection := slack.NewContextBlock("", elements...)
	return tipsSection
}

// 頻用セクション: OKボタン
func BtnOK(text string, actionID string) *slack.ActionBlock {
	btnText := TxtBlockObj(PlainText, text)
	btn := NewButtonBlockElementWithStyle(actionID, "", btnText, slack.StylePrimary)
	btnBlock := slack.NewActionBlock("", btn)
	return btnBlock
}

// 頻用セクション: メンバー選択
func SelectMembersSection(userIDs []string, actionID string, initUserIDs []string, isMulti, isMember bool) *slack.SectionBlock {
	var selectOptionType, text string
	if isMulti {
		selectOptionType = slack.MultiOptTypeStatic
	} else {
		selectOptionType = slack.OptTypeStatic
	}
	if isMember {
		text = "メンバー"
	} else {
		text = "ユーザ"
	}

	options, initOptions := OptionBlockObjectList(userIDs, true), OptionBlockObjectList(initUserIDs, true)
	selectOptionText := TxtBlockObj(PlainText, fmt.Sprintf("%sを選択", text))
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(Markdown, fmt.Sprintf("*%s*", text))
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 頻用セクション: チーム選択
func SelectTeamsSection(teamNames []string, actionID string, initTeamNames []string, isMulti bool) *slack.SectionBlock {
	var selectOptionType string
	if isMulti {
		selectOptionType = slack.MultiOptTypeStatic
	} else {
		selectOptionType = slack.OptTypeStatic
	}

	options, initOptions := OptionBlockObjectList(teamNames, false), OptionBlockObjectList(initTeamNames, false)
	selectOptionText := TxtBlockObj(PlainText, "チームを選択")
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(Markdown, "*チーム*")
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 頻用セクション: メンバー情報
func InfoMemberSection(userID string, newTeamNames, oldTeamNames []string) *slack.SectionBlock {
	var infoTeams *slack.TextBlockObject
	infoUserID := TxtBlockObj(Markdown, fmt.Sprintf("*ユーザ*:\n<@%s>", userID))

	infoTeamsTextList := []string{}
	for _, teamName := range UniqueConcatSlice(oldTeamNames, newTeamNames) {
		var teamNameText string
		if isOld, isNew := ListContains(oldTeamNames, teamName), ListContains(newTeamNames, teamName); isOld && isNew {
			teamNameText = teamName
		} else if isOld {
			teamNameText = fmt.Sprintf("~%s~", teamName)
		} else if isNew {
			teamNameText = fmt.Sprintf("*%s*", teamName)
		}
		infoTeamsTextList = append(infoTeamsTextList, teamNameText)
	}
	infoTeams = TxtBlockObj(Markdown, fmt.Sprintf("*チーム*:\n%s", strings.Join(infoTeamsTextList, ", ")))

	infoField := []*slack.TextBlockObject{infoUserID, infoTeams}
	infoSection := slack.NewSectionBlock(nil, infoField, nil)
	return infoSection
}

// 頻用セクション: チーム情報
func InfoTeamSection(newTeamName, oldTeamName string, newUserIDs, oldUserIDs []string) *slack.SectionBlock {
	var infoName, infoMembers *slack.TextBlockObject

	if oldTeamName == newTeamName {
		infoName = TxtBlockObj(Markdown, fmt.Sprintf("*チーム名*:\n%s", newTeamName))
	} else {
		infoName = TxtBlockObj(Markdown, fmt.Sprintf("*チーム名*:\n~%s~ → *%s*", oldTeamName, newTeamName))
	}
	if len(oldUserIDs)+len(newUserIDs) > 0 {
		infoMembersTextList := []string{}
		for _, userID := range UniqueConcatSlice(oldUserIDs, newUserIDs) {
			var userIDText string
			if isOld, isNew := ListContains(oldUserIDs, userID), ListContains(newUserIDs, userID); isOld && isNew {
				userIDText = fmt.Sprintf("<@%s>", userID)
			} else if isOld {
				userIDText = fmt.Sprintf("~<@%s>~", userID)
			} else if isNew {
				userIDText = fmt.Sprintf("*<@%s>*", userID)
			} else {
			}
			infoMembersTextList = append(infoMembersTextList, userIDText)
		}
		infoMembers = TxtBlockObj(Markdown, fmt.Sprintf("*メンバー*:\n%s", strings.Join(infoMembersTextList, ", ")))
	} else {
		infoMembers = TxtBlockObj(Markdown, "*メンバー*:\n所属メンバーなし")
	}

	infoField := []*slack.TextBlockObject{infoName, infoMembers}
	infoSection := slack.NewSectionBlock(nil, infoField, nil)
	return infoSection
}

// 頻用セクション: チーム名入力
func InputTeamNameSection(actionID string, initTeamName string) *slack.InputBlock {
	inputSectionText := TxtBlockObj(PlainText, "チーム名")
	inputSectionHint := TxtBlockObj(PlainText, "1〜20文字で入力してください ／ スペースは使用できません")
	inputText := TxtBlockObj(PlainText, "チーム名を入力")
	input := &slack.PlainTextInputBlockElement{
		Type: slack.METPlainTextInput, ActionID: actionID, Placeholder: inputText, InitialValue: initTeamName,
	}
	input.MinLength, input.MaxLength = 1, 20
	inputSection := slack.NewInputBlock("", inputSectionText, inputSectionHint, input)
	return inputSection
}

// 頻用ブロック: シングルテキスト
func SingleTextBlock(text string) []slack.Block {
	blocks := []slack.Block{SingleTextSectionBlock(Markdown, text)}
	return blocks
}
