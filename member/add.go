// メンバー管理
package member

import (
	"fmt"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メンバー追加リクエスト
func getBlockAdd() []slack.Block {
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

	// ブロック: ヘッダ
	headerText := post.InfoText("labotGo にメンバーを追加します\n\n")
	headerText += "*追加したいユーザと所属チームを選択してください*"
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{util.TipsMemberTeam,
		fmt.Sprintf("`%s team add` で追加したチームを選択できます（ `%s member edit` で後から変更可能）", util.Cmd, util.Cmd),
		"チーム `all` は全メンバーが入るチームです．削除しないでください．",
	}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: ユーザ選択
	userSelectSection := post.SelectMembersSection(data.GetAllNonMembers(md), aid.AddMemberSelectUser, []string{}, false, false)

	// ブロック: チーム選択
	teamsSelectSection := post.SelectTeamsSection(td.GetAllNames(), aid.AddMemberSelectTeams, []string{util.MasterTeamName}, true)

	// ブロック: 追加ボタン
	actionBtnBlock := post.CustomBtnSection("OK", "追加", aid.AddMember)

	blocks := []slack.Block{
		headerSection, headerTipsSection, util.Divider(), userSelectSection, teamsSelectSection, actionBtnBlock,
	}
	return blocks
}

// メンバー追加
func AddMember(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var ok bool
	util.Logger.Printf("メンバー追加リクエスト (from:%s): %+v\n", actionUserID, blockActions)

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		var (
			userID    string
			teamNames []string
		)
		// ユーザID・所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.AddMemberSelectUser:
					userID = data.RawUserID(values.SelectedOption.Value)
				case aid.AddMemberSelectTeams:
					for _, opt := range values.SelectedOptions {
						teamNames = append(teamNames, opt.Value)
					}
				default:
				}
			}
		}
		util.Logger.Printf("ユーザID: %s / 所属チーム: %v\n", userID, teamNames)

		// バリデーションチェック
		isEmptyUserID, isEmptyTeams := (userID == ""), len(teamNames) == 0
		if isEmptyUserID && isEmptyTeams {
			blocks = post.SingleTextBlock(post.ErrText("ユーザ／チームともに選択されていません"))
		} else if isEmptyUserID {
			blocks = post.SingleTextBlock(post.ErrText("登録したいユーザが指定されていません"))
		} else if isEmptyTeams {
			headerText := post.ErrText("所属チームは1つ以上選択してください")
			headerSection := post.SingleTextSectionBlock(util.PlainText, headerText)
			tipsSection := post.TipsSection([]string{util.TipsMasterTeam})
			blocks = []slack.Block{headerSection, tipsSection}
		} else if util.ListContains(md.GetAllUserIDs(), userID) {
			blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf("ユーザ <@%s> は既にメンバーに追加されています", userID)))
		} else if !util.ListContains(teamNames, util.MasterTeamName) {
			blocks = post.SingleTextBlock(post.ErrText(util.TipsMasterTeam))
		} else {
			// メンバーデータ 更新
			md.Add(userID, teamNames, actionUserID)

			if err = md.Reload(); err != nil {
				blocks = post.ErrBlocksMembersData(err, util.DataReloadErr)
			} else {
				if err := md.SynchronizeTeam(); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*以下ユーザのメンバー追加に成功しました*")
					headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
					memberInfoSections := post.InfoMemberSection(md[userID].Image24, userID, teamNames, teamNames, nil)
					tipsText := []string{"続けてメンバーを追加したい場合，同じフォームを再利用できます"}
					tipsSection := post.TipsSection(tipsText)

					blocks, ok = []slack.Block{headerSection, memberInfoSections[0], tipsSection}, true
					util.Logger.Println("メンバー編集に成功しました")
				}
			}
		}
	}

	if !ok {
		util.Logger.Println("メンバー追加に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
