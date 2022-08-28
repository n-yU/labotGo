// チーム管理
package team

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// チームリスト表示
func getBlockListTeam(values []string) (blocks []slack.Block) {
	var ok bool
	util.Logger.Println("チームリスト表示リクエスト")

	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, util.DataLoadErr)
	} else {
		var teamNames []string

		if len(values) > 1 {
			blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
				"表示するチームを指定する場合， `%s team list A,B,C` のようにカンマ区切りで指定してください", util.Cmd,
			)))
		} else {
			if len(values) == 0 {
				// "/lab team": 全チーム表示
				teamNames = td.GetAllNames()

				// ブロック: ヘッダ
				headerText := post.InfoText(fmt.Sprintf("*labotGo に追加されている全%dチームは以下の通りです*", len(td)))
				headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
				// ブロック: ヘッダ Tips
				headerTips := post.TipsSection([]string{util.TipsMasterTeam})

				blocks = []slack.Block{headerSection, headerTips, util.Divider()}
			} else {
				// "/lab team team1,team2,...": 指定チーム表示
				teamNames = util.UniqueSlice(strings.Split(values[0], ","))

				// ブロック: ヘッダ
				headerText := post.InfoText(fmt.Sprintf("*指定した%dチームの詳細は以下の通りです*", len(teamNames)))
				headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

				blocks = []slack.Block{headerSection, util.Divider()}
			}

			var unknownTeams []string // 追加されていない指定チーム名リスト
			for _, teamName := range teamNames {
				if _, ok := td[teamName]; !ok {
					unknownTeams = append(unknownTeams, teamName)
					continue
				}
				// ブロック: チーム情報
				teamUserIDs := td[teamName].UserIDs
				teamInfoSection := post.InfoTeamSection(teamName, teamName, teamUserIDs, teamUserIDs)
				blocks = append(blocks, teamInfoSection)
			}

			if len(unknownTeams) > 0 {
				blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
					"指定したチーム `%s` は labotGo に追加されていません", strings.Join(unknownTeams, "`, `"),
				)))
			} else {
				ok = true
				util.Logger.Println("チームリスト表示に成功しました")
			}
		}
	}

	if !ok {
		util.Logger.Println("チームリスト表示に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
