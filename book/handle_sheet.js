//------ 設定項目 ------//

// シートID
const SHEET_ID = "<シートID>";
// シート名
const SHEET_NAME = "<シート名>";

//------ 以下編集不要 ------//

// スプレッドシート 取得
function get_sheet() {
    var spreadSheet = SpreadsheetApp.openById(SHEET_ID);
    var sheet = spreadSheet.getSheetByName(SHEET_NAME);
    return sheet;
}

// スプレッドシート 書籍登録
function registerBook(prm) {
  var sheet = get_sheet();
  var isbn10 = prm.isbn10;

  // ISBN-10による登録済み書籍検索
  searchIsbn10 = sheet.getRange(2, 7, sheet.getLastRow(), 1).createTextFinder(isbn10).findNext();
  if(searchIsbn10) {
    // 検索ヒット -> 登録せずヒット位置返却
    return {'status': 'registered', 'cell': search_isbn10.getA1Notation()};
  }

  // 仮登録
  sheet.appendRow(['登録中', '登録中', '登録中', '登録中', '登録中', '登録中', isbn10, '登録中']);

  // 書籍情報準備
  var book = [prm.location, prm.category, prm.title, prm.publisher, prm.firstAuthor, prm.pubdate, isbn10, prm.amazon];
  var last = sheet.getLastRow();

  // 書籍登録
  sheet.getRange(last - 1, 1, 1, 8).copyTo(sheet.getRange(last, 1, 1, 8), SpreadsheetApp.CopyPasteType.PASTE_NORMAL);
  sheet.getRange(last, 1, 1, 8).setValues([book]);
  sheet.setRowHeight(sheet.getLastRow(), 50);

  return {'status': 'success', 'row': last};
}

// スプレッドシート 書籍削除
function deleteBook(prm) {
  var sheet = get_sheet();
  var isbn10 = prm.isbn10;

  // ISBN-10による登録済み書籍検索
  searchIsbn10 = sheet.getRange(2, 7, sheet.getLastRow(), 1).createTextFinder(isbn10).findNext();
  if(!searchIsbn10) {
    // 検索ヒットせず -> 削除失敗通知
    return {'status': 'unregistered'};
  }

  // 書籍削除
  foundIsbn10Cell = searchIsbn10.getA1Notation();
  sheet.deleteRows(parseInt(foundIsbn10Cell.slice(1), 10))

  console.log(foundIsbn10Cell);
  return {'status': 'success', 'cell': foundIsbn10Cell};
}

function doGet(e) {
  // パラメータ 取得
  var prm = e.parameter;
  handle = prm.handle

  // シート 操作
  if(handle == 'register') {
    result = registerBook(prm);
  } else if(handle == 'delete') {
    result = deleteBook(prm);
  } else {
    result = {'status': 'undifined'};
  }

  // 結果 出力
  var output = ContentService.createTextOutput();
  output.setMimeType(ContentService.MimeType.JSON);
  output.setContent(JSON.stringify(result));
  return output;
}
