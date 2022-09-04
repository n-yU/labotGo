// メッセージ投稿
package post

import (
	"fmt"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 定型セクション: シャッフル結果
func ShuffleResultSections(
	teamName string, memberUserIDs []string, md data.MembersData,
) (resultSections []*slack.ContextBlock) {
	teamNameElements := []slack.MixedElement{TxtBlockObj(util.Markdown, fmt.Sprintf("チーム名: *%s*", teamName))}
	resultSections = append(resultSections, slack.NewContextBlock("", teamNameElements...))

	elements := []slack.MixedElement{}
	for i, userID := range memberUserIDs {
		if len(elements) == 0 && i > 0 {
			elements = append(elements, TxtBlockObj(util.Markdown, "→"))
		}
		elements = append(elements, slack.NewImageBlockElement(md[userID].Image24, userID))
		if i < len(memberUserIDs)-1 {
			elements = append(elements, TxtBlockObj(util.Markdown, fmt.Sprintf("*<@%s>* →", userID)))
		} else {
			elements = append(elements, TxtBlockObj(util.Markdown, fmt.Sprintf("*<@%s>*", userID)))
		}

		// Context の element の上限は10個のためセクション化してリセット
		if len(elements) >= 8 {
			resultSections = append(resultSections, slack.NewContextBlock("", elements...))
			elements = []slack.MixedElement{}
		}
	}

	if len(elements) > 0 {
		resultSections = append(resultSections, slack.NewContextBlock("", elements...))
	}
	return resultSections
}
