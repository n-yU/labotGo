// 機能: ヘルプ
package help

import (
	"fmt"

	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

func commandHelpMap(sub, desc string) map[string]string {
	return map[string]string{"sub": sub, "desc": desc}
}

// コマンド一覧（上位ヘルプ）
func getCommandHelps() map[string]map[string]string {
	commandHelps := map[string]map[string]string{
		"hello": commandHelpMap("", "labotGo が挨拶します（ボット動作チェック用）"),
	}
	return commandHelps
}

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string, isEmptyValues bool) (blocks []slack.Block, responseType string, ok bool) {
	responseType = util.Ephemeral

	if isEmptyValues {
		// コマンド一覧（全体ヘルプ）
		ok = true

		// ブロック: ヘッダ
		headerSection := post.SingleTextSectionBlock(util.Markdown, post.ScsText(fmt.Sprintf("labotGo では以下コマンド *%s ###* を使用できます", util.Cmd)))
		blocks = append(blocks, headerSection, util.Divider())

		// ブロック: 上位ヘルプ
		for cmdName, cmdInfo := range getCommandHelps() {
			helpSection := post.TopHelpSection(cmdName, cmdInfo)
			blocks = append(blocks, helpSection)
		}
	} else {
		// 指定コマンドヘルプ
		if len(cmdValues) == 1 {
			ok = true
			cmdName := cmdValues[0]
			switch cmdName {
			case "hello":
				blocks = helloBlocks()
			default:
				// 未定義コマンド指定
				textSection := post.SingleTextSectionBlock(util.Markdown, post.ErrText(fmt.Sprintf("指定したコマンド `%s` は存在しません", cmdName)))
				tipsSection := post.TipsSection([]string{fmt.Sprintf("指定できるコマンド一覧は `%s help` で確認できます", util.Cmd)})
				ok, blocks = false, append(blocks, textSection, tipsSection)
			}
		} else {
			textSection := post.SingleTextSectionBlock(util.Markdown, post.ErrText(fmt.Sprintf("%s *help* に2つ以上の引数を与えることはできません", util.Cmd)))
			tipsText := []string{
				fmt.Sprintf("例えば `shuffle` のヘルプを確認したいとき `%s help shuffle` と指定してください", util.Cmd),
				fmt.Sprintf("指定できるコマンドを確認したい場合，何も指定せずに `%s help` と入力してください", util.Cmd),
			}
			tipsSection := post.TipsSection(tipsText)

			ok, blocks = false, append(blocks, textSection, tipsSection)
		}
	}
	return blocks, responseType, ok
}

// 指定コマンドヘルプ
func helloBlocks() (blocks []slack.Block) {
	return blocks
}
