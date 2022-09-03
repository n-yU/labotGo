// 機能: メンバーグルーピング
package group

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/shuffle"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// グルーピングリクエスト: チームメンバーモード
func getBlocksTeam() []slack.Block {
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		return post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("指定したチームのメンバーをグルーピングします\n\n")
		headerText += "*チーム名とグルーピング設定を指定してください*"
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

		// ブロック: ヘッダ Tips
		headerTipsText := []string{
			util.TipsTeamList(), "チームを複数選択した場合，複合チームでグルーピングされます",
			"グループ*サイズ*指定時の注意: `/lab help group-size`",
		}
		headerTipsSection := post.TipsSection(headerTipsText)

		// ブロック: チームリスト選択
		teamsSelectSection := post.SelectTeamsSection(td.GetAllNames(), aid.GroupTeamSelectNames, []string{}, true)

		// ブロック: グルーピングタイプ選択
		typeOptions := post.OptionBlockObjectList([]string{util.GroupTypeOptionNum, util.GroupTypeOptionSize}, false)
		typeSelectOptionText := post.TxtBlockObj(util.PlainText, "グルーピングタイプを選択")
		typeSelectOption := slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic, typeSelectOptionText, aid.GroupTeamSelectType, typeOptions...,
		)
		typeSelectText := post.TxtBlockObj(util.Markdown, "*グルーピングタイプ*")
		typeSelectSection := slack.NewSectionBlock(typeSelectText, nil, slack.NewAccessory(typeSelectOption))

		// ブロック: グルーピングバリュー入力
		valueInputSectionText := post.TxtBlockObj(util.PlainText, "グループ数・グループサイズ")
		valueInputSectionHint := post.TxtBlockObj(
			util.PlainText, "指定グループのメンバー数以下の自然数を入力してください\nグループサイズを指定する場合は末尾に +/- を付けてください",
		)
		valueInputText := post.TxtBlockObj(util.PlainText, "グループ数・グループサイズを入力")
		valueInput := slack.NewPlainTextInputBlockElement(valueInputText, aid.GroupTeamInputValue)
		valueInputSection := slack.NewInputBlock("", valueInputSectionText, valueInputSectionHint, valueInput)

		// ブロック: グルーピングボタン
		actionBtnBlock := post.BtnOK("グルーピング", aid.GroupTeam)

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
	if md, err = data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
		return blocks, util.Ephemeral
	}
	if td, err = data.LoadTeam(); err != nil {
		blocks = post.ErrBlocksTeamsData(err, util.DataLoadErr)
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
	isEmptyTeamNames := (len(teamNames) == 0)
	isEmptyGroupType, isEmptyGroupValue := (groupType == ""), (groupValue == "")

	if isEmptyTeamNames || isEmptyGroupType || isEmptyGroupValue {
		emptyElements := []string{}
		if isEmptyTeamNames {
			emptyElements = append(emptyElements, "チーム")
		}
		if isEmptyGroupType {
			emptyElements = append(emptyElements, "タイプ")
		}
		if isEmptyGroupValue {
			emptyElements = append(emptyElements, "グループ数・グループサイズ")
		}
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"%s が指定されていません", strings.Join(emptyElements, "／"),
		)))
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
	memberUserIDs = util.UniqueSlice(memberUserIDs)
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
			"グループ数・グループサイズ `%d` が指定グループのメンバー数 `%d` を超えています", groupValueInt, len(memberUserIDs),
		)))
	} else {
		shuffledMemberUIDs := shuffle.ShuffleMemberUserIDs(memberUserIDs)

		switch groupType {
		case util.GroupTypeOptionNum:
			// タイプ: グループ数指定
			headerText := post.ScsText(fmt.Sprintf(
				"指定チーム *%s* を *グループ数=%d* でグルーピングしました", teamNamesString, groupValueInt,
			))
			headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
			blocks = []slack.Block{headerSection, util.Divider()}

			// 各メンバー グループ割当
			groupsUserIDs := make([][]string, groupValueInt, groupValueInt)
			for i, userID := range shuffledMemberUIDs {
				groupNo := i % groupValueInt
				groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], userID)
			}

			// グルーピング情報セクション追加
			for i, groupUserIDs := range groupsUserIDs {
				if len(groupUserIDs) == 0 {
					continue
				}
				groupResultSections := getGroupResultSections(groupUserIDs, i+1, md)
				for _, groupInfoSec := range groupResultSections {
					blocks = append(blocks, groupInfoSec)
				}
				blocks = append(blocks, util.Divider())
			}
			ok = true
		case util.GroupTypeOptionSize:
			// タイプ: グループサイズ指定
			headerText := post.ScsText(fmt.Sprintf(
				"指定チーム *%s* を *グループサイズ=%d%s* でグルーピングしました",
				teamNamesString, groupValueInt, groupValueOption,
			))
			headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
			blocks = []slack.Block{headerSection, util.Divider()}

			// 各メンバー グループ割当
			groupNum := len(shuffledMemberUIDs) / groupValueInt
			groupsUserIDs := make([][]string, groupNum, groupNum+1)
			for i, userID := range shuffledMemberUIDs[:(groupNum * groupValueInt)] {
				groupNo := i % groupNum
				groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], userID)
			}

			// 剰余メンバー グループ割当
			if groupValueOption == "+" {
				// チーム数を維持して，剰余メンバーを各チームに割り当てる
				for i := groupNum * groupValueInt; i < len(shuffledMemberUIDs); i++ {
					groupNo := i % groupValueInt
					groupsUserIDs[groupNo] = append(groupsUserIDs[groupNo], shuffledMemberUIDs[i])
				}
			} else if groupValueOption == "-" {
				// チームを1つ増やし，そのチームに剰余メンバーを全員割り当てる
				groupsUserIDs = append(groupsUserIDs, shuffledMemberUIDs[(groupNum*groupValueInt):])
			} else {
			}

			// グルーピング情報セクション追加
			for i, groupUserIDs := range groupsUserIDs {
				if len(groupUserIDs) == 0 {
					continue
				}
				groupResultSections := getGroupResultSections(groupUserIDs, i+1, md)
				for _, groupInfoSec := range groupResultSections {
					blocks = append(blocks, groupInfoSec)
				}
				blocks = append(blocks, util.Divider())
			}
			ok = true
		default:
		}
	}

	if ok {
		util.Logger.Println("メンバーグルーピングに成功しました")
		responseType = util.InChannel
	} else {
		util.Logger.Println("メンバーグルーピングに失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
		responseType = util.Ephemeral
	}
	return blocks, responseType
}
