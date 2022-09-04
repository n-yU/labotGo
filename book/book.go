// 機能: スプレッドシート書籍管理
package book

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType, subValues := cmdValues[0], cmdValues[1:]; subType {
	case "register":
		blocks, ok = getBlockRegisterRequest(), true
	case "delete":
		blocks, ok = getBlockDeleteRequest(), true
	default:
		text := post.ErrText(fmt.Sprintf("コマンド %s team *%s* %s を使用することはできません", util.Cmd, subType, strings.Join(subValues, " ")))
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
	case actionID == aid.RegisterBook:
		blocks = RegisterBook(actionUserID, callback.BlockActionState.Values)
	}

	if len(blocks) > 0 {
		err = post.PostMessage(callback, blocks, util.Ephemeral)
	}
	return err
}

// ISBNコード 取得
func getISBN(blockActions map[string]map[string]slack.BlockAction) (ISBNString string) {
	for _, action := range blockActions {
		for actionID, values := range action {
			switch actionID {
			case aid.RegisterBookInputISBN:
				ISBNString = values.Value
			}
		}
	}
	return ISBNString
}

// ISBNコード バリデーションチェック
func validateISBN(ISBN string) (blocks []slack.Block) {
	if _, err := strconv.Atoi(ISBN); err != nil {
		blocks = post.SingleTextBlock(post.ErrText("指定した ISBNコード に数字以外が含まれています"))
	} else if len(ISBN) != 10 && len(ISBN) != 13 {
		blocks = post.SingleTextBlock(post.ErrText("ISBNコード は10桁もしくは13桁で指定する必要があります"))
	} else if len(ISBN) == 13 && !strings.HasPrefix(ISBN, "978") {
		blocks = post.SingleTextBlock(post.ErrText("13桁の ISBNコード の接頭記号は必ず `978` です（もしくはこれを省略できます）"))
	}
	return blocks
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
