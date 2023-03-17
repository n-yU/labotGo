// 機能: 書籍管理
package book

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string, cmdUserID string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType, subValues := cmdValues[0], cmdValues[1:]; subType {
	case "register":
		blocks, ok = getBlocksRegisterRequest(), true
	case "register-bulk":
		blocks, ok = getBlocksRegisterBulkRequest(), true
	case "reset":
		blocks, ok = getBlocksResetRequest(), true
	case "delete":
		blocks, ok = getBlocksDeleteRequest(), true
	case "search":
		blocks, ok = getBlocksSearch(subValues, cmdUserID), true
	case "list":
		blocks, ok = getBlocksList(cmdUserID), true
	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s book *%s* %s を使用することはできません", util.Cmd, subType, strings.Join(subValues, " ")))
		blocks, ok = post.SingleTextBlock(text), false
	}
	responseType = util.Ephemeral
	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionID string, callback slack.InteractionCallback) (err error) {
	var (
		blocks       = []slack.Block{}
		actionUserID = callback.User.ID
	)

	switch {
	case actionID == aid.RegisterBookRequest:
		blocks = RegisterBookConfirm(actionUserID, callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.RegisterBook+"_"):
		ISBN := strings.Split(actionID, "_")[1]
		blocks = RegisterBook(actionUserID, callback.BlockActionState.Values, ISBN)
	case actionID == aid.RegisterBulkBook:
		blocks = RegisterBulkBook(actionUserID, callback.BlockActionState.Values)
	case actionID == aid.ResetBook:
		blocks = ResetBook(actionUserID, callback.BlockActionState.Values)
	case actionID == aid.DeleteBookRequest:
		blocks = DeleteBookConfirm(actionUserID, callback.BlockActionState.Values)
	case strings.HasPrefix(actionID, aid.DeleteBook+"_"):
		ISBN := strings.Split(actionID, "_")[1]
		blocks = DeleteBook(actionUserID, callback.BlockActionState.Values, ISBN)
	case strings.HasPrefix(actionID, aid.BorrowBook):
		blocks = BorrowBook(actionUserID, actionID)
	case strings.HasPrefix(actionID, aid.ReturnBook):
		blocks = ReturnBook(actionUserID, actionID)
	}

	if len(blocks) > 0 {
		err = post.PostMessage(callback, blocks, util.Ephemeral)
	}
	return err
}

// ISBNコード 取得
func getISBN(actionID string, blockActions map[string]map[string]slack.BlockAction) (ISBNString string) {
	for _, action := range blockActions {
		for aID, values := range action {
			switch aID {
			case actionID:
				ISBNString = values.Value
			}
		}
	}
	return ISBNString
}

// 複数ISBNコード 取得
func getMultiISBN(actionID string, blockActions map[string]map[string]slack.BlockAction) (ISBNStringList []string) {
	var ISBNStringValues string
	for _, action := range blockActions {
		for aID, values := range action {
			switch aID {
			case actionID:
				ISBNStringValues = values.Value
			}
		}
	}

	for _, ISBNString := range strings.Split(ISBNStringValues, "\n") {
		// 空行無視
		if len(ISBNString) > 0 {
			ISBNStringList = append(ISBNStringList, ISBNString)
		}
	}
	return ISBNStringList
}

// ISBN10 -> ISBN13 変換
func convertISBN10to13(ISBN10 string) string {
	ISBN13, checkDigit := "978"+ISBN10[:9], 0

	for i, r := range ISBN13 {
		d := int(r - '0')
		if i%2 > 0 {
			d *= 3
		}
		checkDigit += d
	}

	checkDigit = (10 - checkDigit%10) % 10
	ISBN13 += strconv.Itoa(checkDigit)
	util.Logger.Printf("ISBN10: %s -> ISBN13: %s\n", ISBN10, ISBN13)
	return ISBN13
}

// ISBNコード バリデーションチェック
func validateISBN(ISBN string) (string, []slack.Block) {
	var blocks []slack.Block
	codeLen := len(ISBN)

	if codeLen == 10 {
		// ISBN10
		if _, err := strconv.Atoi(ISBN[:9]); err != nil {
			blocks = post.SingleTextBlock(post.ErrText("指定した ISBNコード に数字以外が含まれています（10桁の場合は末尾 `X` も可）"))
		} else if _, err := strconv.Atoi(ISBN[9:10]); (err != nil) && (ISBN[9:10] != "X") {
			blocks = post.SingleTextBlock(post.ErrText("10桁の ISBNコード の末尾は必ず数字か `X` です"))
		} else {
			// ISBN10 -> ISBN13
			ISBN = convertISBN10to13(ISBN)
		}
	} else if codeLen == 13 {
		// ISBN13
		if !strings.HasPrefix(ISBN, "978") {
			blocks = post.SingleTextBlock(post.ErrText("13桁の ISBNコード の先頭3桁は必ず `978` です"))
		} else if _, err := strconv.Atoi(ISBN); err != nil {
			blocks = post.SingleTextBlock(post.ErrText("指定した ISBNコード に数字以外が含まれています"))
		}
	} else {
		blocks = post.SingleTextBlock(post.ErrText("ISBNコード は 10桁 or 13桁 で指定する必要があります"))
	}

	return ISBN, blocks
}

// OpenBD 書籍情報取得
func requestOpenBD(ISBN string) (res *http.Response, err error) {
	// ref.) https://noumenon-th.net/programming/2019/09/04/http-get/
	request, err := http.NewRequest("GET", util.OpenBD, nil)
	if err != nil {
		return res, err
	}

	params := request.URL.Query()
	params.Add("isbn", ISBN)
	request.URL.RawQuery = params.Encode()

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{Timeout: timeout}

	res, err = client.Do(request)
	return res, err
}

// 書籍サマリ 取得
func getBookSummary(ISBN string) (bookSummary data.BookSummary, blocks []slack.Block) {
	body, err := es.GetDocument(data.BookSummary{ISBN: ISBN})

	if err != nil {
		text := post.ErrText(fmt.Sprintf("次のエラーにより書籍取得に失敗しました\n\n%v", err))
		blocks = post.SingleTextBlock(text)

		util.Logger.Printf("書籍サマリの取得に失敗しました（ISBN: %s）\n", ISBN)
		util.Logger.Println(util.ReferErrorDetail)
		util.Logger.Println(err)
		return bookSummary, blocks
	}

	bookSummary = body.(data.BookSummary)
	util.Logger.Printf("書籍情報: %s %s %s\n", bookSummary.Title, bookSummary.Publisher, bookSummary.Authors)
	return bookSummary, blocks
}
