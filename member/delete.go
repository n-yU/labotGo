// メンバー管理
package member

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メンバー削除リクエスト（メンバー選択）
func getBlockDeleteMemberSelect() (blocks []slack.Block) {
	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*削除したいユーザを選択してください*")
		headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTipsText := []string{"メンバーを選択すると確認のメッセージが表示されます"}
		headerTipsSection := post.TipsSection(headerTipsText)
		// ブロック: メンバー選択
		memberSelectSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.DeleteMemberSelectMember, []string{}, false, true)

		blocks = []slack.Block{headerSection, headerTipsSection, util.Divider(), memberSelectSection}
	}
	return blocks
}

// メンバー削除リクエスト（確認）
func DeleteMemberConfirm(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	util.Logger.Printf("メンバー削除リクエスト (from:%s): %+v\n", actionUserID, blockActions)

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		// ユーザID・チームリスト 取得
		var userID string
		for _, action := range blockActions {
			for actionId, values := range action {
				switch actionId {
				case aid.DeleteMemberSelectMember:
					userID = data.RawUserID(values.SelectedOption.Value)
				default:
				}
			}
		}
		memberTeamNames := md[userID].TeamNames
		util.Logger.Printf("ユーザID: %s / チームリスト: %v\n", userID, memberTeamNames)

		headerSection := post.SingleTextSectionBlock(util.Markdown, "*以下メンバーを削除しますか？*")
		memberInfoSections := post.InfoMemberSection(md[userID].Image24, userID, memberTeamNames, memberTeamNames, nil)
		actionBtnActionId := strings.Join([]string{aid.DeleteMember, userID}, "_")
		actionBtnBlock := post.BtnOK("削除", actionBtnActionId)

		blocks = []slack.Block{headerSection, memberInfoSections[0], actionBtnBlock}
	}

	return blocks
}

// メンバー削除
func DeleteMember(actionUserID string, blockActions map[string]map[string]slack.BlockAction, userID string) (blocks []slack.Block) {
	var ok bool

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = post.ErrBlocksMembersData(err, util.DataLoadErr)
	} else {
		// メンバー削除
		md.Delete(userID)

		if err = md.Reload(); err != nil {
			blocks = post.ErrBlocksMembersData(err, util.DataReloadErr)
		} else {
			if err := md.SynchronizeTeam(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(util.ErrorSynchronizeData))
			} else {
				text := post.ScsText(fmt.Sprintf("メンバー <@%s> の削除に成功しました", userID))
				blocks, ok = post.SingleTextBlock(text), true

				util.Logger.Println("メンバー削除に成功しました")
			}
		}
	}

	if !ok {
		util.Logger.Println("メンバー削除に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
