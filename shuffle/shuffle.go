// 機能: メンバーシャッフル
package shuffle

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	teamNamesString, values := cmdValues[0], cmdValues[1:]

	if len(values) > 0 {
		text := post.ErrText(fmt.Sprintf("コマンド %s shuffle に2つ以上の引数を与えることはできません", util.Cmd))
		textSection := post.SingleTextSectionBlock(util.Markdown, text)
		tipsText := []string{fmt.Sprintf("複数のチームを指定する場合は `%s shuffle A,B,C` のようにカンマ区切りで指定してください", util.Cmd)}
		tipsSection := post.TipsSection(tipsText)
		blocks, responseType, ok = []slack.Block{textSection, tipsSection}, util.Ephemeral, false
	} else {
		blocks, responseType = getBlocksShuffle(teamNamesString)
		ok = true
	}
	return blocks, responseType, ok
}

// メンバーシャッフル結果 取得
func getBlocksShuffle(teamNamesString string) ([]slack.Block, string) {
	var (
		md  data.MembersData
		err error
	)
	util.Logger.Println("メンバーシャッフルリクエスト")

	// メンバーデータ 読み込み
	if md, err = data.LoadMember(); err != nil {
		blocks := post.ErrBlocksMembersData(err, util.DataLoadErr)
		return blocks, util.Ephemeral
	}

	// 指定チームメンバー 取得
	memberUserIDs, blocks := getTeamsMembers(teamNamesString)
	if len(blocks) > 0 {
		return blocks, util.Ephemeral
	}

	headerText := post.ScsText("指定チームのメンバーの順番をシャッフルしました")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)
	blocks = []slack.Block{headerSection, util.Divider()}

	for teamName, memberUIDs := range memberUserIDs {
		util.Logger.Printf("指定メンバー: %s\n", strings.Join(memberUIDs, ", "))

		// メンバーシャッフル
		shuffledMemberUIDs := ShuffleMemberUserIDs(memberUIDs)
		if strings.Contains(teamName, "+") {
			teamName = strings.Replace(teamName, "+", " + ", -1)
		}

		// シャッフル結果セクション 追加
		shuffleResultSections := shuffleResultSections(teamName, shuffledMemberUIDs, md)
		for _, shuffleResSec := range shuffleResultSections {
			blocks = append(blocks, shuffleResSec)
		}
		blocks = append(blocks, util.Divider())
	}

	util.Logger.Println("メンバーシャッフルに成功しました")
	return blocks, util.InChannel
}

// 指定チームメンバー 取得
func getTeamsMembers(teamNamesString string) (memberUserIDs map[string][]string, blocks []slack.Block) {
	memberUserIDs = map[string][]string{}
	// チームデータ 読み込み
	if td, err := data.LoadTeam(); err != nil {
		blocks = post.ErrBlocksTeamsData(err, util.DataLoadErr)
	} else {
		// バリデーションチェック
		isContainComma, isContainPlus := strings.Contains(teamNamesString, ","), strings.Contains(teamNamesString, "+")
		if isContainComma && isContainPlus {
			text := post.ErrText("チーム指定の文字列には \",\" と \"+\" を併用できません")
			blocks = post.SingleTextBlock(text)
		} else if isContainComma {
			// 複数チーム指定
			teamNames := util.UniqueSlice(strings.Split(teamNamesString, ","))
			for _, teamName := range teamNames {
				memberUserIDs[teamName] = []string{}
				if team, ok := td[teamName]; ok {
					for _, userID := range team.UserIDs {
						memberUserIDs[teamName] = append(memberUserIDs[teamName], userID)
					}
				} else {
					blocks = post.ErrBlocksUnknownTeam(teamName)
				}
			}
		} else if isContainPlus {
			// 複合チーム指定
			teamNames := util.UniqueSlice(strings.Split(teamNamesString, "+"))
			complexTeamName := strings.Join(teamNames, "+")
			memberUserIDs[complexTeamName], _ = td.GetComplexTeamMemberUserIDs(teamNames)
		} else {
			// 単独チーム指定
			memberUserIDs[teamNamesString] = []string{}
			if team, ok := td[teamNamesString]; ok {
				for _, userID := range team.UserIDs {
					memberUserIDs[teamNamesString] = append(memberUserIDs[teamNamesString], userID)
				}
			} else {
				blocks = post.ErrBlocksUnknownTeam(teamNamesString)
			}
		}
	}

	// チームメンバー有無 チェック
	for teamName, userIDs := range memberUserIDs {
		if len(userIDs) == 0 {
			text := post.ErrText(fmt.Sprintf("指定したチーム `%s` のメンバー数が 0人 のため，シャッフルできません", teamName))
			blocks = post.SingleTextBlock(text)
			break
		}
	}

	if len(blocks) == 0 {
		util.Logger.Printf("指定チーム: %s\n", teamNamesString)
	} else {
		util.Logger.Printf("指定チーム \"%s\" は不適切な形式です", teamNamesString)
	}
	return memberUserIDs, blocks
}

// メンバーシャッフル
func ShuffleMemberUserIDs(userIDs []string) (shuffledUserIDs []string) {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	shuffledUserIDs = append(shuffledUserIDs, userIDs...)
	rnd.Shuffle(len(shuffledUserIDs), func(i, j int) {
		shuffledUserIDs[i], shuffledUserIDs[j] = shuffledUserIDs[j], shuffledUserIDs[i]
	})

	return shuffledUserIDs
}

// 定型セクション: シャッフル結果
func shuffleResultSections(
	teamName string, memberUserIDs []string, md data.MembersData,
) (resultSections []*slack.ContextBlock) {
	teamNameElements := []slack.MixedElement{post.TxtBlockObj(util.Markdown, fmt.Sprintf("チーム名: *%s*", teamName))}
	resultSections = append(resultSections, slack.NewContextBlock("", teamNameElements...))

	elements := []slack.MixedElement{}
	for i, userID := range memberUserIDs {
		if len(elements) == 0 && i > 0 {
			elements = append(elements, post.TxtBlockObj(util.Markdown, "→"))
		}
		elements = append(elements, slack.NewImageBlockElement(md[userID].Image24, userID))
		if i < len(memberUserIDs)-1 {
			elements = append(elements, post.TxtBlockObj(util.Markdown, fmt.Sprintf("*<@%s>* →", userID)))
		} else {
			elements = append(elements, post.TxtBlockObj(util.Markdown, fmt.Sprintf("*<@%s>*", userID)))
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
