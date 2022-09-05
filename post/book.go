// メッセージ投稿
package post

import (
	"fmt"
	"net/http"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型セクション: ISBN Tips
func TipsIBSNSection() *slack.ContextBlock {
	tipsText := []string{"ISBNコードは *978* から始まる13桁の番号で，書籍の裏表紙に記載されています"}
	tipsSection := TipsSection(tipsText)
	return tipsSection
}

// 定型セクション: ISBNコード入力
func InputISBNSection(actionID string) *slack.InputBlock {
	inputSectionText := TxtBlockObj(util.PlainText, "ISBNコード")
	inputSectionHint := TxtBlockObj(util.PlainText, "13桁もしくは接頭記号の 978 に続く10桁のみ入力してください")
	inputText := TxtBlockObj(util.PlainText, "ISBNコードを入力")
	input := slack.NewPlainTextInputBlockElement(inputText, actionID)
	input.MinLength, input.MaxLength = 10, 13
	inputSection := slack.NewInputBlock("", inputSectionText, inputSectionHint, input)
	return inputSection
}

// 定型セクション: 書籍情報
func InfoBookSection(bookSummary data.BookSummary) (infoSection *slack.SectionBlock) {
	// 書籍情報
	util.Logger.Printf("%+v\n", bookSummary)
	titleField := TxtBlockObj(util.Markdown, fmt.Sprintf("タイトル:\n*%s*", bookSummary.Title))
	publisherField := TxtBlockObj(util.Markdown, fmt.Sprintf("出版社:\n*%s*", bookSummary.Publisher))
	pubdateField := TxtBlockObj(util.Markdown, fmt.Sprintf("出版日:\n*%s*", bookSummary.PubdateYMD))
	authorsField := TxtBlockObj(util.Markdown, fmt.Sprintf("著者:\n*%s*", bookSummary.Authors))
	infoFields := []*slack.TextBlockObject{titleField, publisherField, pubdateField, authorsField}

	// 書籍表紙画像
	if len(bookSummary.Cover) > 0 {
		coverElement := slack.NewImageBlockElement(bookSummary.Cover, bookSummary.Title)
		infoAccessory := slack.NewAccessory(coverElement)
		infoSection = slack.NewSectionBlock(nil, infoFields, infoAccessory)
	} else {
		infoSection = slack.NewSectionBlock(nil, infoFields, nil)
	}

	return infoSection
}

// 定型ブロック: 書籍情報有無確認
func UnknownBookBlock(book *data.Book, ISBN string) (blocks []slack.Block) {
	if book == nil {
		headerText := ErrText(fmt.Sprintf("<https://openbd.jp/|OpenBD> に指定した書籍 `%s` が存在しない可能性があります", ISBN))
		headerSection := SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{fmt.Sprintf(
			"ブラウザで <https://api.openbd.jp/v1/get?isbn=%s&pretty|コチラ> にアクセスして `null` と表示されている場合は書籍情報が存在しません", ISBN,
		)}
		tipsSection := TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection}
	}
	return blocks
}

// 定型ブロック: OpenBD リクエストエラー
func ErrBlocksRequestOpenBD(err error, res *http.Response) []slack.Block {
	util.Logger.Println(err)

	text := "<https://openbd.jp/|OpenBD> への書籍情報の取得を試みましたが，次のエラーにより失敗しました\n\n"
	if err != nil {
		text += fmt.Sprintf("%v", err)
	} else {
		text += res.Status
	}

	blocks := SingleTextBlock(ErrText(text))
	return blocks
}

// 定型ブロック: レスポンスボディ 読み込みエラー
func ErrBlocksReadResponseBody(err error) []slack.Block {
	util.Logger.Println(err)
	text := fmt.Sprintf("レスポンスボディの読み込み時に，次のエラーにより失敗しました\n\n%v", err)
	blocks := SingleTextBlock(ErrText(text))
	return blocks
}

// 定型ブロック: 書籍情報 読み込みエラー
func ErrBlocksLoadBookInfo(err error) []slack.Block {
	util.Logger.Println(err)
	text := ErrText(fmt.Sprintf("書籍情報の読み込み時に，次のエラーにより失敗しました\n\n%v", err))
	blocks := SingleTextBlock(text)
	return blocks
}
