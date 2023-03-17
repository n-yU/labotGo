// 機能: 書籍管理
package book

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/n-yU/labotGo/aid"
	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// 書籍一括登録リクエスト
func getBlocksRegisterBulkRequest() (blocks []slack.Block) {
	// ブロック: ヘッダ
	headerText := post.InfoText("*登録したい複数の書籍の ISBNコード を改行で区切って入力してください*")
	headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

	// ブロック: ヘッダ Tips
	headerTipsSection := post.TipsIBSNSection()

	// ブロック: ISBNコード 入力
	ISBNInputSection := post.InputISBNSection(aid.RegisterBulkBookInputISBN, true)

	// ブロック: 確認ボタン
	actionBtnBlock := post.CustomBtnSection("OK", "一括登録", aid.RegisterBulkBook)

	blocks = []slack.Block{
		headerSection, headerTipsSection, util.Divider(), ISBNInputSection, actionBtnBlock,
	}
	return blocks
}

// 書籍一括登録
func RegisterBulkBook(actionUserID string, blockActions map[string]map[string]slack.BlockAction) (blocks []slack.Block) {
	var registerResults []string
	util.Logger.Printf("書籍一括登録リクエスト (from %s): %+v\n", actionUserID, blockActions)

	// 複数ISBNコード 取得
	ISBNs := getMultiISBN(aid.RegisterBulkBookInputISBN, blockActions)

	if nRequestBooks := len(ISBNs); nRequestBooks > 100 {
		// 登録書籍 > 100冊
		blocks = post.SingleTextBlock(post.ErrText(fmt.Sprintf(
			"一括登録できる書籍数の上限は *100冊* です（現在%d冊）", nRequestBooks,
		)))
	} else if nRequestBooks == 0 {
		// 登録書籍 = 0冊
		blocks = post.SingleTextBlock(post.ErrText("登録したい書籍の ISBNコード が入力されていません"))
	} else {
		// 登録書籍 = 1-100冊
		var nRegisterBooks int
		for _, tmpISBN := range ISBNs {
			util.Logger.Printf("ISBNコード: %v\n", tmpISBN)

			var (
				ISBN        string
				innerBlocks []slack.Block
				books       data.Books
			)

			// バリデーションチェック
			ISBN, innerBlocks = validateISBN(tmpISBN)
			if len(innerBlocks) > 0 {
				registerResults = append(registerResults, "バリデーションエラー")
				continue
			}

			// 書籍情報 取得（from OpenBD）
			res, err := requestOpenBD(ISBN)
			if err != nil || res.StatusCode != 200 {
				registerResults = append(registerResults, "OpenBDリクエストエラー")
				continue
			}

			// レスポンスボディ 読み込み
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				registerResults = append(registerResults, "レスポンスボディ読み込みエラー")
				continue
			}

			// 書籍情報 読み込み
			if err := json.Unmarshal(body, &books); err != nil {
				registerResults = append(registerResults, "書籍情報読み込みエラー")
				continue
			}

			// 書籍情報有無 確認
			if innerBlocks := post.UnknownBookBlock(books[0], ISBN); len(innerBlocks) > 0 {
				registerResults = append(registerResults, "書籍情報不明エラー")
				continue
			}

			// 書籍サマリ 取得
			(&books[0].BookSummary).SetPubdateYMD()
			books[0].SetContent()
			bookSummary := books[0].BookSummary
			util.Logger.Printf("書籍情報: %s %s %s\n", bookSummary.Title, bookSummary.Publisher, bookSummary.Authors)

			// book-doc への書籍追加
			if err := es.PutDocument(bookSummary); err != nil {
				switch err {
				case es.ErrDocAlreadyExist:
					registerResults = append(registerResults, "登録済")
				default:
					registerResults = append(registerResults, "書籍ドキュメント追加エラー")
				}
				continue
			}

			// 書籍DB追加
			if err := bookSummary.AddDB(); err != nil {
				registerResults = append(registerResults, "書籍データベース追加エラー")
				continue
			}

			registerResults = append(registerResults, "書籍登録成功")
			nRegisterBooks++
		}

		if nRequestBooks == nRegisterBooks {
			// 全指定書籍 登録成功
			blocks = post.SingleTextBlock(post.ScsText(fmt.Sprintf("指定した *%d冊* の書籍のすべての登録に成功しました", nRegisterBooks)))
		} else {
			// 一部指定書籍 登録失敗
			registerErrorTypes := map[string][]string{}

			// ブロック: ヘッダ
			headerText := post.ScsText(fmt.Sprintf(
				"指定した%d冊の書籍のうち *%d冊* の登録に成功しました\n\n登録に失敗した書籍は次の通りです（[i]: i行目）", nRequestBooks, nRegisterBooks,
			))
			headerSection := post.SingleTextSectionBlock(util.Markdown, headerText)

			// ブロック: ヘッダ Tips
			headerTipsText := []string{
				"各エラーの詳細を確認したい場合は該当書籍の個別登録を試みてください",
				fmt.Sprintf("現在 *%d* 冊の書籍が *labotGo* に登録されています", es.CountDoc(util.EsBookIndex)),
			}
			headerTipsSection := post.TipsSection(headerTipsText)
			blocks = []slack.Block{headerSection, headerTipsSection, util.Divider()}

			// 登録エラー情報 格納
			for i, rr := range registerResults {
				if rr != "書籍登録成功" {
					if _, ok := registerErrorTypes[rr]; !ok {
						registerErrorTypes[rr] = []string{}
					}
					resDetail := fmt.Sprintf("[%d] %s", i+1, ISBNs[i])
					registerErrorTypes[rr] = append(registerErrorTypes[rr], resDetail)
				}
			}
			util.Logger.Printf("登録エラー: %v\n", registerErrorTypes)

			// ブロック: 登録エラー情報
			for errorType, errorDetails := range registerErrorTypes {
				errorTitleSection := post.SingleTextSectionBlock(util.Markdown, post.ErrText(errorType))
				errorISBNs := post.SingleTextSectionBlock(util.Markdown, strings.Join(errorDetails, ", "))

				errorBlocks := []slack.Block{errorTitleSection, errorISBNs}
				blocks = append(blocks, errorBlocks...)
			}
		}
	}

	return blocks
}
