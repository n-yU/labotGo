// チーム管理
package team

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// Block Kit: チーム追加リクエスト
func getBlockAdd() []slack.Block {
	// ヘッダー
	headerText := post.InfoText("labotGo にチームを追加します\n\n")
	headerText += "*チーム名と所属メンバーを選択してください*"
	headerSection := slack.NewSectionBlock(post.TxtBlockObj(Markdown, headerText), nil, nil)

	// ヘッダー Tips
	headerTipsText := []string{TipsMemberTeam,
		fmt.Sprintf("所属メンバーは `%s team edit` で後から変更できます（ `%s member edit` でも可能）", Cmd, Cmd),
	}
	headerTipsSection := post.CreateTipsSection(headerTipsText)

	// チーム名入力
	nameText := post.TxtBlockObj(PlainText, "チーム名")
	nameHint := post.TxtBlockObj(PlainText, "1〜20文字で入力してください ／ スペースは使用できません")
	nameInputText := post.TxtBlockObj(PlainText, "チーム名を入力")
	nameInput := slack.NewPlainTextInputBlockElement(nameInputText, "teamAddInputName")
	nameInput.MinLength, nameInput.MaxLength = 1, 20
	nameSection := slack.NewInputBlock("", nameText, nameHint, nameInput)

	// 所属メンバー選択
	memberText := post.TxtBlockObj(PlainText, "所属メンバー")
	memberHint := post.TxtBlockObj(PlainText, "複数のメンバーを選択できます ／ 空欄の場合はメンバー0人のチームが作成されます")
	memberInputText := post.TxtBlockObj(PlainText, "所属メンバーを選択")
	memberInput := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, memberInputText, "teamAddSelectMembers")
	memberSection := slack.NewInputBlock("", memberText, memberHint, memberInput)

	// 追加ボタン
	addBtnText := post.TxtBlockObj(PlainText, "追加")
	addBtn := post.NewButtonBlockElementWithStyle("teamAdd", "", addBtnText, slack.StylePrimary)
	actionBtnBlock := slack.NewActionBlock("", addBtn)

	blocks := []slack.Block{
		headerSection, headerTipsSection, slack.NewDividerBlock(),
		nameSection, memberSection, actionBtnBlock,
	}
	return blocks
}

// チーム追加
func AddMember(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var ok bool
	Logger.Printf("チーム追加リクエスト: %+v\n", blockActions)

	// チームデータ 読み込み
	data, err := LoadData()
	if err != nil {
		text := "チームデータの読み込みに失敗しました"
		headerSection := slack.NewSectionBlock(post.TxtBlockObj(PlainText, post.ErrText(text)), nil, nil)
		tipsSection := post.CreateTipsSection(post.TipsDataError(TeamDataPath))
		blocks = []slack.Block{headerSection, tipsSection}

		Logger.Println(text)
		Logger.Println(err)
	} else {
		var (
			teamName string
			members  []string
		)
		// ユーザID・所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case "teamAddInputName":
					teamName = values.Value
				case "teamAddSelectMembers":
					for _, userId := range values.SelectedUsers {
						members = append(members, string(userId))
					}
				default:
				}
			}
		}
		Logger.Printf("チーム名: %s / チームリスト: %v\n", teamName, members)

		// バリデーションチェック
		if teamName == "" {
			text := post.ErrText("チーム名が入力されていません")
			blocks = post.CreateSingleTextBlock(text)
		} else if idx := strings.Index(teamName, " "); idx >= 0 {
			text := post.ErrText(fmt.Sprintf("チーム名にスペースを含めることはできません（%d文字目）\n", idx+1))
			blocks = post.CreateSingleTextBlock(text)
		} else {
			// チームデータ更新
			data[teamName] = members

			if err = UpdateData(data); err != nil {
				text := "チームデータの更新に失敗しました"
				headerSection := slack.NewSectionBlock(post.TxtBlockObj(PlainText, post.ErrText(text)), nil, nil)
				tipsSection := post.CreateTipsSection(post.TipsDataError(MemberDataPath))
				blocks = []slack.Block{headerSection, tipsSection}

				Logger.Println(text)
				Logger.Println(err)
			} else {
				var teamInfoMembers *slack.TextBlockObject
				headerText := post.ScsText("*以下チームの追加に成功しました*")
				headerSection := slack.NewSectionBlock(post.TxtBlockObj(Markdown, headerText), nil, nil)
				teamInfoName := post.TxtBlockObj(Markdown, fmt.Sprintf("*チーム名*:\n%s", teamName))
				if len(members) > 0 {
					teamInfoMembers = post.TxtBlockObj(Markdown, fmt.Sprintf("*メンバー*:\n<@%s>", strings.Join(members, ">, <@")))
				} else {
					teamInfoMembers = post.TxtBlockObj(Markdown, "*メンバー*:\n所属メンバーなし")
				}
				teamInfoField := []*slack.TextBlockObject{teamInfoName, teamInfoMembers}
				teamInfoSection := slack.NewSectionBlock(nil, teamInfoField, nil)

				tipsText := []string{"続けてチームを追加したい場合，同じフォームを再利用できます"}
				tipsSection := post.CreateTipsSection(tipsText)
				blocks, ok = []slack.Block{headerSection, teamInfoSection, tipsSection}, true
			}
		}
	}

	if ok {
		Logger.Println("チーム追加に成功しました")
	} else {
		Logger.Println("チーム追加に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
