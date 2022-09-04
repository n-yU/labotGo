// メッセージ投稿
package post

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型セクション: メンバー選択
func SelectMembersSection(userIDs []string, actionID string, initUserIDs []string, isMulti, isMember bool) *slack.SectionBlock {
	var selectOptionType, text string
	if isMulti {
		selectOptionType = slack.MultiOptTypeStatic
	} else {
		selectOptionType = slack.OptTypeStatic
	}
	if isMember {
		text = "メンバー"
	} else {
		text = "ユーザ"
	}

	options, initOptions := OptionBlockObjectList(userIDs, true), OptionBlockObjectList(initUserIDs, true)
	selectOptionText := TxtBlockObj(util.PlainText, fmt.Sprintf("%sを選択", text))
	selectOption := &slack.MultiSelectBlockElement{
		Type: selectOptionType, Placeholder: selectOptionText, ActionID: actionID, Options: options, InitialOptions: initOptions,
	}
	selectText := TxtBlockObj(util.Markdown, fmt.Sprintf("*%s*", text))
	selectSection := slack.NewSectionBlock(selectText, nil, slack.NewAccessory(selectOption))
	return selectSection
}

// 定型セクション: メンバー情報
func InfoMemberSection(
	profImage, userID string, newTeamNames, oldTeamNames []string, createdInfo *data.CreatedInfo,
) []*slack.ContextBlock {
	var teamNamesField *slack.TextBlockObject

	userInfoObject := TxtBlockObj(util.Markdown, "*ユーザ*:")
	profImageObject := slack.NewImageBlockElement(profImage, userID)
	userIDObject := TxtBlockObj(util.Markdown, fmt.Sprintf("<@%s>", userID))

	infoTeamsTextList := []string{}
	for _, teamName := range util.UniqueConcatSlice(oldTeamNames, newTeamNames) {
		var teamNameText string
		if isOld, isNew := util.ListContains(oldTeamNames, teamName), util.ListContains(newTeamNames, teamName); isOld && isNew {
			teamNameText = teamName
		} else if isOld {
			teamNameText = fmt.Sprintf("~%s~", teamName)
		} else if isNew {
			teamNameText = fmt.Sprintf("*%s*", teamName)
		}
		infoTeamsTextList = append(infoTeamsTextList, teamNameText)
	}
	teamInfoObject := TxtBlockObj(util.Markdown, "*チーム*:")
	teamNamesField = TxtBlockObj(util.Markdown, fmt.Sprintf("%s", strings.Join(infoTeamsTextList, ", ")))

	elements := []slack.MixedElement{
		userInfoObject, profImageObject, userIDObject, teamInfoObject, teamNamesField,
	}
	infoSections := []*slack.ContextBlock{slack.NewContextBlock("", elements...)}

	// 詳細版（listコマンド使用時）
	if createdInfo != nil {
		elements := []slack.MixedElement{
			TxtBlockObj(util.Markdown, "*作成*"), slack.NewImageBlockElement(createdInfo.Image, userID),
			TxtBlockObj(util.Markdown, fmt.Sprintf("<@%s>", createdInfo.UserID)),
			TxtBlockObj(util.Markdown, util.FormatTime(createdInfo.At)),
		}
		infoSections = append(infoSections, slack.NewContextBlock("", elements...))
	}

	return infoSections
}

// 定型ブロック: メンバーデータエラー
func ErrBlocksMembersData(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case util.DataLoadErr:
		text = "メンバーデータの読み込みに失敗しました"
	case util.DataReloadErr:
		text = "メンバーデータの更新に失敗しました"
	default:
		util.Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := SingleTextSectionBlock(util.PlainText, ErrText(text))
	tipsSection := TipsSection(TipsDataError(util.MemberDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	util.Logger.Println(text)
	util.Logger.Println(err)
	return blocks
}
