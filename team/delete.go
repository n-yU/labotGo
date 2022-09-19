// チーム管理
package team

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// チーム削除リクエスト（チーム選択）
func getBlockDeleteTeamSelect() (blocks []slack.Block) {
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*削除したいチームを選択してください*")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTipsText := []string{"チームを選択すると確認のメッセージが表示されます"}
		headerTipsSection := post.TipsSection(headerTipsText)
		// ブロック: チーム選択
		teamSelectSection := post.SelectTeamsSection(td.GetAllEditedNames(), aid.DeleteTeamSelectTeam, []string{}, false)

		blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), teamSelectSection}
	}
	return blocks
}

// チーム削除リクエスト（確認）
func DeleteTeamConfirm(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var (
		td  data.TeamsData
		md  data.MembersData
		err error
	)
	util.Logger.Printf("チーム削除リクエスト (from %s):%+v", actionUserID, blockActions)

	// チーム・メンバーデータ 読み込み
	if td, err = data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	}
	if md, err = data.LoadMember(); err != nil {
		return post.ErrBlocksMembersData(err, util.DataLoadErr)
	}

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
	util.Logger.Printf("チーム名: %s / メンバーリスト: %v\n", teamName, teamUserIDs)

	headerSection := post.SingleTextSectionBlock(util.Markdown, "*以下チームを削除しますか？*")
	profImages := md.GetProfImages(teamUserIDs)
	teamInfoSections := post.InfoTeamSections(teamName, teamName, profImages, teamUserIDs, teamUserIDs, nil)
	actionBtnActionId := strings.Join([]string{aid.DeleteTeam, teamName}, "_")
	actionBtnBlock := post.CustomBtnSection("OK", "削除", actionBtnActionId)

	blocks = []slack.Block{headerSection}
	for _, teamInfoSec := range teamInfoSections {
		blocks = append(blocks, teamInfoSec)
	}
	blocks = append(blocks, actionBtnBlock)
	return blocks
}

// チーム削除
func DeleteTeam(actionUserID string, blockActions map[string]map[string]slack.BlockAction, teamName string) (blocks []slack.Block) {
	var ok bool

	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// チーム削除
		td.Delete(teamName)

		if err = td.Reload(); err != nil {
			blocks = post.ErrBlocksTeamsData(err, util.DataReloadErr)
		} else {
			if err := td.SynchronizeMember(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
			} else {
				text := post.ScsText(fmt.Sprintf("チーム `%s` の削除に成功しました", teamName))
				blocks, ok = post.SingleTextBlock(text), true

				util.Logger.Println("チーム削除に成功しました")
			}
		}
	}

	if !ok {
		util.Logger.Println("チーム削除に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
