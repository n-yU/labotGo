// 機能: 書籍管理
package book

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍登録リクエスト
func getBlocksRegisterRequest() (blocks []slack.Block) {
	// ブロック: ヘッダ
	headerText := post.InfoText("*登録したい書籍の ISBNコード を入力してください*")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsSection := post.TipsIBSNSection()

	// ブロック: ISBNコード 入力
	ISBNInputSection := post.InputISBNSection(aid.RegisterBookInputISBN)

	// ブロック: 確認ボタン
	actionBtnBlock := post.CustomBtnSection("OK", "確認", aid.RegisterBookRequest)

	blocks = []slack.Block{
		headerSection, headerTipsSection, util.Divider(), ISBNInputSection, actionBtnBlock,
	}
	return blocks
}

// 書籍登録確認（書籍情報取得）
func RegisterBookConfirm(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var books data.Books
	util.Logger.Printf("書籍登録リクエスト (from %s): %+v\n", actionUserID, blockActions)

	// ISBNコード 取得
	ISBN := getISBN(aid.RegisterBookInputISBN, blockActions)
	util.Logger.Printf("ISBNコード: %s\n", ISBN)

	// バリデーションチェック
	if ISBN, blocks = validateISBN(ISBN); len(blocks) > 0 {
		util.Logger.Println("ISBNコードの形式が不適切です．詳細は Slack に投稿されたメッセージを確認してください．")
		return blocks
	}

	// 書籍情報 取得（from OpenBD）
	res, err := requestOpenBD(ISBN)
	if err != nil || res.StatusCode != 200 {
		return post.ErrBlocksRequestOpenBD(err, res)
	}

	// レスポンスボディ 読み込み
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return post.ErrBlocksReadResponseBody(err)
	}

	// 書籍情報 読み込み
	if err := json.Unmarshal(body, &books); err != nil {
		return post.ErrBlocksLoadBookInfo(err)
	}

	// 書籍情報有無 確認
	if blocks := post.UnknownBookBlock(books[0], ISBN); len(blocks) > 0 {
		return blocks
	}

	// 書籍サマリ 取得
	(&books[0].BookSummary).SetPubdateYMD()
	books[0].SetContent()
	bookSummary := books[0].BookSummary
	util.Logger.Printf("書籍情報: %s %s %s\n", bookSummary.Title, bookSummary.Publisher, bookSummary.Authors)

	// 書籍サマリ 一時保存
	if _, ok := data.BookBuffer[actionUserID]; !ok {
		data.BookBuffer[actionUserID] = map[string]data.BookSummary{}
	}
	data.BookBuffer[actionUserID][ISBN] = bookSummary

	// ブロック: ヘッダ
	headerText := post.InfoText("以下の書籍を登録します．情報が正しければ *登録* ボタンをクリックしてください．")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{"間違いがある場合，上記フォームを入力し直して確認ボタンを再度クリックしてください"}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: 書籍情報
	infoBookSection := post.InfoBookSection(bookSummary)

	// ブロック: 登録ボタン
	actionBtnActionID := strings.Join([]string{aid.RegisterBook, ISBN}, "_")
	actionBtnBlock := post.CustomBtnSection("OK", "登録", actionBtnActionID)

	blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), infoBookSection, actionBtnBlock}
	return blocks
}

// 書籍登録
func RegisterBook(actionUserID string, blockActions map[string]map[string]slack.BlockAction, ISBN string) (blocks []slack.Block) {
	util.Logger.Printf("書籍登録 (from %s): %s %+v\n", actionUserID, ISBN, blockActions)
	bookSummary, ok := data.BookBuffer[actionUserID][ISBN]
	if !ok {
		text := post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は既に登録が完了しています", ISBN))
		blocks = post.SingleTextBlock(text)
		return blocks
	}

	// book-doc への書籍追加
	if err := es.PutDocument(bookSummary); err != nil {
		var text string
		switch err {
		case es.ErrDocAlreadyExist:
			text = post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は既に登録されています", ISBN))
		default:
			text = post.ErrText(fmt.Sprintf("次のエラーにより書籍登録に失敗しました\n\n%v", err))
		}
		blocks = post.SingleTextBlock(text)
		return blocks
	}

	// 書籍DB追加
	bookSummary.AddDB()

	// ブロック: ヘッダ
	headerText := post.ScsText(fmt.Sprintf("書籍: *%s* - ISBN: %s の登録に成功しました", bookSummary.Title, bookSummary.ISBN))
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{
		"上記フォームは ISBNコード の入力欄を書き換えることで再利用できます",
		fmt.Sprintf("現在 *%d* 冊の書籍が *labotGo* に登録されています", es.CountDoc(util.EsBookIndex)),
	}
	headerTipsSection := post.TipsSection(headerTipsText)
	blocks = []slack.Block{headerSection, headerTipsSection}

	util.Logger.Printf("書籍（ISBN: %s）の登録に成功しました\n", bookSummary.ISBN)
	delete(data.BookBuffer[actionUserID], ISBN)
	return blocks
}
