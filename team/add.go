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

// チーム追加リクエスト
func getBlockAdd() (blocks []slack.Block) {
	var (
		memberData map[string][]string
		err        error
	)

	// メンバーデータ 読み込み
	if memberData, err = data.LoadMember(); err != nil {
		blocks = data.GetTeamErrBlocks(err, DataLoadErr)
		return blocks
	}

	// ブロック: ヘッダー
	headerText := post.InfoText("labotGo にチームを追加します\n\n")
	headerText += "*チーム名と所属メンバーを選択してください*"
	headerSection := post.SingleTextSectionBlock(Markdown, headerText)

	// ブロック: ヘッダー Tips
	headerTipsText := []string{TipsMemberTeam,
		fmt.Sprintf("所属メンバーは `%s team edit` で後から変更できます（ `%s member edit` でも可能）", Cmd, Cmd),
	}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: チーム名入力
	nameSection := post.InputTeamNameSection(aid.AddTeamInputName, "")

	// ブロック: 所属メンバー選択
	memberSection := post.SelectMembersSection(data.GetAllMembers(memberData), aid.AddTeamSelectMembers, []string{})

	// ブロック: 追加ボタン
	actionBtnBlock := post.BtnOK("追加", aid.AddTeam)

	blocks = []slack.Block{
		headerSection, headerTipsSection, Divider(), nameSection, memberSection, actionBtnBlock,
	}
	return blocks
}

// チーム追加
func AddMember(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var ok bool
	Logger.Printf("チーム追加リクエスト: %+v\n", blockActions)

	// チームデータ 読み込み
	if teamData, err := data.LoadTeam(); err != nil {
		blocks = data.GetTeamErrBlocks(err, DataLoadErr)
	} else {
		var (
			teamName string
			members  []string
		)
		// ユーザID・所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.AddTeamInputName:
					teamName = values.Value
				case aid.AddTeamSelectMembers:
					for _, uId := range values.SelectedOptions {
						userID := data.RawUserID(string(uId.Value))
						members = append(members, userID)
					}
				default:
				}
			}
		}
		Logger.Printf("チーム名: %s / 所属メンバー: %v\n", teamName, members)

		// バリデーションチェック
		if teamName == "" {
			text := post.ErrText("チーム名が入力されていません")
			blocks = post.SingleTextBlock(text)
		} else if idx := strings.Index(teamName, " "); idx >= 0 {
			text := post.ErrText(fmt.Sprintf("チーム名にスペースを含めることはできません（%d文字目）\n", idx+1))
			blocks = post.SingleTextBlock(text)
		} else if ListContains(data.GetAllTeams(teamData), teamName) {
			headerText := post.ErrText(fmt.Sprintf("指定したチーム名 `%s` は既に存在するため追加できません\n", teamName))
			headerSection := post.SingleTextSectionBlock(Markdown, headerText)
			tipsText := []string{fmt.Sprintf("チームの一覧を確認するには `%s team list` を実行してください\n", Cmd)}
			tipsSection := post.TipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else {
			// チームデータ更新
			teamData[teamName] = members

			if err = data.UpdateTeam(teamData); err != nil {
				blocks = data.GetTeamErrBlocks(err, DataUpdateErr)
			} else {
				if err := data.SynchronizeMember(teamData); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*以下チームの追加に成功しました*")
					headerSection := post.SingleTextSectionBlock(Markdown, headerText)
					teamInfoSection := post.InfoTeamSection(teamName, "", members, []string{})
					tipsText := []string{"続けてチームを追加したい場合，同じフォームを再利用できます"}
					tipsSection := post.TipsSection(tipsText)
					blocks, ok = []slack.Block{headerSection, teamInfoSection, tipsSection}, true

					Logger.Println("チーム追加に成功しました")
				}
			}
		}
	}

	if !ok {
		Logger.Println("チーム追加に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
