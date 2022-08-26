// メッセージ投稿
package post

import (
	"fmt"
	"strings"

	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// スタイル指定有 ボタンブロック要素
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

// Option ブロックオブジェクトリスト 作成
func CreateOptionBlockObject(options []string, isUser bool) []*slack.OptionBlockObject {
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
		fmt.Sprintf("データファイル `%s` が存在しないか，ファイル／データ形式が不適切です\n", dataPath),
		"データファイルを削除した上で，botを再起動すると解消されます（但しデータはリセットされます）",
	}
}

// 頻用構文: slack.NewTextBlockObject()
func TxtBlockObj(elementType string, text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(elementType, text, false, false)
}

// 頻用構文: slack.NewSectionBlock(TxtBlockObj(), nil, nil)
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
func SelectMembersSection(members []string, actionID string, initMembers []string) *slack.SectionBlock {
	options, initOptions := CreateOptionBlockObject(members, true), CreateOptionBlockObject(initMembers, true)
	selectOptionText := TxtBlockObj(PlainText, "メンバーを選択")
	selectOption := &slack.MultiSelectBlockElement{
		Type: slack.MultiOptTypeStatic, Placeholder: selectOptionText, ActionID: actionID,
		Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(Markdown, "*メンバー*")
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 頻用セクション: チーム選択
func SelectTeamsSection(teams []string, actionID string, initTeams []string) *slack.SectionBlock {
	options, initOptions := CreateOptionBlockObject(teams, false), CreateOptionBlockObject(initTeams, false)
	selectOptionText := TxtBlockObj(PlainText, "チームを選択")
	selectOption := &slack.MultiSelectBlockElement{
		Type: slack.MultiOptTypeStatic, Placeholder: selectOptionText, ActionID: actionID,
		Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(Markdown, "*チーム*")
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 頻用セクション: メンバー情報
func InfoMemberSection(userID string, teams []string) *slack.SectionBlock {
	infoUserId := TxtBlockObj(Markdown, fmt.Sprintf("*ユーザ*:\n<@%s>", userID))
	infoTeams := TxtBlockObj(Markdown, fmt.Sprintf("*チーム*:\n%s", strings.Join(teams, ", ")))
	infoField := []*slack.TextBlockObject{infoUserId, infoTeams}
	infoSection := slack.NewSectionBlock(nil, infoField, nil)
	return infoSection
}

// 頻用ブロック: シングルテキスト
func SingleTextBlock(text string) []slack.Block {
	blocks := []slack.Block{SingleTextSectionBlock(Markdown, text)}
	return blocks
}
