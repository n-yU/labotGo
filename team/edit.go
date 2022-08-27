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

// チーム編集リクエスト（チーム選択）
func getBlockEditTeamSelect() (blocks []slack.Block) {
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
	} else {
		// ブロック: ヘッダー
		headerText := post.InfoText("*編集したいチームを選択してください*")
		headerSection := post.SingleTextSectionBlock(Markdown, headerText)

		// ブロック: ヘッダー Tips
		headerTipsText := []string{
			fmt.Sprintf("チームを追加する場合は `%s team add` を実行してください", Cmd),
			"チーム `all` の編集・削除はできません",
		}
		headerTipsSection := post.TipsSection(headerTipsText)

		// ブロック: チーム選択
		teamSelectOptionText := post.TxtBlockObj(PlainText, "チームを選択")
		teamOption := post.OptionBlockObjectList(td.GetAllEditedNames(), false)
		teamSelectOption := slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic, teamSelectOptionText, aid.EditTeamSelectName, teamOption...,
		)
		teamSelectText := post.TxtBlockObj(Markdown, "*チーム*")
		teamSelectSection := slack.NewSectionBlock(teamSelectText, nil, slack.NewAccessory(teamSelectOption))

		blocks = []slack.Block{
			headerSection, headerTipsSection, Divider(), teamSelectSection,
		}
	}
	return blocks
}

// チーム編集リクエスト（チーム名入力・メンバー選択）
func getBlockEditTeamInfo(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var (
		md  data.MembersData
		td  data.TeamsData
		err error
	)

	// メンバー・チームデータ 読み込み
	if md, err = data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
		return blocks
	}
	if td, err = data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
		return blocks
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
	Logger.Printf("チーム名: %s / 変更前メンバーリスト: %v\n", teamName, teamUserIDs)

	// ブロック: ヘッダー
	headerText := post.InfoText(fmt.Sprintf("*指定したチーム `%s` の情報を編集してください*", teamName))
	headerSection := post.SingleTextSectionBlock(Markdown, headerText)
	// ブロック: ヘッダー Tips
	headerTipsText := []string{fmt.Sprintf("既存のチーム名への変更はできません（ `%s team list` で確認可能）", Cmd)}
	headerTipsSection := post.TipsSection(headerTipsText)
	// ブロック: チーム名入力
	nameSection := post.InputTeamNameSection(aid.EditTeamInputName, teamName)
	// ブロック: 所属メンバー選択
	membersSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.EditTeamSelectMembers, teamUserIDs, true, true)
	// ブロック: 変更ボタン
	actionBtnActionId := strings.Join([]string{aid.EditTeam, teamName}, "_")
	actionBtnBlock := post.BtnOK("変更", actionBtnActionId)

	blocks = []slack.Block{headerSection, headerTipsSection, Divider(), nameSection, membersSection, actionBtnBlock}
	return blocks
}

// チーム編集
func EditTeam(blockActions map[string]map[string]slack.BlockAction, oldTeamName string) (blocks []slack.Block) {
	var ok bool
	Logger.Printf("チーム編集リクエスト: %+v\n", blockActions)

	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = td.GetErrBlocks(err, DataLoadErr)
	} else {
		var (
			newTeamName string
			newUserIDs  []string
		)
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
		Logger.Printf("チーム名: %s → %s / メンバーリスト: %v\n", oldTeamName, newTeamName, newUserIDs)

		// バリデーションチェック
		if newTeamName == "" {
			blocks = post.SingleTextBlock(post.ErrText("チーム名が入力されていません"))
		} else if idx := strings.Index(newTeamName, " "); idx >= 0 {
			text := post.ErrText(fmt.Sprintf("チーム名にスペースを含めることはできません（%d文字目）", idx+1))
			blocks = post.SingleTextBlock(text)
		} else if newTeamName != oldTeamName && ListContains(td.GetAllNames(), newTeamName) {
			headerText := post.ErrText(fmt.Sprintf("新しいチーム名 `%s` は既に存在するため変更できません", newTeamName))
			headerSection := post.SingleTextSectionBlock(Markdown, headerText)
			tipsText := []string{fmt.Sprintf("チームの一覧を確認するには `%s team list` を実行してください", Cmd)}
			tipsSection := post.TipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else {
			// チームデータ更新
			oldUserIDs := td[oldTeamName].UserIDs
			delete(td, oldTeamName)
			td[newTeamName] = &data.TeamData{UserIDs: newUserIDs}

			if err = td.Update(); err != nil {
				blocks = td.GetErrBlocks(err, DataUpdateErr)
			} else {
				if err := td.SynchronizeMember(); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*チーム情報を以下のように変更しました*")
					headerSection := post.SingleTextSectionBlock(Markdown, headerText)
					tipsText := []string{"指定したチームかつチーム名を変更していない限り，上記フォームを再利用できます"}
					tipsSection := post.TipsSection(tipsText)
					teamInfoSection := post.InfoTeamSection(newTeamName, oldTeamName, newUserIDs, oldUserIDs)

					blocks, ok = []slack.Block{headerSection, teamInfoSection, tipsSection}, true
					Logger.Println("チーム編集に成功しました")
				}
			}
		}
	}

	if !ok {
		Logger.Println("チーム編集に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
