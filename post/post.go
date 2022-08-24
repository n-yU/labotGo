// メッセージ投稿
package post

import (
	"errors"
	"fmt"

	. "github.com/n-yU/labotGo/util"
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
	switch data.(type) {
	case slack.SlashCommand:
		cmd := data.(slack.SlashCommand)
		channelId, userId = cmd.ChannelID, cmd.UserID
	case slack.InteractionCallback:
		callback := data.(slack.InteractionCallback)
		channelId, userId = callback.Channel.Conversation.ID, callback.User.ID
	default:
		err = errors.New(fmt.Sprintf("%T 型のデータはメッセージ投稿に対応していません\n", data))
		return err
	}

	// メッセージ投稿
	switch responseType {
	case InChannel:
		_, _, err = SocketModeClient.PostMessage(channelId, slack.MsgOptionBlocks(blocks...))
	case Ephemeral:
		_, err = SocketModeClient.PostEphemeral(channelId, userId, slack.MsgOptionBlocks(blocks...))
	default:
		err = errors.New(fmt.Sprintf("レスポンスタイプ %s は存在しません\n", responseType))
	}

	return err
}

// 動作チェック用 メッセージ投稿
func Start(defaultCh string) error {
	text := ScsText(fmt.Sprintf("*<https://github.com/n-yU/labotGo|labotGo> v%s を起動しました*\n", Version))
	message := slack.MsgOptionText(text, false)

	_, _, err := SocketModeClient.PostMessage(defaultCh, message)
	return err
}
