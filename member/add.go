// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// Block Kit: メンバー追加リクエスト
func getBlockAdd() []slack.Block {
	// ヘッダー
	headerText := post.InfoText("labotGo にメンバーを追加します\n\n")
	headerText += "*追加したいユーザと所属チームを選択してください*"
	headerSection := slack.NewSectionBlock(slack.NewTextBlockObject(Markdown, headerText, false, false), nil, nil)

	// ヘッダー Tips
	headerTipsText := []string{
		"*labotGo* に追加されたユーザを `メンバー` とし， `メンバー` のグループを `チーム` と呼びます",
		fmt.Sprintf("`%s team add` で追加したチームを選択できます（ `%s member edit` で後から変更可能）", Cmd, Cmd),
	}
	headerTipsSection := post.CreateTipsSection(headerTipsText)

	// ユーザ選択
	// ref.) https://stackoverflow.com/questions/67975904/slack-block-kit-multi-users-select-remove-defaults-app
	userOptionText := slack.NewTextBlockObject(PlainText, "ユーザを選択", false, false)
	userOption := &slack.SelectBlockElement{
		Type: slack.OptTypeConversations, Placeholder: userOptionText, ActionID: "memberAddSelectUser",
		Filter: &slack.SelectBlockElementFilter{Include: []string{"im"}, ExcludeBotUsers: true},
	}
	userSelectText := slack.NewTextBlockObject(Markdown, "*ユーザ*", false, false)
	userSelectSection := slack.NewSectionBlock(userSelectText, nil, slack.NewAccessory(userOption))

	// チーム選択
	teamOptions := post.CreateOptionBlockObject([]string{"B4", "M1", "M2"})
	teamsSelectOptionText := slack.NewTextBlockObject(PlainText, "チームを選択", false, false)
	teamsSelectOption := slack.NewOptionsMultiSelectBlockElement(
		slack.MultiOptTypeStatic, teamsSelectOptionText, "memberAddSelectTeams", teamOptions...,
	)
	teamsSelectText := slack.NewTextBlockObject(Markdown, "*チーム*", false, false)
	teamsSelectSection := slack.NewSectionBlock(teamsSelectText, nil, slack.NewAccessory(teamsSelectOption))

	// 追加ボタン
	addBtnText := slack.NewTextBlockObject(PlainText, "追加", false, false)
	addBtn := post.NewButtonBlockElementWithStyle("memberAdd", "", addBtnText, slack.StylePrimary)
	actionBtnBlock := slack.NewActionBlock("", addBtn)

	blocks := []slack.Block{
		headerSection, headerTipsSection, slack.NewDividerBlock(),
		userSelectSection, teamsSelectSection, actionBtnBlock,
	}
	return blocks
}

// メンバー追加
func AddMember(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var (
		userId       string
		teams        []string
		ok           bool
		data         map[string]interface{}
		dataTipsText = []string{
			fmt.Sprintf("データファイル `%s` が存在しないか，ファイル／データ形式が不適切です\n", MemberDataPath),
			"データファイルを削除した上で，botを再起動すると解消されます（但しデータはリセットされます）",
		}
	)

	Logger.Printf("メンバー追加リクエスト: %+v\n", blockActions)

	// メンバーデータ 読み込み
	data, err := LoadData()
	if err != nil {
		text := "メンバーデータの読み込みに失敗しました"
		headerSection := slack.NewSectionBlock(
			slack.NewTextBlockObject(PlainText, post.ErrText(text), false, false), nil, nil,
		)
		tipsSection := post.CreateTipsSection(dataTipsText)
		blocks = []slack.Block{headerSection, tipsSection}

		Logger.Println(text)
		Logger.Println(err)
	} else {
		// ユーザID・所属チーム 取得
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case "memberAddSelectUser":
					userId = values.SelectedConversation
				case "memberAddSelectTeams":
					for _, opt := range values.SelectedOptions {
						teams = append(teams, opt.Value)
					}
				default:
				}
			}
		}
		Logger.Printf("ユーザID: %s / チームリスト: %v\n", userId, teams)

		// バリデーションチェック
		isEmptyUserId, isEmptyTeams := (userId == ""), len(teams) == 0
		if isEmptyUserId && isEmptyTeams {
			text := post.ErrText("ユーザ／チームともに選択されていません")
			blocks = post.CreateSingleTextBlock(text)
		} else if isEmptyUserId {
			text := post.ErrText("登録したいユーザが指定されていません")
			blocks = post.CreateSingleTextBlock(text)
		} else if isEmptyTeams {
			headerText := post.ErrText("所属チームは1つ以上選択してください")
			headerSection := slack.NewSectionBlock(
				slack.NewTextBlockObject(PlainText, headerText, false, false), nil, nil,
			)
			tipsText := []string{"チーム `all` は全メンバーが入るチームです．削除する必要はありません．"}
			tipsSection := post.CreateTipsSection(tipsText)
			blocks = []slack.Block{headerSection, tipsSection}
		} else {
			// メンバーデータ 更新
			data[userId] = teams

			if err = UpdateData(data); err != nil {
				text := "メンバーデータの更新に失敗しました"
				headerSection := slack.NewSectionBlock(
					slack.NewTextBlockObject(PlainText, post.ErrText(text), false, false), nil, nil,
				)
				tipsSection := post.CreateTipsSection(dataTipsText)
				blocks = []slack.Block{headerSection, tipsSection}

				Logger.Println(text)
				Logger.Println(err)
			} else {
				headerText := post.ScsText("*以下ユーザのメンバー追加に成功しました*")
				headerSection := slack.NewSectionBlock(
					slack.NewTextBlockObject(Markdown, headerText, false, false), nil, nil,
				)
				memberInfoUserId := slack.NewTextBlockObject(
					Markdown, fmt.Sprintf("*ユーザ*:\n<@%s>", userId), false, false,
				)
				memberInfoTeams := slack.NewTextBlockObject(
					Markdown, fmt.Sprintf("*チーム*:\n%s", strings.Join(teams, ", ")), false, false,
				)
				memberInfoField := []*slack.TextBlockObject{memberInfoUserId, memberInfoTeams}
				memberInfoSection := slack.NewSectionBlock(nil, memberInfoField, nil)

				tipsText := []string{"続けてメンバーを追加したい場合，同じフォームを再利用できます"}
				tipsSection := post.CreateTipsSection(tipsText)
				blocks, ok = []slack.Block{headerSection, memberInfoSection, tipsSection}, true
			}
		}
	}

	if ok {
		Logger.Println("メンバー追加に成功しました")
	} else {
		Logger.Println("メンバー追加に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}

	return blocks
}
