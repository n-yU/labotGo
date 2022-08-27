// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メンバー編集リクエスト（ユーザ選択）
func getBlockEditMemberSelect() (blocks []slack.Block) {
	if md, err := data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
	} else {
		// ヘッダー
		headerText := post.InfoText("*編集したいメンバーを選択してください*")
		headerSection := post.SingleTextSectionBlock(Markdown, headerText)

		// ヘッダー Tips
		headerTipsText := []string{fmt.Sprintf("メンバーを追加する場合は `%s member add` を実行してください", Cmd)}
		headerTipsSection := post.TipsSection(headerTipsText)

		// メンバー選択
		memberSelectOptionText := post.TxtBlockObj(PlainText, "メンバーを選択")
		memberOption := post.OptionBlockObjectList(md.GetAllUserIDs(), true)
		memberSelectOption := slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic, memberSelectOptionText, aid.EditMemberSelectMember, memberOption...,
		)
		memberSelectText := post.TxtBlockObj(Markdown, "*メンバー*")
		memberSelectSection := slack.NewSectionBlock(memberSelectText, nil, slack.NewAccessory(memberSelectOption))

		blocks = []slack.Block{
			headerSection, headerTipsSection, Divider(), memberSelectSection,
		}
	}
	return blocks
}

// メンバー編集リクエスト（チーム選択）
func getBlockEditTeamsSelect(blockActions map[string]map[string]slack.BlockAction) []slack.Block {
	var (
		md     data.MembersData
		td     data.TeamsData
		err    error
		userID string
	)

	// メンバー・チームデータ 読み込み
	if md, err = data.LoadMember(); err != nil {
		return md.GetErrBlocks(err, DataLoadErr)
	}
	if td, err = data.LoadTeam(); err != nil {
		return td.GetErrBlocks(err, DataLoadErr)
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
	Logger.Printf("ユーザID: %s / 変更前チームリスト: %v\n", userID, memberTeamNames)

	// ブロック: ヘッダー
	headerText := post.InfoText(fmt.Sprintf("*指定したメンバー <@%s> のチームを選択してください*\n", userID))
	headerSection := post.SingleTextSectionBlock(Markdown, headerText)
	// ブロック: ヘッダー Tips
	headerTipsText := []string{"`all` は全メンバーが入るチームのため削除できません"}
	headerTipsSection := post.TipsSection(headerTipsText)
	// ブロック: チーム選択
	teamSelectSection := post.SelectTeamsSection(td.GetAllNames(), aid.EditMemberSelectTeams, memberTeamNames)
	// ブロック: 変更ボタン
	actionBtnActionId := strings.Join([]string{aid.EditMember, userID}, "_")
	actionBtnBlock := post.BtnOK("変更", actionBtnActionId)

	blocks := []slack.Block{headerSection, headerTipsSection, Divider(), teamSelectSection, actionBtnBlock}
	return blocks
}

// メンバー編集
func EditMember(blockActions map[string]map[string]slack.BlockAction, userID string) (blocks []slack.Block) {
	var ok bool
	Logger.Printf("メンバー編集リクエスト: %+v\n", blockActions)

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
	} else {
		var newTeams []string
		// 所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.EditMemberSelectTeams:
					for _, opt := range values.SelectedOptions {
						newTeams = append(newTeams, opt.Value)
					}
				default:
				}
			}
		}
		Logger.Printf("ユーザID: %s / チームリスト: %v\n", userID, newTeams)

		// バリデーションチェック
		if userID == MasterUserID {
			blocks = post.SingleTextBlock(post.ErrText(TipsMasterUser))
		} else if len(newTeams) == 0 {
			headerText := post.ErrText("所属チームは1つ以上選択してください")
			headerSection := post.SingleTextSectionBlock(PlainText, headerText)
			tipsText := []string{TipsMasterTeam}
			tipsSection := post.TipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else if !ListContains(newTeams, MasterTeamName) {
			blocks = post.SingleTextBlock(post.ErrText(TipsMasterTeam))
		} else {
			// メンバーデータ 更新
			oldTeamNames := md[userID].TeamNames
			md[userID].TeamNames = newTeams

			if err = md.Update(); err != nil {
				blocks = md.GetErrBlocks(err, DataUpdateErr)
			} else {
				if err := md.SynchronizeTeam(); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*メンバー情報を以下のように変更しました*")
					headerSection := post.SingleTextSectionBlock(Markdown, headerText)
					memberInfoSection := post.InfoMemberSection(userID, newTeams, oldTeamNames)
					blocks, ok = []slack.Block{headerSection, memberInfoSection}, true

					Logger.Println("メンバー編集に成功しました")
				}
			}
		}
	}

	if !ok {
		Logger.Println("メンバー編集に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
