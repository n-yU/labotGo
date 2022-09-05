// 機能: 書籍管理
package book

import (
	"fmt"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍データリセット リクエスト
func getBlocksResetRequest() (blocks []slack.Block) {
	// ブロック: ヘッダ
	headerText := post.InfoText("書籍データをリセットします\n\n")
	headerText += "*確認のため，リセットコードを入力してください*"
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{"コードが正しい場合，リセットボタンをクリックすると即座にリセットされます"}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: リセットコード 入力
	CodeInputSection := post.InputResetCodeSection(aid.ResetBookInputCode)

	// ブロック: リセットボタン
	actionBtnBlock := post.BtnOK("リセット", aid.ResetBook)

	blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), CodeInputSection, actionBtnBlock}
	return blocks
}

// 書籍データリセット
func ResetBook(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var resetCode string
	util.Logger.Printf("書籍データリセット (from %s): %+v\n", actionUserID, blockActions)

	// リセットコード 取得
	for _, action := range blockActions {
		for actionID, values := range action {
			switch actionID {
			case aid.ResetBookInputCode:
				resetCode = values.Value
			default:
			}
		}
	}

	// リセットコード 検証
	if blocks = post.ResetCodeValidateBlock(resetCode); len(blocks) > 0 {
		return blocks
	}

	// book index リセット
	if err := es.ResetIndex(util.EsBookIndex, util.EsBookMappingPath()); err != nil {
		text := post.ErrText(fmt.Sprintf("以下のエラーにより書籍データのリセットに失敗しました\n\n%s", err))
		blocks = post.SingleTextBlock(text)
		return blocks
	}

	blocks = post.SingleTextBlock(post.ScsText("書籍データのリセットに成功しました"))
	util.Logger.Println("書籍データのリセットに成功しました")
	return blocks
}
