// 機能: 書籍管理
package book

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍検索
func getBlocksSearch(values []string, cmdUserID string) (blocks []slack.Block) {
	var (
		bookSummary data.BookSummary
		hitISBNs    []string
	)
	util.Logger.Println("書籍検索リクエスト")

	// コマンド形式 チェック
	if len(values) < 1 || len(values) > 2 {
		headerText := post.ErrText("書籍検索コマンドの形式が不適切です")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		blocks = []slack.Block{headerSection, post.TipsSearchSection()}
		return blocks
	}

	// クエリ／最大表示数 取得 -> バリデーションチェック
	query, maxDisplayBooks := strings.Replace(values[0], "_", " ", -1), 5
	if len(values) == 2 {
		val1, err := strconv.Atoi(values[1])
		if err != nil || val1 <= 0 {
			headerText := post.ErrText(fmt.Sprintf("指定した最大表示数 `%s` は自然数ではありません", values[1]))
			headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
			blocks = []slack.Block{headerSection, post.TipsSearchSection()}
			return blocks
		} else {
			maxDisplayBooks = val1
		}
	}
	util.Logger.Printf("クエリ: %s / 最大表示数: %d\n", query, maxDisplayBooks)

	// book 内の書籍検索
	results, err := es.SearchDocument(bookSummary, query, maxDisplayBooks)
	if err != nil {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf("次のエラーにより書籍検索に失敗しました\n\n%v", err)))
		return blocks
	}

	// 検索ヒット数0件
	bookResults := results.([]data.BookSummary)
	nHits := len(bookResults)
	if nHits == 0 {
		headerText := post.ErrText(fmt.Sprintf("キーワード `%s` で書籍を検索しましたが見つかりませんでした\n", query))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{"より一般的な表現かつ複数のキーワードを使用するとヒットしやすくなります"}
		tipsSection := post.TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection}
		return blocks
	}

	// ブロック: ヘッダ
	headerText := post.ScsText(fmt.Sprintf("書籍検索結果: *%s* （全%d件）", query, nHits))
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
	blocks = []slack.Block{headerSection, util.Divider()}

	// ブロック: 検索結果
	for i, bookSummary := range bookResults {
		hitISBNs = append(hitISBNs, bookSummary.ISBN)

		searchInfoText := post.TxtBlockObj(util.Markdown, fmt.Sprintf("[%d] %s", i+1, bookSummary.ISBN))
		bookOwner := bookSummary.GetOwner()
		bookOwnerDisplay := fmt.Sprintf("%s（借りられます）", bookOwner)
		if bookOwner != util.DefaultBookOwner() {
			bookOwnerDisplay = fmt.Sprintf("<@%s>", bookOwner)
		}
		bookOwnerText := post.TxtBlockObj(util.Markdown, fmt.Sprintf("オーナー: %s", bookOwnerDisplay))
		searchInfoFields := []*slack.TextBlockObject{searchInfoText, bookOwnerText}

		// 図書館ボタン
		var searchInfoBorrowBtn *slack.ButtonBlockElement
		switch bookOwner {
		case util.DefaultBookOwner():
			searchInfoBorrowBtnActionID := strings.Join([]string{aid.BorrowBook, bookSummary.ISBN}, "_")
			searchInfoBorrowBtn = post.CustomBtn("OK", "借りる", searchInfoBorrowBtnActionID)
		case cmdUserID:
			searchInfoBorrowBtnActionID := strings.Join([]string{aid.ReturnBook, bookSummary.ISBN}, "_")
			searchInfoBorrowBtn = post.CustomBtn("DEF", "返却", searchInfoBorrowBtnActionID)
		default:
			searchInfoBorrowBtnActionID := strings.Join([]string{aid.BorrowBookDeny, bookSummary.ISBN}, "_")
			searchInfoBorrowBtn = post.CustomBtn("NG", "貸出中", searchInfoBorrowBtnActionID)
		}

		searchInfoSection := slack.NewSectionBlock(nil, searchInfoFields, slack.NewAccessory(searchInfoBorrowBtn))
		blocks = append(blocks, searchInfoSection, post.InfoBookSection(bookSummary), util.Divider())
	}

	util.Logger.Printf("書籍検索 \"%s\"（全%d件）に成功しました\n", query, nHits)
	util.Logger.Printf("ヒット書籍: %s\n", strings.Join(hitISBNs, ", "))
	return blocks
}
