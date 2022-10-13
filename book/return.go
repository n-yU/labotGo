// 機能: 書籍管理
package book

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍返却
func ReturnBook(actionUserID string, actionID string) (blocks []slack.Block) {
	var text string
	util.Logger.Printf("書籍返却（from %s）\n", actionUserID)
	actionTypeISBN := strings.Split(actionID, "_")
	_, ISBN := actionTypeISBN[0], actionTypeISBN[1]

	// 書籍サマリ 取得
	bookSummary, blocks := getBookSummary(ISBN)
	if len(blocks) > 0 {
		return blocks
	}

	// 書籍返却 応答
	if err := bookSummary.ChangeOwner(util.DefaultBookOwner()); err != nil {
		text = post.ErrText(fmt.Sprintf("次のエラーにより，指定した書籍（ISBN: %s）の返却に失敗しました\n\n%v", bookSummary.ISBN, err))
	} else {
		text = post.ScsText(fmt.Sprintf("書籍: *%s* - ISBN: %s が *%s* に返却されました",
			bookSummary.Title, bookSummary.ISBN, util.DefaultBookOwner(),
		))
	}

	blocks = post.SingleTextBlock(text)
	return blocks
}
