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

// メンバー削除リクエスト（メンバー選択）
func getBlockDeleteMemberSelect() (blocks []slack.Block) {
	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
	} else {
		// ブロック: ヘッダ
		headerText := post.InfoText("*削除したいユーザを選択してください*")
		headerSection := post.SingleTextSectionBlock(Markdown, headerText)
		// ブロック: ヘッダ Tips
		headerTipsText := []string{"メンバーを選択すると確認のメッセージが表示されます"}
		headerTipsSection := post.TipsSection(headerTipsText)
		// ブロック: メンバー選択
		memberSelectSection := post.SelectMembersSection(md.GetAllEditedUserIDs(), aid.DeleteMemberSelectMember, []string{}, false, true)

		blocks = []slack.Block{headerSection, headerTipsSection, memberSelectSection}
	}
	return blocks
}

// メンバー削除リクエスト（確認）
func DeleteMemberConfirm(blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	Logger.Println("メンバー削除リクエスト")

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
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
		Logger.Printf("ユーザID: %s / チームリスト: %v\n", userID, memberTeamNames)

		headerSection := post.SingleTextSectionBlock(Markdown, "*以下メンバーを削除しますか？*")
		memberInfoSection := post.InfoMemberSection(userID, memberTeamNames, memberTeamNames)
		actionBtnActionId := strings.Join([]string{aid.DeleteMember, userID}, "_")
		actionBtnBlock := post.BtnOK("削除", actionBtnActionId)

		blocks = []slack.Block{headerSection, memberInfoSection, actionBtnBlock}
	}

	return blocks
}

// メンバー削除
func DeleteMember(blockActions map[string]map[string]slack.BlockAction, userID string) (blocks []slack.Block) {
	var ok bool

	// メンバーデータ 読み込み
	if md, err := data.LoadMember(); err != nil {
		blocks = md.GetErrBlocks(err, DataLoadErr)
	} else {
		delete(md, userID)

		if err = md.Update(); err != nil {
			blocks = md.GetErrBlocks(err, DataUpdateErr)
		} else {
			if err := md.SynchronizeTeam(); err != nil {
				blocks = post.SingleTextBlock(post.ErrText(ErrorSynchronizeData))
			} else {
				text := post.ScsText(fmt.Sprintf("メンバー <@%s> の削除に成功しました", userID))
				blocks, ok = post.SingleTextBlock(text), true

				Logger.Println("メンバー削除に成功しました")
			}
		}
	}

	if !ok {
		Logger.Println("メンバー削除に失敗しました．詳細は Slack に投稿されたメッセージを確認してください．")
	}
	return blocks
}
