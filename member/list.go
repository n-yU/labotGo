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
func getBlockListMember() (blocks []slack.Block) {
	var ok bool
	util.Logger.Println("メンバーリスト表示リクエスト")

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.GetErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText(fmt.Sprintf("*labotGo に追加されている全メンバー（%d人）は以下の通りです*", len(md)))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTips := post.TipsSection([]string{util.TipsMasterUser()})

		blocks = []slack.Block{headerSection, headerTips, util.Divider()}

		// ブロック: メンバー情報
		for userID, m := range md {
			memberInfoSections := post.InfoMemberSection(md[userID].Image24, userID, m.TeamNames, m.TeamNames, m.Created)
			for _, memberInfoSec := range memberInfoSections {
				blocks = append(blocks, memberInfoSec)
			}
			blocks = append(blocks, util.Divider())
		}

		ok = true
		util.Logger.Println("メンバーリスト表示に成功しました")
	}

	if !ok {
		util.Logger.Println("メンバーリスト表示に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
