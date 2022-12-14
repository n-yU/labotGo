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

// チーム追加リクエスト
func getBlocksAdd() (blocks []slack.Block) {
	var (
		md  data.MembersData
		err error
	)

	// メンバーデータ 読み込み
	if md, err = data.ReadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataReadErr)
		return blocks
	}

	// ブロック: ヘッダ
	headerText := post.InfoText("labotGo にチームを追加します\n\n")
	headerText += "*チーム名と所属メンバーを選択してください*"
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsText := []string{util.TipsMemberTeam,
		fmt.Sprintf("所属メンバーは `%s team edit` で後から変更できます（ `%s member edit` でも可能）", util.Cmd, util.Cmd),
	}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: チーム名入力
	nameSection := post.InputTeamNameSection(aid.AddTeamInputName, "")

	// ブロック: 所属メンバー選択
	membersSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.AddTeamSelectMembers, []string{}, true, true)

	// ブロック: 追加ボタン
	actionBtnBlock := post.CustomBtnSection("OK", "追加", aid.AddTeam)

	blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), nameSection, membersSection, actionBtnBlock}
	return blocks
}

// チーム追加
func AddMember(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var (
		td          data.TeamsData
		md          data.MembersData
		err         error
		ok          bool
		teamName    string
		teamUserIDs []string
	)
	util.Logger.Printf("チーム追加リクエスト (from %s): %+v\n", actionUserID, blockActions)

	// チーム・メンバー データ 読み込み
	if td, err = data.ReadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataReadErr)
	}
	if md, err = data.ReadMember(); err != nil {
		return post.ErrBlocksMembersData(err, util.DataReadErr)
	}

	// ユーザID・所属チーム 取得
	for _, action := range blockActions {
		for actionId, values := range action {
			switch actionId {
			case aid.AddTeamInputName:
				teamName = values.Value
			case aid.AddTeamSelectMembers:
				for _, uID := range values.SelectedOptions {
					teamUserIDs = append(teamUserIDs, data.RawUserID(string(uID.Value)))
				}
			default:
			}
		}
	}
	util.Logger.Printf("チーム名: %s / 所属メンバー: %v\n", teamName, teamUserIDs)

	// バリデーションチェック
	if teamName == "" {
		blocks = post.SingleTextBlock(post.ErrText("チーム名が入力されていません"))
	} else if idx := strings.Index(teamName, " "); idx >= 0 {
		text := post.ErrText(fmt.Sprintf("チーム名に半角スペースを含めることはできません（%d文字目）", idx+1))
		blocks = post.SingleTextBlock(text)
	} else if idx := strings.Index(teamName, ","); idx >= 0 {
		text := post.ErrText(fmt.Sprintf("チーム名に半角カンマを含めることはできません（%d文字目）", idx+1))
		blocks = post.SingleTextBlock(text)
	} else if util.ListContains(td.GetAllNames(), teamName) {
		headerText := post.ErrText(fmt.Sprintf("指定したチーム名 `%s` は既に存在するため追加できません", teamName))
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		tipsText := []string{fmt.Sprintf("チームの一覧を確認するには `%s team list` を実行してください", util.Cmd)}
		tipsSection := post.TipsSection(tipsText)
		blocks = []slack.Block{headerSection, tipsSection}
	} else if util.ListContains(teamUserIDs, util.MasterUserID) {
		blocks = post.SingleTextBlock(post.ErrText(util.TipsMasterUser()))
	} else {
		// チームデータ更新
		td.Add(teamName, teamUserIDs, actionUserID)

		if err = td.Write(); err != nil {
			blocks = post.ErrBlocksTeamsData(err, util.DataWriteErr)
		} else {
			if err := td.SynchronizeMember(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
			} else {
				headerText := post.ScsText("*以下チームの追加に成功しました*")
				headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
				profImages := md.GetProfImages(teamUserIDs)
				teamInfoSections := post.InfoTeamSections(teamName, teamName, profImages, teamUserIDs, []string{}, nil)
				tipsText := []string{"続けてチームを追加したい場合，同じフォームを再利用できます"}
				tipsSection := post.TipsSection(tipsText)

				blocks = []slack.Block{headerSection}
				for _, teamInfoSec := range teamInfoSections {
					blocks = append(blocks, teamInfoSec)
				}
				blocks, ok = append(blocks, tipsSection), true
				util.Logger.Println("チーム追加に成功しました")
			}
		}
	}

	if !ok {
		util.Logger.Println("チーム追加に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
