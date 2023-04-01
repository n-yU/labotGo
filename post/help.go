// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型セクション: 上位ヘルプ
func TopHelpSection(cmdName string, cmdInfo map[string]string) *slack.ContextBlock {
	cmdNameObject := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", cmdName))
	if len(cmdInfo["sub"]) == 0 {
		cmdInfo["sub"] = "なし"
	}
	cmdSubObject := TxtBlockObj(util.Markdown, fmt.Sprintf("サブ: %s", cmdInfo["sub"]))
	cmdDescObject := TxtBlockObj(util.Markdown, cmdInfo["desc"])

	elements := []slack.MixedElement{cmdNameObject, cmdSubObject, cmdDescObject}
	helpSection := slack.NewContextBlock("", elements...)
	return helpSection
}
