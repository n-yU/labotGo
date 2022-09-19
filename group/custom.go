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

// グルーピングリクエスト: カスタムメンバーモード
func getBlocksCustom() []slack.Block {
	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("指定したメンバーをグルーピングします\n\n")
		headerText += "*メンバーとグルーピング設定を指定してください*"
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

		// ブロック: ヘッダ Tips
		headerTipsText := []string{"グループ *サイズ* 指定時の注意: `/lab help group-size`"}
		headerTipsSection := post.TipsSection(headerTipsText)

		// ブロック: メンバー選択
		membersSelectSection := post.SelectMembersSection(md.GetAllUserIDs(), aid.GroupCustomSelectMembers, []string{}, true, true)

		// ブロック: グルーピングタイプ選択
		typeSelectSection := TypeSelectSection(aid.GroupCustomSelectType)

		// ブロック: グルーピングバリュー入力
		valueInputSection := ValueInputSection(aid.GroupCustomInputValue)

		// ブロック: グルーピングボタン
		actionBtnBlock := post.CustomBtnSection("OK", "グルーピング", aid.GroupCustom)

		blocks := []slack.Block{
			headerSection, headerTipsSection, util.Divider(),
			membersSelectSection, typeSelectSection, valueInputSection, actionBtnBlock,
		}
		return blocks
	}
}

// カスタムメンバーグルーピング
func GroupCustom(
	actionUserID string, blockActions map[string]map[string]slack.BlockAction,
) (blocks []slack.Block, responseType string) {
	var (
		md            data.MembersData
		err           error
		memberUserIDs []string
		groupType     string
		groupValue    string
		ok            bool
	)
	util.Logger.Printf("チームメンバーグルーピングリクエスト (from %s): %+v\n", actionUserID, blockActions)

	// メンバー・データ 読み込み
	if md, err = data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
		return blocks, util.Ephemeral
	}

	// メンバーリスト・グルーピング設定 取得
	for _, action := range blockActions {
		for actionID, values := range action {
			switch actionID {
			case aid.GroupCustomSelectMembers:
				for _, opt := range values.SelectedOptions {
					memberUserIDs = append(memberUserIDs, data.RawUserID(opt.Value))
				}
			case aid.GroupCustomSelectType:
				groupType = values.SelectedOption.Value
			case aid.GroupCustomInputValue:
				groupValue = values.Value
			}
		}
	}
	util.Logger.Printf("メンバーリスト: %v / グルーピングタイプ: %s / グループングバリュー: %s\n", memberUserIDs, groupType, groupValue)

	// 欠損値チェック
	blocks = CheckMissingValue(memberUserIDs, groupType, groupValue, true)
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

	if groupType == util.GroupTypeOptionSize && groupValueOption != "+" && groupValueOption != "-" {
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"グループサイズをタイプ指定する場合，グループサイズは `自然数+` / `自然数-` のフォーマットにする必要があります",
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
			"グループ数・グループサイズ `%d` は指定メンバー数 `%d` を超えています", groupValueInt, len(memberUserIDs),
		)))
	} else {
		blocks, ok = GroupBlocks(memberUserIDs, groupType, groupValueInt, groupValueOption, "", md)
	}

	if ok {
		util.Logger.Println("メンバーグルーピング（カスタムメンバーモード）に成功しました")
		responseType = util.InChannel
	} else {
		util.Logger.Println("メンバーグルーピング（カスタムメンバーモード）に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
		responseType = util.Ephemeral
	}
	return blocks, responseType
}
