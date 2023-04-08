// 機能: ヘルプ
package help

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 上位ヘルプ 取得
func getSuperiorHelps(hd data.HelpsData) map[string]map[string]string {
	superiorHelps := map[string]map[string]string{}

	for cmdName, cmdInfo := range hd {
		// サブコマンドリスト 作成
		sub_names := []string{}
		for _, subCmdInfo := range cmdInfo.Sub {
			sub_names = append(sub_names, subCmdInfo.Name)
		}

		sub_names_str := strings.Join(sub_names, ", ")
		superiorHelps[cmdName] = map[string]string{"desc": cmdInfo.Desc, "sub": sub_names_str}
	}

	return superiorHelps
}

// 下位ヘルプ 取得
func getInferiorHelps(hd data.HelpsData, cmdName string) ([]map[string]string, map[string][]map[string]string) {
	inferiorHelps, inferiorExs := []map[string]string{}, map[string][]map[string]string{}

	for _, subCmdInfo := range hd[cmdName].Sub {
		subCmdName := subCmdInfo.Name

		// サブコマンド 説明
		inferiorHelps = append(inferiorHelps, map[string]string{"name": subCmdName, "desc": subCmdInfo.Desc})
		inferiorExs[subCmdName] = []map[string]string{}

		// サブコマンド 使用例
		for _, ex := range subCmdInfo.Ex {
			q := fmt.Sprintf("%s %s %s %s", util.Cmd, cmdName, subCmdName, ex.Query)
			inferiorExs[subCmdName] = append(inferiorExs[subCmdName], map[string]string{"q": q, "desc": ex.Desc})
		}
	}

	return inferiorHelps, inferiorExs
}

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string, isEmptyValues bool) (blocks []slack.Block, responseType string, ok bool) {
	var (
		hd  data.HelpsData
		err error
	)

	responseType = util.Ephemeral

	// ヘルプデータ 読み込み
	if hd, err = data.ReadHelps(); err != nil {
		blocks, ok = post.ErrBlocksHelpsData(err, util.DataReadErr), false
		return blocks, responseType, ok
	}

	if isEmptyValues {
		// コマンド一覧（全体ヘルプ）
		ok = true

		// ブロック: ヘッダ・Tips
		headerSection := post.SingleTextSectionBlock(util.Markdown, post.InfoText(fmt.Sprintf("labotGo では以下コマンド *%s ###* を使用できます", util.Cmd)))
		tipsSection := post.TipsSection([]string{fmt.Sprintf("サブコマンド `$$$` は `%s ### $$$` の形式で使用できます", util.Cmd)})
		blocks = append(blocks, headerSection, tipsSection, util.Divider())

		// ブロック: 上位ヘルプ
		for cmdName, cmdInfo := range getSuperiorHelps(hd) {
			helpSection := post.SuperiorHelpSection(cmdName, cmdInfo)
			blocks = append(blocks, helpSection)
		}
	} else {
		// 指定コマンドヘルプ
		if len(cmdValues) == 1 {
			ok = true
			cmdName := cmdValues[0]
			cmdInfo, cmdOK := hd[cmdName]

			if !cmdOK {
				// 未定義コマンド指定
				textSection := post.SingleTextSectionBlock(util.Markdown, post.ErrText(fmt.Sprintf("指定したコマンド `%s` は存在しません", cmdName)))
				tipsSection := post.TipsSection([]string{fmt.Sprintf("指定できるコマンド一覧は `%s help` で確認できます", util.Cmd)})
				ok, blocks = false, append(blocks, textSection, tipsSection)
			} else {
				// ブロック: ヘッダ
				headerText := post.InfoText(fmt.Sprintf("labotGo コマンド - *%s %s*: %s", util.Cmd, cmdName, cmdInfo.Desc))
				headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
				blocks = append(blocks, headerSection)

				if len(hd[cmdName].Sub) == 0 {
					// サブコマンドなし
					for _, ex := range hd[cmdName].Ex {
						exFormatObject := post.TxtBlockObj(util.Markdown, fmt.Sprintf("*%s %s %s*", util.Cmd, cmdName, ex.Query))
						exDescOBject := post.TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", ex.Desc))

						exSection := slack.NewContextBlock("", []slack.MixedElement{exFormatObject, exDescOBject}...)
						blocks = append(blocks, exSection)
					}
				} else {
					// ブロック: 下位ヘルプ
					inferiorHelps, inferiorExs := getInferiorHelps(hd, cmdName)
					for _, subCmdInfo := range inferiorHelps {
						subCmdName := subCmdInfo["name"]
						helpSections := post.InferiorHelpSections(subCmdInfo, inferiorExs[subCmdName], subCmdName)
						blocks = append(blocks, helpSections...)
					}
				}
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
