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

// チーム編集リクエスト（チーム選択）
func getBlockEditTeamSelect() []slack.Block {
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*編集したいチームを選択してください*")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTipsText := []string{
			fmt.Sprintf("チームを追加する場合は `%s team add` を実行してください", util.Cmd), "チーム `all` の編集・削除はできません",
		}
		headerTipsSection := post.TipsSection(headerTipsText)
		// ブロック: チーム選択
		teamSelectSection := post.SelectTeamsSection(td.GetAllEditedNames(), aid.EditTeamSelectName, []string{}, false)

		blocks := []slack.Block{headerSection, headerTipsSection, util.Divider(), teamSelectSection}
		return blocks
	}
}

// チーム編集リクエスト（チーム名入力・メンバー選択）
func getBlockEditTeamInfo(actionUserID string, blockActions map[string]map[string]slack.BlockAction) []slack.Block {
	var (
		md  data.MembersData
		td  data.TeamsData
		err error
	)

	// メンバー・チームデータ 読み込み
	if md, err = data.LoadMember(); err != nil {
		return post.ErrBlocksMembersData(err, util.DataLoadErr)
	}
	if td, err = data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	}

	// 変更前チーム名／メンバーリスト 取得
	var teamName string
	for _, action := range blockActions {
		for actionId, values := range action {
			switch actionId {
			case aid.EditTeamSelectName:
				teamName = values.SelectedOption.Value
			default:
			}
		}
	}
	teamUserIDs := td[teamName].UserIDs
	util.Logger.Printf("チーム名: %s / 変更前メンバーリスト: %v\n", teamName, teamUserIDs)

	// ブロック: ヘッダ
	headerText := post.InfoText(fmt.Sprintf("*指定したチーム `%s` の情報を編集してください*", teamName))
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
	// ブロック: ヘッダ Tips
	headerTipsText := []string{fmt.Sprintf("既存のチーム名への変更はできません（ `%s team list` で確認可能）", util.Cmd)}
	headerTipsSection := post.TipsSection(headerTipsText)
	// ブロック: チーム名入力
	nameSection := post.InputTeamNameSection(aid.EditTeamInputName, teamName)
	// ブロック: 所属メンバー選択
	membersSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.EditTeamSelectMembers, teamUserIDs, true, true)
	// ブロック: 変更ボタン
	actionBtnActionId := strings.Join([]string{aid.EditTeam, teamName}, "_")
	actionBtnBlock := post.BtnOK("変更", actionBtnActionId)

	blocks := []slack.Block{headerSection, headerTipsSection, util.Divider(), nameSection, membersSection, actionBtnBlock}
	return blocks
}

// チーム編集
func EditTeam(actionUserID string, blockActions map[string]map[string]slack.BlockAction, oldTeamName string) (blocks []slack.Block) {
	var (
		td          data.TeamsData
		md          data.MembersData
		err         error
		ok          bool
		newTeamName string
		newUserIDs  []string
	)
	util.Logger.Printf("チーム編集リクエスト (from %s): %+v\n", actionUserID, blockActions)

	// チーム・メンバー データ 読み込み
	if td, err = data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	}
	if md, err = data.LoadMember(); err != nil {
		return post.ErrBlocksMembersData(err, util.DataLoadErr)
	}

	// チーム名・所属メンバー 取得
	for _, action := range blockActions {
		for actionId, values := range action {
			switch actionId {
			case aid.EditTeamInputName:
				newTeamName = values.Value
			case aid.EditTeamSelectMembers:
				for _, mID := range values.SelectedOptions {
					newUserIDs = append(newUserIDs, data.RawUserID(string(mID.Value)))
				}
			default:
			}
		}
	}
	util.Logger.Printf("チーム名: %s → %s / メンバーリスト: %v\n", oldTeamName, newTeamName, newUserIDs)

	// バリデーションチェック
	if newTeamName == "" {
		blocks = post.SingleTextBlock(post.ErrText("チーム名が入力されていません"))
	} else if idx := strings.Index(newTeamName, " "); idx >= 0 {
		text := post.ErrText(fmt.Sprintf("チーム名に半角スペースを含めることはできません（%d文字目）", idx+1))
		blocks = post.SingleTextBlock(text)
	} else if idx := strings.Index(newTeamName, ","); idx >= 0 {
		text := post.ErrText(fmt.Sprintf("チーム名に半角カンマを含めることはできません（%d文字目）", idx+1))
		blocks = post.SingleTextBlock(text)
	} else if newTeamName != oldTeamName && util.ListContains(td.GetAllNames(), newTeamName) {
		headerText := post.ErrText(fmt.Sprintf("新しいチーム名 `%s` は既に存在するため変更できません", newTeamName))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{fmt.Sprintf("チームの一覧を確認するには `%s team list` を実行してください", util.Cmd)}
		tipsSection := post.TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection}
	} else {
		// チームデータ更新
		oldUserIDs := td.Update(oldTeamName, newTeamName, newUserIDs, actionUserID)

		if err = td.Reload(); err != nil {
			blocks = post.ErrBlocksTeamsData(err, util.DataReloadErr)
		} else {
			if err := td.SynchronizeMember(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
			} else {
				headerText := post.ScsText("*チーム情報を以下のように変更しました*")
				headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
				tipsText := []string{"指定したチームかつチーム名を変更していない限り，上記フォームを再利用できます"}
				tipsSection := post.TipsSection(tipsText)
				profImages := md.GetProfImages(util.UniqueConcatSlice(newUserIDs, oldUserIDs))
				teamInfoSections := post.InfoTeamSections(newTeamName, oldTeamName, profImages, newUserIDs, oldUserIDs, nil)

				blocks = []slack.Block{headerSection}
				for _, teamInfoSec := range teamInfoSections {
					blocks = append(blocks, teamInfoSec)
				}
				blocks, ok = append(blocks, tipsSection), true
				util.Logger.Println("チーム編集に成功しました")
			}
		}
	}

	if !ok {
		util.Logger.Println("チーム編集に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
