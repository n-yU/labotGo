// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// メッセージ投稿
func PostMessage(data interface{}, blocks []slack.Block, responseType string) error {
	var (
		channelId string
		userId    string
		err       error
	)

	// メッセージ投稿 チャンネル/ユーザID 取得
	switch data := data.(type) {
	case slack.SlashCommand:
		cmd := data
		channelId, userId = cmd.ChannelID, cmd.UserID
	case slack.InteractionCallback:
		callback := data
		channelId, userId = callback.Channel.Conversation.ID, callback.User.ID
	default:
		err = fmt.Errorf("%T 型のデータはメッセージ投稿に対応していません", data)
		return err
	}

	// メッセージ投稿
	switch responseType {
	case util.InChannel:
		_, _, err = util.SocketModeClient.PostMessage(channelId, slack.MsgOptionBlocks(blocks...))
	case util.Ephemeral:
		_, err = util.SocketModeClient.PostEphemeral(channelId, userId, slack.MsgOptionBlocks(blocks...))
	default:
		err = fmt.Errorf("レスポンスタイプ %s は存在しません", responseType)
	}

	return err
}

// 動作チェック用 メッセージ投稿
func Start(defaultCh string) error {
	text := InfoText(fmt.Sprintf("*<https://github.com/n-yU/labotGo|labotGo> v%s を起動しました*\n", util.Version))
	message := slack.MsgOptionText(text, false)

	_, _, err := util.SocketModeClient.PostMessage(defaultCh, message)
	return err
}
