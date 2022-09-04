// データ管理
package data

import (
	"fmt"
	"strings"
)

// 書籍サマリ
type BookSummary struct {
	Title     string `json:"title"`
	ISBN      string `json:"isbn"`
	Publisher string `json:"publisher"`
	Pubdate   string `json:"pubdate"`
	Cover     string `json:"cover"`
	Authors   string `json:"author"`
}

// 書籍
type Book struct {
	BookSummary `json:"summary"`
}

// 書籍リスト
type Books []*Book

// 書籍出版日 フォーマット
func (b *BookSummary) FormatPubdate() {
	if len(b.Pubdate) == 8 {
		b.Pubdate = strings.Join([]string{b.Pubdate[:4], b.Pubdate[4:6], b.Pubdate[6:]}, "-")
	} else {
		b.Pubdate = fmt.Sprintf("%s-01", b.Pubdate)
	}
}
