// メッセージ投稿
package post

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/util"
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
		textBlockObject := slack.NewTextBlockObject(util.Markdown, fmt.Sprintf(":bulb: %s", str), false, false)
		elements = append(elements, textBlockObject)
	}

	tipsSection := slack.NewContextBlock("", elements...)
	return tipsSection
}

// 頻用セクション: OKボタン
func BtnOK(text string, actionID string) *slack.ActionBlock {
	btnText := TxtBlockObj(util.PlainText, text)
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
	selectOptionText := TxtBlockObj(util.PlainText, fmt.Sprintf("%sを選択", text))
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", text))
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
	selectOptionText := TxtBlockObj(util.PlainText, "チームを選択")
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(util.Markdown, "*チーム*")
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 頻用セクション: メンバー情報
func InfoMemberSection(profImage, userID string, newTeamNames, oldTeamNames []string) *slack.ContextBlock {
	var teamNamesField *slack.TextBlockObject

	userInfoObject := TxtBlockObj(util.Markdown, "*ユーザ*:")
	profImageObject := slack.NewImageBlockElement(profImage, userID)
	userIDObject := TxtBlockObj(util.Markdown, fmt.Sprintf("<@%s>", userID))

	infoTeamsTextList := []string{}
	for _, teamName := range util.UniqueConcatSlice(oldTeamNames, newTeamNames) {
		var teamNameText string
		if isOld, isNew := util.ListContains(oldTeamNames, teamName), util.ListContains(newTeamNames, teamName); isOld && isNew {
			teamNameText = teamName
		} else if isOld {
			teamNameText = fmt.Sprintf("~%s~", teamName)
		} else if isNew {
			teamNameText = fmt.Sprintf("*%s*", teamName)
		}
		infoTeamsTextList = append(infoTeamsTextList, teamNameText)
	}
	teamInfoObject := TxtBlockObj(util.Markdown, "*チーム*:")
	teamNamesField = TxtBlockObj(util.Markdown, fmt.Sprintf("%s", strings.Join(infoTeamsTextList, ", ")))

	elements := []slack.MixedElement{
		userInfoObject, profImageObject, userIDObject, teamInfoObject, teamNamesField,
	}
	infoSection := slack.NewContextBlock("", elements...)

	return infoSection
}

// 頻用セクション: チーム情報
func InfoTeamSections(newTeamName, oldTeamName string, profImages map[string]string, newUserIDs, oldUserIDs []string) (infoSections []*slack.ContextBlock) {
	var nameObj, userIDsInfo *slack.TextBlockObject

	nameInfoObj := TxtBlockObj(util.Markdown, "*チーム名*:")
	if oldTeamName == newTeamName {
		nameObj = TxtBlockObj(util.Markdown, fmt.Sprintf("%s", newTeamName))
	} else {
		nameObj = TxtBlockObj(util.Markdown, fmt.Sprintf("~%s~ → *%s*", oldTeamName, newTeamName))
	}
	infoSections = append(infoSections, slack.NewContextBlock("", []slack.MixedElement{nameInfoObj, nameObj}...))

	userIDsInfoObj := TxtBlockObj(util.Markdown, "*メンバー*:")
	elements := []slack.MixedElement{userIDsInfoObj}
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

// 頻用セクション: チーム名入力
func InputTeamNameSection(actionID string, initTeamName string) *slack.InputBlock {
	inputSectionText := TxtBlockObj(util.PlainText, "チーム名")
	inputSectionHint := TxtBlockObj(util.PlainText, "1〜20文字で入力してください ／ スペースは使用できません")
	inputText := TxtBlockObj(util.PlainText, "チーム名を入力")
	input := &slack.PlainTextInputBlockElement{
		Type: slack.METPlainTextInput, ActionID: actionID, Placeholder: inputText, InitialValue: initTeamName,
	}
	input.MinLength, input.MaxLength = 1, 20
	inputSection := slack.NewInputBlock("", inputSectionText, inputSectionHint, input)
	return inputSection
}

// 頻用ブロック: シングルテキスト
func SingleTextBlock(text string) []slack.Block {
	blocks := []slack.Block{SingleTextSectionBlock(util.Markdown, text)}
	return blocks
}
