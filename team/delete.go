// チーム管理
package team

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// チーム削除リクエスト（チーム選択）
func getBlockDeleteTeamSelect() (blocks []slack.Block) {
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*削除したいチームを選択してください*")
		headerSection := post.SingleTextSectionBlock(Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTipsText := []string{"チームを選択すると確認のメッセージが表示されます"}
		headerTipsSection := post.TipsSection(headerTipsText)
		// ブロック: チーム選択
		teamSelectSection := post.SelectTeamsSection(td.GetAllEditedNames(), aid.DeleteTeamSelectTeam, []string{}, false)

		blocks = []slack.Block{headerSection, headerTipsSection, teamSelectSection}
	}
	return blocks
}

// チーム削除リクエスト（確認）
func DeleteTeamConfirm(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	Logger.Println("チーム削除リクエスト")

	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
	} else {
		var teamName string
		// ユーザID・チームリスト 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.DeleteTeamSelectTeam:
					teamName = values.SelectedOption.Value
				default:
				}
			}
		}
		teamUserIDs := td[teamName].UserIDs
		Logger.Printf("チーム名: %s / メンバーリスト: %v\n", teamName, teamUserIDs)

		headerSection := post.SingleTextSectionBlock(Markdown, "*以下チームを削除しますか？*")
		teamInfoSection := post.InfoTeamSection(teamName, teamName, teamUserIDs, teamUserIDs)
		actionBtnActionId := strings.Join([]string{aid.DeleteTeam, teamName}, "_")
		actionBtnBlock := post.BtnOK("削除", actionBtnActionId)

		blocks = []slack.Block{headerSection, teamInfoSection, actionBtnBlock}
	}

	return blocks
}

// チーム削除
func DeleteTeam(blockActions map[string]map[string]slack.BlockAction, teamName string) (blocks []slack.Block) {
	var ok bool

	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
	} else {
		delete(td, teamName)

		if err = td.Update(); err != nil {
			blocks = td.GetErrBlocks(err, DataUpdateErr)
		} else {
			if err := td.SynchronizeMember(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
			} else {
				text := post.ScsText(fmt.Sprintf("チーム `%s` の削除に成功しました", teamName))
				blocks, ok = post.SingleTextBlock(text), true

				Logger.Println("チーム削除に成功しました")
			}
		}
	}

	if !ok {
		Logger.Println("チーム削除に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
