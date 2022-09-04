// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型テキスト: エラー
func ErrText(text string) string {
	return fmt.Sprintf(":x: %s", text)
}

// 定型テキスト: 成功
func ScsText(text string) string {
	return fmt.Sprintf(":white_check_mark: %s", text)
}

// 定型テキスト: 情報
func InfoText(text string) string {
	return fmt.Sprintf(":diamond_shape_with_a_dot_inside: %s", text)
}

// 定型テキスト: データファイルエラー Tips
func TipsDataError(dataPath string) []string {
	return []string{
		fmt.Sprintf("データファイル `%s` が存在しないか，ファイル／データ形式が不適切です", dataPath),
		"データファイルを削除した上で，botを再起動すると解消されます（但しデータはリセットされます）",
	}
}

// 定型ブロック要素: スタイル指定有ボタン
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

// 定型ブロックオブジェクト: slack.NewTextBlockObject()
func TxtBlockObj(elementType string, text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(elementType, text, false, false)
}

// 定型ブロックオブジェクト: Option オブジェクトリスト
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

// 定型セクション: slack.NewSectionBlock(TxtBlockObj(), nil, nil)
func SingleTextSectionBlock(elementType string, text string) *slack.SectionBlock {
	return slack.NewSectionBlock(TxtBlockObj(elementType, text), nil, nil)
}

// 定型セクション: Tips
func TipsSection(tips []string) *slack.ContextBlock {
	elements := make([]slack.MixedElement, 0, len(tips))
	for _, str := range tips {
		textBlockObject := slack.NewTextBlockObject(util.Markdown, fmt.Sprintf(":bulb: %s", str), false, false)
		elements = append(elements, textBlockObject)
	}

	tipsSection := slack.NewContextBlock("", elements...)
	return tipsSection
}

// 定型セクション: OKボタン
func BtnOK(text string, actionID string) *slack.ActionBlock {
	btnText := TxtBlockObj(util.PlainText, text)
	btn := NewButtonBlockElementWithStyle(actionID, "", btnText, slack.StylePrimary)
	btnBlock := slack.NewActionBlock("", btn)
	return btnBlock
}

// 定型ブロック: シングルテキスト
func SingleTextBlock(text string) []slack.Block {
	blocks := []slack.Block{SingleTextSectionBlock(util.Markdown, text)}
	return blocks
}
