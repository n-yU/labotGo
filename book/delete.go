// 機能: 書籍管理
package book

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍削除リクエスト
func getBlocksDeleteRequest() (blocks []slack.Block) {
	// ブロック: ヘッダ
	headerText := post.InfoText("*削除したい書籍の ISBNコード を入力してください*")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsSection := post.TipsIBSNSection()

	// ブロック: ISBNコード 入力
	ISBNInputSection := post.InputISBNSection(aid.DeleteBookInputISBN)

	// ブロック: 確認ボタン
	actionBtnBlock := post.CustomBtnSection("OK", "確認", aid.DeleteBookRequest)

	blocks = []slack.Block{
		headerSection, headerTipsSection, util.Divider(), ISBNInputSection, actionBtnBlock,
	}
	return blocks
}

// 書籍削除確認（書籍情報取得）
func DeleteBookConfirm(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	util.Logger.Printf("書籍削除リクエスト (from %s): %+v\n", actionUserID, blockActions)

	// ISBNコード 取得
	ISBN := getISBN(aid.DeleteBookInputISBN, blockActions)
	util.Logger.Printf("ISBNコード: %s\n", ISBN)

	// バリデーションチェック
	if ISBN, blocks = validateISBN(ISBN); len(blocks) > 0 {
		util.Logger.Println("ISBNコードの形式が不適切です．詳細は Slack に投稿されたメッセージを確認してください．")
		return blocks
	}

	// 書籍サマリ 取得
	bookSummary, blocks := getBookSummary(ISBN)
	if len(blocks) > 0 {
		return blocks
	}

	// 書籍サマリ 一時保存
	if _, ok := data.BookBuffer[actionUserID]; !ok {
		data.BookBuffer[actionUserID] = map[string]data.BookSummary{}
	}
	data.BookBuffer[actionUserID][ISBN] = bookSummary

	// ブロック: ヘッダ
	headerText := post.InfoText("以下の書籍を削除します．情報が正しければ *削除* ボタンをクリックしてください．")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{"間違いがある場合，上記フォームを入力し直して確認ボタンを再度クリックしてください"}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: 書籍情報
	infoBookSection := post.InfoBookSection(bookSummary)

	// ブロック: 削除ボタン
	actionBtnActionID := strings.Join([]string{aid.DeleteBook, ISBN}, "_")
	actionBtnBlock := post.CustomBtnSection("OK", "削除", actionBtnActionID)

	blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), infoBookSection, actionBtnBlock}
	return blocks
}

// 書籍削除
func DeleteBook(actionUserID string, blockActions map[string]map[string]slack.BlockAction, ISBN string) (blocks []slack.Block) {
	util.Logger.Printf("書籍削除 (from %s): %s %+v\n", actionUserID, ISBN, blockActions)
	bookSummary, ok := data.BookBuffer[actionUserID][ISBN]
	if !ok {
		text := post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は既に削除が完了しています\n", ISBN))
		blocks = post.SingleTextBlock(text)
		return blocks
	}

	// book-doc からの書籍削除
	if err := es.DeleteDocument(bookSummary); err != nil {
		var text string
		switch err {
		case es.ErrDocNotFound:
			text = post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は既に削除されています\n", ISBN))
		default:
			text = post.ErrText(fmt.Sprintf("次のエラーにより書籍削除に失敗しました\n\n%v", err))
		}
		blocks = post.SingleTextBlock(text)
		return blocks
	}

	// ブロック: ヘッダ
	headerText := post.ScsText(fmt.Sprintf("書籍: *%s* - ISBN: %s の削除に成功しました", bookSummary.Title, bookSummary.ISBN))
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{
		"上記フォームは ISBNコード の入力欄を書き換えることで再利用できます",
		fmt.Sprintf("現在 *%d* 冊の書籍が *labotGo* に登録されています\n", es.CountDoc(util.EsBookIndex)),
	}
	headerTipsSection := post.TipsSection(headerTipsText)
	blocks = []slack.Block{headerSection, headerTipsSection}

	util.Logger.Println(fmt.Sprintf("書籍（ISBN: %s）の削除に成功しました\n", bookSummary.ISBN))
	delete(data.BookBuffer[actionUserID], ISBN)
	return blocks
}
