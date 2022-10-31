// 機能: 書籍管理
package book

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍借り出し
func BorrowBook(actionUserID string, actionID string) (blocks []slack.Block) {
	var text string
	util.Logger.Printf("書籍借り出し（from %s）\n", actionUserID)
	actionTypeISBN := strings.Split(actionID, "_")
	actionType, ISBN := actionTypeISBN[0], actionTypeISBN[1]

	// 書籍サマリ 取得
	bookSummary, blocks := getBookSummary(ISBN)
	if len(blocks) > 0 {
		return blocks
	}

	if actionType == aid.BorrowBookDeny {
		// 書籍貸し出し 拒否
		bookOwner := bookSummary.GetOwner()
		text = post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は <@%s> に貸出中のため借りられません", bookSummary.ISBN, bookOwner))
	} else {
		// 書籍貸し出し 応答
		if bookSummary.GetOwner() == actionUserID {
			text = post.ErrText(fmt.Sprintf("指定した書籍（ISBN: %s）は既に *あなた* が借りています", bookSummary.ISBN))
		} else if err := bookSummary.ChangeOwner(actionUserID); err != nil {
			text = post.ErrText(fmt.Sprintf("次のエラーにより，指定した書籍（ISBN: %s）の貸し出しに失敗しました\n\n%v", bookSummary.ISBN, err))
		} else {
			text = post.ScsText(fmt.Sprintf("書籍: *%s* - ISBN: %s が *あなた* に貸し出されました", bookSummary.Title, bookSummary.ISBN))
		}
	}

	blocks = post.SingleTextBlock(text)
	return blocks
}
