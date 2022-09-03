// 機能: メンバーグルーピング
package group

import "github.com/slack-go/slack"

// グルーピングリクエスト: カスタムメンバーモード
func getBlocksCustom() (blocks []slack.Block) {
	return blocks
}

// カスタムメンバーグルーピング
func GroupCustom(
	actionUserID string, blockActions map[string]map[string]slack.BlockAction,
) (blocks []slack.Block, responseType string) {

	return blocks, responseType
}
