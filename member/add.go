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

// Block Kit: メンバー追加リクエスト
func getBlockAdd() (blocks []slack.Block) {
	var (
		memberData map[string][]string
		teamData   map[string][]string
		err        error
	)

	// メンバー・チームデータ 読み込み
	if memberData, err = data.LoadMember(); err != nil {
		blocks = data.GetMemberErrBlocks(err, DataLoadErr)
		return blocks
	}
	if teamData, err = data.LoadTeam(); err != nil {
		blocks = data.GetTeamErrBlocks(err, DataLoadErr)
		return blocks
	}

	// ブロック: ヘッダー
	headerText := post.InfoText("labotGo にメンバーを追加します\n\n")
	headerText += "*追加したいユーザと所属チームを選択してください*"
	headerSection := post.SingleTextSectionBlock(Markdown, headerText)

	// ブロック: ヘッダー Tips
	headerTipsText := []string{TipsMemberTeam,
		fmt.Sprintf("`%s team add` で追加したチームを選択できます（ `%s member edit` で後から変更可能）", Cmd, Cmd),
		"チーム `all` は全メンバーが入るチームです．削除しないでください．",
	}
	headerTipsSection := post.TipsSection(headerTipsText)

	// ブロック: ユーザ選択
	userSelectOptionText := post.TxtBlockObj(PlainText, "ユーザを選択")
	userOption := post.CreateOptionBlockObject(data.GetAllNonMembers(memberData), true)
	userSelectOption := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic, userSelectOptionText, aid.AddMemberSelectUser, userOption...,
	)
	userSelectText := post.TxtBlockObj(Markdown, "*ユーザ*")
	userSelectSection := slack.NewSectionBlock(userSelectText, nil, slack.NewAccessory(userSelectOption))

	// ブロック: チーム選択
	teamsSelectSection := post.SelectTeamsSection(data.GetAllTeams(teamData), aid.AddMemberSelectTeams, []string{"all"})

	// ブロック: 追加ボタン
	actionBtnBlock := post.BtnOK("追加", aid.AddMember)

	blocks = []slack.Block{
		headerSection, headerTipsSection, Divider(), userSelectSection, teamsSelectSection, actionBtnBlock,
	}
	return blocks
}

// メンバー追加
func AddMember(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var ok bool
	Logger.Printf("メンバー追加リクエスト: %+v\n", blockActions)

	// メンバーデータ 読み込み
	if memberData, err := data.LoadMember(); err != nil {
		blocks = data.GetMemberErrBlocks(err, DataLoadErr)
	} else {
		var (
			userID string
			teams  []string
		)
		// ユーザID・所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.AddMemberSelectUser:
					userID = data.RawUserID(values.SelectedOption.Value)
				case aid.AddMemberSelectTeams:
					for _, opt := range values.SelectedOptions {
						teams = append(teams, opt.Value)
					}
				default:
				}
			}
		}
		Logger.Printf("ユーザID: %s / 所属チーム: %v\n", userID, teams)

		// バリデーションチェック
		isEmptyUserID, isEmptyTeams := (userID == ""), len(teams) == 0
		if isEmptyUserID && isEmptyTeams {
			text := post.ErrText("ユーザ／チームともに選択されていません")
			blocks = post.SingleTextBlock(text)
		} else if isEmptyUserID {
			text := post.ErrText("登録したいユーザが指定されていません")
			blocks = post.SingleTextBlock(text)
		} else if isEmptyTeams {
			headerText := post.ErrText("所属チームは1つ以上選択してください")
			headerSection := post.SingleTextSectionBlock(PlainText, headerText)
			tipsText := []string{TipsTeamALL}
			tipsSection := post.TipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else if !ListContains(teams, "all") {
			text := post.ErrText(TipsTeamALL)
			blocks = post.SingleTextBlock(text)
		} else {
			// メンバーデータ 更新
			memberData[userID] = teams

			if err = data.UpdateMember(memberData); err != nil {
				blocks = data.GetMemberErrBlocks(err, DataUpdateErr)
			} else {
				if err := data.SynchronizeTeam(memberData); err != nil {
					blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
				} else {
					headerText := post.ScsText("*以下ユーザのメンバー追加に成功しました*")
					headerSection := post.SingleTextSectionBlock(Markdown, headerText)
					memberInfoUserID := post.TxtBlockObj(Markdown, fmt.Sprintf("*ユーザ*:\n<@%s>", userID))
					memberInfoTeams := post.TxtBlockObj(Markdown, fmt.Sprintf("*チーム*:\n%s", strings.Join(teams, ", ")))
					memberInfoField := []*slack.TextBlockObject{memberInfoUserID, memberInfoTeams}
					memberInfoSection := slack.NewSectionBlock(nil, memberInfoField, nil)

					tipsText := []string{"続けてメンバーを追加したい場合，同じフォームを再利用できます"}
					tipsSection := post.TipsSection(tipsText)
					blocks, ok = []slack.Block{headerSection, memberInfoSection, tipsSection}, true

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
