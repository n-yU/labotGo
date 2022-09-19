// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メンバー編集リクエスト（メンバー選択）
func getBlockEditMemberSelect() (blocks []slack.Block) {
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*編集したいメンバーを選択してください*")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

		// ブロック: ヘッダ Tips
		headerTipsText := []string{fmt.Sprintf("メンバーを追加する場合は `%s member add` を実行してください", util.Cmd)}
		headerTipsSection := post.TipsSection(headerTipsText)

		// ブロック: メンバー選択
		memberSelectSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.EditMemberSelectMember, []string{}, false, true)

		blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), memberSelectSection}
	}
	return blocks
}

// メンバー編集リクエスト（チーム選択）
func getBlockEditTeamsSelect(actionUserID string, blockActions map[string]map[string]slack.BlockAction) []slack.Block {
	var (
		md     data.MembersData
		td     data.TeamsData
		err    error
		userID string
	)

	// メンバー・チームデータ 読み込み
	if md, err = data.LoadMember(); err != nil {
		return post.ErrBlocksMembersData(err, util.DataLoadErr)
	}
	if td, err = data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	}

	// ユーザID・変更前チームリスト 取得
	for _, action := range blockActions {
		for actionId, values := range action {
			switch actionId {
			case aid.EditMemberSelectMember:
				userID = data.RawUserID(values.SelectedOption.Value)
			default:
			}
		}
	}
	memberTeamNames := md[userID].TeamNames
	util.Logger.Printf("ユーザID: %s / 変更前チームリスト: %v\n", userID, memberTeamNames)

	// ブロック: ヘッダ
	headerText := post.InfoText(fmt.Sprintf("*指定したメンバー <@%s> のチームを選択してください*", userID))
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
	// ブロック: ヘッダ Tips
	headerTipsText := []string{"`all` は全メンバーが入るチームのため削除できません"}
	headerTipsSection := post.TipsSection(headerTipsText)
	// ブロック: チーム選択
	teamSelectSection := post.SelectTeamsSection(td.GetAllNames(), aid.EditMemberSelectTeams, memberTeamNames, true)
	// ブロック: 変更ボタン
	actionBtnActionId := strings.Join([]string{aid.EditMember, userID}, "_")
	actionBtnBlock := post.CustomBtnSection("OK", "変更", actionBtnActionId)

	blocks := []slack.Block{headerSection, headerTipsSection, util.Divider(), teamSelectSection, actionBtnBlock}
	return blocks
}

// メンバー編集
func EditMember(actionUserID string, blockActions map[string]map[string]slack.BlockAction, userID string) (blocks []slack.Block) {
	var ok bool
	util.Logger.Printf("メンバー編集リクエスト (from:%s): %+v\n", actionUserID, blockActions)

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		var newTeamNames []string
		// 所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.EditMemberSelectTeams:
					for _, opt := range values.SelectedOptions {
						newTeamNames = append(newTeamNames, opt.Value)
					}
				default:
				}
			}
		}
		util.Logger.Printf("ユーザID: %s / チームリスト: %v\n", userID, newTeamNames)

		// バリデーションチェック
		if userID == util.MasterUserID {
			blocks = post.SingleTextBlock(post.ErrText(util.TipsMasterUser()))
		} else if len(newTeamNames) == 0 {
			headerText := post.ErrText("所属チームは1つ以上選択してください")
			headerSection := post.SingleTextSectionBlock(util.PlainText, headerText)
			tipsText := []string{util.TipsMasterTeam}
			tipsSection := post.TipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else if !util.ListContains(newTeamNames, util.MasterTeamName) {
			blocks = post.SingleTextBlock(post.ErrText(util.TipsMasterTeam))
		} else {
			// メンバーデータ 更新
			oldTeamNames := md.Update(userID, newTeamNames, actionUserID)

			if err = md.Reload(); err != nil {
				blocks = post.ErrBlocksMembersData(err, util.DataReloadErr)
			} else {
				if err := md.SynchronizeTeam(); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*メンバー情報を以下のように変更しました*")
					headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
					memberInfoSections := post.InfoMemberSection(md[userID].Image24, userID, newTeamNames, oldTeamNames, nil)
					blocks, ok = []slack.Block{headerSection, memberInfoSections[0]}, true

					util.Logger.Println("メンバー編集に成功しました")
				}
			}
		}
	}

	if !ok {
		util.Logger.Println("メンバー編集に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
