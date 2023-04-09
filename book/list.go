// 機能: 書籍管理
package book

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 借り出し書籍リスト表示
func getBlocksList(cmdUserID string) (blocks []slack.Block) {
	util.Logger.Println("借り出し書籍リスト表示リクエスト")
	ISBNs := data.GetOwnerBooks(cmdUserID)
	util.Logger.Printf("オーナー: %s - %v", cmdUserID, ISBNs)

	if nBooks := len(ISBNs); nBooks == 0 {
		// 借り出し書籍なし
		headerText := post.InfoText("*あなた* が現在借りている書籍はありません")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{fmt.Sprintf("書籍は `%s book search キーワード` で表示される検索結果から借りることができます", util.Cmd)}
		tipsSection := post.TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection}
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText(fmt.Sprintf("*あなた* は次の %d冊 の書籍を借りています", nBooks))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{"既に本棚に戻している書籍がある場合，返却ボタンを押してください"}
		tipsSection := post.TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection, util.Divider()}

		// ブロック: 借り出し書籍一覧
		for i, ISBN := range ISBNs {
			if bookSummary, BookSummaryBlocks := getBookSummary(ISBN); len(BookSummaryBlocks) > 0 {
				return BookSummaryBlocks
			} else {
				searchInfoText := post.TxtBlockObj(util.Markdown, fmt.Sprintf("[%d] %s", i+1, bookSummary.ISBN))
				bookOwner := bookSummary.GetOwner()
				bookOwnerText := post.TxtBlockObj(util.Markdown, fmt.Sprintf("オーナー: <@%s>", bookOwner))
				searchInfoFields := []*slack.TextBlockObject{searchInfoText, bookOwnerText}

				// 図書館ボタン（返却固定）
				searchInfoBorrowBtnActionID := strings.Join([]string{aid.ReturnBook, bookSummary.ISBN}, "_")
				searchInfoBorrowBtn := post.CustomBtn("DEF", "返却", searchInfoBorrowBtnActionID)

				searchInfoSection := slack.NewSectionBlock(nil, searchInfoFields, slack.NewAccessory(searchInfoBorrowBtn))
				blocks = append(blocks, searchInfoSection, post.InfoBookSection(bookSummary), util.Divider())
			}
		}
	}

	util.Logger.Println("借り出し書籍リスト表示に成功しました")
	return blocks
}
