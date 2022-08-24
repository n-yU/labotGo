// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) ([]slack.Block, string, bool) {
	var (
		blocks       []slack.Block
		responseType string
		ok           bool
		subType      = cmdValues[0]
	)

	switch subType {
	case "add":
		blocks, responseType, ok = getBlockAdd(), Ephemeral, true
	case "edit":

	case "delete":

	case "list":

	default:
		text := fmt.Sprintf("コマンド %s member *%s* を使用することはできません\n", Cmd, subType)
		blocks, responseType, ok = post.CreateSingleTextBlock(text), Ephemeral, true
	}

	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionId string, callback slack.InteractionCallback) error {
	var err error

	switch {
	case strings.HasSuffix(actionId, "Add"):
		blocks := AddMember(callback.BlockActionState.Values)
		err = post.PostMessage(callback, blocks, Ephemeral)
	}

	return err
}
