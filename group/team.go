// 機能: メンバーグルーピング
package group

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// グルーピングリクエスト: チームメンバーモード
func getBlocksTeam() []slack.Block {
	// チームデータ 読み込み
	if td, err := data.ReadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataReadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("指定したチームのメンバーをグルーピングします\n\n")
		headerText += "*チーム名とグルーピング設定を指定してください*"
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

		// ブロック: ヘッダ Tips
		headerTipsText := []string{
			util.TipsTeamList(), "チームを複数選択した場合，複合チームでグルーピングされます",
			"グループ *サイズ* 指定時の注意: `/lab help group-size`",
		}
		headerTipsSection := post.TipsSection(headerTipsText)

		// ブロック: チームリスト選択
		teamsSelectSection := post.SelectTeamsSection(td.GetAllNames(), aid.GroupTeamSelectNames, []string{}, true)

		// ブロック: グルーピングタイプ選択
		typeSelectSection := TypeSelectSection(aid.GroupTeamSelectType)

		// ブロック: グルーピングバリュー入力
		valueInputSection := ValueInputSection(aid.GroupTeamInputValue)

		// ブロック: グルーピングボタン
		actionBtnBlock := post.CustomBtnSection("OK", "グルーピング", aid.GroupTeam)

		blocks := []slack.Block{
			headerSection, headerTipsSection, util.Divider(),
			teamsSelectSection, typeSelectSection, valueInputSection, actionBtnBlock,
		}
		return blocks
	}
}

// チームメンバーグルーピング
func GroupTeam(
	actionUserID string, blockActions map[string]map[string]slack.BlockAction,
) (blocks []slack.Block, responseType string) {
	var (
		md         data.MembersData
		td         data.TeamsData
		err        error
		teamNames  []string
		groupType  string
		groupValue string
		ok         bool
	)
	util.Logger.Printf("チームメンバーグルーピングリクエスト (from %s): %+v\n", actionUserID, blockActions)

	// メンバー・チームデータ 読み込み
	if md, err = data.ReadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataReadErr)
		return blocks, util.Ephemeral
	}
	if td, err = data.ReadTeam(); err != nil {
		blocks = post.ErrBlocksTeamsData(err, util.DataReadErr)
		return blocks, util.Ephemeral
	}

	// チームリスト・グルーピング設定 取得
	for _, action := range blockActions {
		for actionID, values := range action {
			switch actionID {
			case aid.GroupTeamSelectNames:
				for _, opt := range values.SelectedOptions {
					teamNames = append(teamNames, opt.Value)
				}
			case aid.GroupTeamSelectType:
				groupType = values.SelectedOption.Value
			case aid.GroupTeamInputValue:
				groupValue = values.Value
			}
		}
	}
	util.Logger.Printf("チームリスト: %v / グルーピングタイプ: %s / グループングバリュー: %s\n", teamNames, groupType, groupValue)

	// 欠損値チェック
	blocks = CheckMissingValue(teamNames, groupType, groupValue, false)
	if len(blocks) > 0 {
		return blocks, util.Ephemeral
	}

	// バリデーションチェック
	if groupType == util.GroupTypeOptionNum {
		// グループサイズ指定時のフォーマットへの統一
		groupValue += "_"
	}
	groupValueInt, groupValueErr := strconv.Atoi(groupValue[:len(groupValue)-1])
	groupValueOption := string(groupValue[len(groupValue)-1])
	memberUserIDs, _ := td.GetComplexTeamMemberUserIDs(teamNames)
	teamNamesString := strings.Join(teamNames, " + ")

	if len(memberUserIDs) == 0 {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"指定したチーム `%s` のメンバーが 0人 のため，グルーピングできません", teamNamesString,
		)))
	} else if groupType == util.GroupTypeOptionSize && groupValueOption != "+" && groupValueOption != "-" {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"グループサイズをタイプ指定する場合，グループサイズは `自然数+` or `自然数-` のフォーマットにする必要があります",
		)))
	} else if groupType == util.GroupTypeOptionNum && (strings.Contains(groupValue, "+") || strings.Contains(groupValue, "-")) {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"グループ数をタイプ指定する場合，グループサイズに `+` / `-` を含める必要はありません",
		)))
	} else if groupValueErr != nil || groupValueInt <= 0 {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"指定したグループ数・グループサイズ `%d` は自然数ではありません", groupValueInt,
		)))
	} else if groupValueInt > len(memberUserIDs) {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"グループ数・グループサイズ `%d` は指定グループのメンバー数 `%d` を超えています", groupValueInt, len(memberUserIDs),
		)))
	} else {
		blocks, ok = GroupBlocks(memberUserIDs, groupType, groupValueInt, groupValueOption, teamNamesString, md)
	}

	if ok {
		util.Logger.Println("メンバーグルーピング（チームメンバーモード）に成功しました")
		responseType = util.InChannel
	} else {
		util.Logger.Println("メンバーグルーピング（チームメンバーモード）に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
		responseType = util.Ephemeral
	}
	return blocks, responseType
}
