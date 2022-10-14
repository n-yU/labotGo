// メンバー管理
package member

import (
	"fmt"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メンバーリスト表示
func getBlocksListMember() (blocks []slack.Block) {
	var ok bool
	util.Logger.Println("メンバーリスト表示リクエスト")

	// メンバーデータ 読み込み
	if md, err := data.ReadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataReadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText(fmt.Sprintf("*labotGo に追加されている全メンバー（%d人）は以下の通りです*", len(md)))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTips := post.TipsSection([]string{util.TipsMasterUser()})

		blocks = []slack.Block{headerSection, headerTips, util.Divider()}

		// ブロック: メンバー情報
		displayCount := 0
		for userID, m := range md {
			memberInfoSections := post.InfoMemberSection(m.Image24, userID, m.TeamNames, m.TeamNames, m.Created)
			for _, memberInfoSec := range memberInfoSections {
				blocks = append(blocks, memberInfoSec)
			}
			blocks = append(blocks, util.Divider())

			// ブロック最大数（=50）超過 -> 残メンバー表示省略
			displayCount++
			if len(blocks)+3 > util.MaxBlocks {
				blockLimitTips := post.TipsSection([]string{
					fmt.Sprintf("メッセージ表示の限界に達したため， *残り %d 人* のメンバーの表示は省略されています", len(md)-displayCount),
					fmt.Sprintf("特定のメンバーの情報を確認・編集したい場合は `%s member edit` を使用してください", util.Cmd),
				})
				blocks = append(blocks, blockLimitTips)
				break
			}
		}

		ok = true
		util.Logger.Println("メンバーリスト表示に成功しました")
	}

	if !ok {
		util.Logger.Println("メンバーリスト表示に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
