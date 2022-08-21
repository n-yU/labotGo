package main

import "github.com/slack-go/slack"

func listen(cmd slack.SlashCommand) map[string]interface{} {
	user_name, cmd_text := cmd.UserName, cmd.Text
	logger.Printf("受信コマンド (from:%s): %s\n", user_name, cmd_text)

	payload := map[string]interface{}{
		"blocks": []slack.Block{slack.NewSectionBlock(
			&slack.TextBlockObject{Type: slack.MarkdownType, Text: "foo"},
			nil, nil,
		)},
	}

	return payload
}
