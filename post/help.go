// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型ブロック: ヘルプデータエラー
func ErrBlocksHelpsData(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case util.DataReadErr:
		text = "ヘルプデータの読み込みに失敗しました"
	case util.DataWriteErr:
		text = "ヘルプデータの書き込みに失敗しました"
	default:
		util.Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := SingleTextSectionBlock(util.PlainText, ErrText(text))
	tipsSection := TipsSection(TipsDataError(util.HelpDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	util.Logger.Println(text)
	util.Logger.Println(err)
	return blocks
}

// 定型セクション: 上位ヘルプ
func SuperiorHelpSection(cmdName string, cmdInfo map[string]string) *slack.ContextBlock {
	cmdNameObject := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s %s*", util.Cmd, cmdName))
	if len(cmdInfo["sub"]) == 0 {
		cmdInfo["sub"] = "なし"
	}
	cmdDescObject := TxtBlockObj(util.Markdown, cmdInfo["desc"])
	cmdSubObject := TxtBlockObj(util.Markdown, fmt.Sprintf("サブ: %s", cmdInfo["sub"]))

	helpSection := slack.NewContextBlock("", []slack.MixedElement{cmdNameObject, cmdDescObject, cmdSubObject}...)
	return helpSection
}

// 定型セクションリスト: 下位ヘルプ
func InferiorHelpSections(subCmdInfo map[string]string, subCmdExs []map[string]string, subCmdName string) (blocks []slack.Block) {
	// ブロック: サブコマンド 説明
	descSection := SingleTextSectionBlock(util.Markdown, fmt.Sprintf("*%s*: %s", subCmdName, subCmdInfo["desc"]))
	blocks = append(blocks, util.Divider(), descSection)

	for _, ex := range subCmdExs {
		// ブロック: サブコマンド 使用例
		exFormatObject := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", ex["q"]))
		exDescOBject := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", ex["desc"]))

		exSection := slack.NewContextBlock("", []slack.MixedElement{exFormatObject, exDescOBject}...)
		blocks = append(blocks, exSection)
	}
	return blocks
}
