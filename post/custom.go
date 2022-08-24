// メッセージ投稿
package post

import (
	"fmt"

	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// シングルテキストブロック 作成
func CreateSingleTextBlock(text string) []slack.Block {
	blocks := []slack.Block{slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, text, false, false), nil, nil,
	)}
	return blocks
}

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
func CreateOptionBlockObject(options []string) []*slack.OptionBlockObject {
	// ref.) https://github.com/slack-go/slack/blob/03f86be11aa50ac65d66f3917e250d3257389028/examples/modal_users/users.go#L92
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))

	for _, text := range options {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(text, optionText, nil))
	}

	return optionBlockObjects
}

// Tips セクション 作成
func CreateTipsSection(tips []string) *slack.ContextBlock {
	elements := make([]slack.MixedElement, 0, len(tips))
	for _, str := range tips {
		textBlockObject := slack.NewTextBlockObject(Markdown, fmt.Sprintf(":bulb: %s", str), false, false)
		elements = append(elements, textBlockObject)
	}

	tipsSection := slack.NewContextBlock("", elements...)
	return tipsSection
}

// エラーテキスト
func ErrText(text string) string {
	return fmt.Sprintf(":x: %s", text)
}

// 成功テキスト
func ScsText(text string) string {
	return fmt.Sprintf(":white_check_mark: %s", text)
}

// 情報テキスト
func InfoText(text string) string {
	return fmt.Sprintf(":diamond_shape_with_a_dot_inside: %s", text)
}
