package main

import (
	"strconv"

	"github.com/nsf/termbox-go"
)

//ChoiceListDisplayPattern 選択肢リストの表示形式
type ChoiceListDisplayPattern int

const (
	//pat1 1.~~
	pat1 ChoiceListDisplayPattern = iota
	//pat2 ・~~
	pat2
	//pat3 [~~]
	pat3
)

//DrawPosition DrawnlineWithPositionの字詰め指定
type DrawPosition int

const (
	//LeftDrawn 左字詰めを指定
	LeftDrawn DrawPosition = iota
	//RightDrawn 右字詰めを指定
	RightDrawn
	//CenterDrawn 中央に字を寄せる
	CenterDrawn
)

//section ある地点と範囲を規定
type section struct {
	origin   int //原点
	distance int //範囲
}

//endPoint 終点
func (r *section) endPoint() int {
	return r.origin + r.distance - 1
}

//Lines Stringの配列。index0が項目のトップメニュー
type Lines []string

//Len Linesの行数を返す
func (l Lines) len() int {
	return len(l)
}

//setStrings 項目達をセット
func setStrings(items []string) {
	lsarr := []Lines{}
	for _, l := range items {
		//string配列をLinesとして代入
		ls := Lines([]string{l})
		lsarr = append(lsarr, ls)
	}
	selectionMultipleLines = lsarr
}

//setMultipleLines 項目たちをセット
func setMultipleLines(items []Lines) {
	selectionMultipleLines = items
}

var (
	currentCursor          int                      //現在選択中の項目
	pattern                ChoiceListDisplayPattern //リストの表示形式
	selectionMultipleLines []Lines                  //複数行にわたる選択肢[選択項目]
	stDrawPos              int                      //現在の描画開始位置
	choiExe                func(index int)          //選択時実行関数
	drawArea               *section                 //描画範囲(y座標指定)
	cancelExist            bool                     //キャンセルを表示するならtrue
	cancelString           string                   //キャンセルの項目の名前
	cancelExe              func()                   //キャンセルを指定した時に実行
	noChoiFg               termbox.Attribute        //選択肢の文字
	noChoiBg               termbox.Attribute        //選択肢のバックグラウンド
	choiFg                 termbox.Attribute        //選択中の文字
	choiBg                 termbox.Attribute        //選択中のバックグラウンド
)

//Init 初期化
func initChoiceList() {
	stDrawPos = 0                      //項目も初期位置
	currentCursor = 0                  //初期位置
	selectionMultipleLines = []Lines{} //表記項目も初期化
	choiExe = nil
	drawArea = &section{
		width,
		height,
	}
	cancelExist = false
	cancelString = "キャンセル"
	cancelExe = nil
	noChoiFg = defaultFg
	noChoiBg = defaultBg
	choiFg = termbox.ColorBlack
	choiBg = termbox.ColorGreen
	pattern = pat1

	//キーを設定
	SetInputFunction(moveUp, moveDown, func(){}, func(){}, cancelExe, selectExecute, func(){}, func(){})
}

//setPattern リストパターンを設定
func setPattern(p ChoiceListDisplayPattern) {
	pattern = p
}

//setExecute selectExecuteで実行される関数を設定
func setExecute(e func(int)) {
	choiExe = e
	//キーを設定
	SetInputFunction(moveUp, moveDown, func(){}, func(){}, cancelExe, selectExecute, func(){}, func(){})
}

//setSection 描画範囲を設定
func setSection(origin, distance int) {
	drawArea = &section{
		origin,
		distance,
	}
}

//selectExecute 選択肢を実行
func selectExecute() {
	if currentCursor < len(selectionMultipleLines) {
		//キャンセル以外の選択肢を実行
		if choiExe != nil {
			choiExe(currentCursor)
		}
	} else {
		if cancelExe != nil {
			cancelExe()
		}
	}
}

//setColorSetting 選択肢のカラーをセッティングする
func setColorSetting(f, b, fg, bg termbox.Attribute) {
	noChoiFg = f
	noChoiBg = b
	choiFg = fg
	choiBg = bg
}

//cancelSetting キャンセルを選択したときの挙動
func cancelSetting(exist bool, str string, exe func()) {
	cancelExist = exist
	cancelString = str
	cancelExe = exe
	//キーを設定
	SetInputFunction(moveUp, moveDown, func(){}, func(){}, cancelExe, selectExecute, func(){}, func(){})
}

//listLen リストの要素数を返す(キャンセルも項目数に含む)
func listLen() int {
	if cancelExist {
		return len(selectionMultipleLines) + 1
	}
	return len(selectionMultipleLines)
}

//getIndexFromPos posによりLinesからpos番目の行のlinesのindexと[]stringのindexを取得
func getSelectionMultipleLinesIndexFromPos(pos int) (linesindex, inlineindex int) {
	num := 0 //行数カウンター
funcloop:

	for i := 0; i < listLen(); i++ {
		itemLines := getSelectionMultipleLinesItem(i)
		for ii := range itemLines {
			if pos == num {
				linesindex = i
				inlineindex = ii
				break funcloop
			}
			num++
		}
	}
	return
}

//getSelectionMultipleLinesItem index値からキャンセル項目を考慮したLinesアイテムが返される
func getSelectionMultipleLinesItem(index int) Lines {
	if index >= len(selectionMultipleLines) && cancelExist {
		//キャンセル項目
		return Lines([]string{cancelString})
	}
	//それ以外
	return selectionMultipleLines[index]
}

//getPosFromLinesIndex index値から項目の始点と終点を取得する
func getPosFromLinesIndex(linesindex int) (startpos, endpos int) {

	pos := 0 //行数カウンター
	for li := 0; li <= linesindex; li++ {
		ls := getSelectionMultipleLinesItem(li)
		for si := range ls {
			if li == linesindex && si == 0 {
				//とある項目の始点
				startpos = pos
			}
			if li == linesindex && (si == len(ls)-1) {
				//とある項目の終点
				endpos = pos
				break
			}
			pos++
		}
	}
	return
}

//getPosFromSelectionMultipleLinesIndexAndInLineIndex []LinesとLinesのインデックスから行の位置を割り出す
func getPosFromSelectionMultipleLinesIndexAndInLineIndex(linesIndex, inLineIndex int) (pos int) {

	for i := 0; i <= linesIndex; i++ {
		ls := getSelectionMultipleLinesItem(i)
		for ii := range ls {
			if i == linesIndex && ii == inLineIndex {
				return
			}
			pos++
		}
	}
	return
}

//getLenFromSelectionMultipleLines 指定したインデックス値のLinesのアイテム数を取得（キャンセル項目にも対応している)
func getLenFromSelectionMultipleLines(index int) int {
	if cancelExist {
		//キャンセル項目あり
		if index < len(selectionMultipleLines) {
			//キャンセル項目以外
			return len(selectionMultipleLines[index])
		}
	} else {
		//キャンセル項目は考慮しない
		return len(selectionMultipleLines[index])
	}

	//キャンセル項目
	return 1
}

//moveUp リストを上へ
func moveUp() {
	if (currentCursor == 0) || (listLen() == 1) {
		//カーソルが一番上か下に要素が存在しないときは何もしない
		return
	}
	currentCursor-- //一つ減算
	drawChoiceList()
}

//moveDown リストを下へ
func moveDown() {
	if currentCursor > (listLen() - 1) {
		//カーソルが項目外（下）にいるときは最後の項目にカーソルを合わせる
		currentCursor = listLen() - 1
	}

	if (currentCursor == (listLen() - 1)) || (listLen() == 1) {
		//カーソルが一番下か、項目が一つのときは表示しない
		return
	}

	currentCursor++ //一つ加算
	drawChoiceList()
}

//drawChoiceLineInLines Lines構造体の特定の行を描画する（インデックスに従い、書き方を変える）
func drawChoiceLineInLines(linesindex, inlineindex, y int, fg, bg termbox.Attribute) {
	lines := getSelectionMultipleLinesItem(linesindex)
	line := lines[inlineindex]

	//パターンに従って文字を描画
	switch pattern {
	case pat1:
		if inlineindex == 0 {
			//先頭の行
			//1.  ~~
			//    ~~
			drawRowNoStatic(" ", y, fg, bg)
			drawLineNoStatic(strconv.Itoa(linesindex)+".", 0, y, fg, bg) //先頭の行のみ描画
			drawLineNoStatic(line, 4, y, fg, bg)
		} else {
			drawLineNoStatic(line, 4, y, defaultFg, defaultBg) //その他の行を描画
		}
	case pat2:
		//>~~
		// ~~
		if inlineindex == 0 {
			drawRowNoStatic(" ", y, fg, bg)
			drawLineNoStatic(">"+line, 0, y, fg, bg)
		} else {
			drawLineNoStatic(line, 3, y, defaultFg, defaultBg)
		}
	case pat3:
		//[~~]
		// ~~
		if inlineindex == 0 {
			drawRowNoStatic(" ", y, fg, bg)
			drawLineNoStatic("["+line+"]", 0, y, fg, bg)
		} else {
			drawLineNoStatic(line, 3, y, defaultFg, defaultBg)
		}

	default:
		//~~~
		// ~~
		if inlineindex == 0 {
			drawRowNoStatic(" ", y, fg, bg)
			drawLineNoStatic(line, 0, y, fg, bg)
		} else {
			drawLineNoStatic(line, 2, y, defaultFg, defaultBg)
		}
	}
}

func drawChoiceList() {

	//選択中の項目の上端と下端の位置を確認
	cursorStartPos, cursorEndPos := getPosFromLinesIndex(currentCursor)
	//現在選択中の項目のはみ出しがないか確認する。下端から先にチェックすることで、下端より上端のはみ出しを進んで修正する
	//選択中の項目が描画範囲の下端から出ているか確認する
	if stDrawPos+drawArea.distance-1 < cursorEndPos {
		//選択項目の下端がはみ出ている
		stDrawPos += cursorEndPos - (drawArea.distance + stDrawPos - 1)
	}
	//選択項目中の項目が描画範囲の上端からはみ出ているか確認する
	if stDrawPos > cursorStartPos {
		//選択項目の上端がはみ出ている
		stDrawPos = cursorStartPos //描画開始位置を選択中項目の一番上に設定
	}

	//始点となるselectionMultipleLinesのindexとLines内のindexを取得する(何項目の何行目から始まるのか)
	stLinesIndex, stInLineIndex := getSelectionMultipleLinesIndexFromPos(stDrawPos) //描画を開始する項目インデックスを取得

	//描画
	currentDrawPos := drawArea.origin                      //現在描画している位置
	var currentDrawLinesIndex, currentDrawInLinesIndex int //現在描画しているLines配列のインデックスとLinesのインデックス

	currentDrawLinesIndex = stLinesIndex
	currentDrawInLinesIndex = stInLineIndex

	//描画していく
	for currentDrawPos = drawArea.origin; currentDrawPos <= drawArea.endPoint(); currentDrawPos++ {
		//描画処理
		if currentCursor == currentDrawLinesIndex {
			//現在選択中の項目なので描画色を変える
			drawChoiceLineInLines(currentDrawLinesIndex, currentDrawInLinesIndex, currentDrawPos, choiFg, choiBg)
		} else {
			//選択中でない項目
			drawChoiceLineInLines(currentDrawLinesIndex, currentDrawInLinesIndex, currentDrawPos, noChoiFg, noChoiBg)
		}

		//描画処理終了でインデックス値の管理
		if currentDrawInLinesIndex >= getLenFromSelectionMultipleLines(currentDrawLinesIndex)-1 {
			//現在描いている項目の最終行に到達したので、inLineIndexを初期化し、LinesIndexを一つ進める(事前に最後の項目でないか確認する)
			if currentDrawLinesIndex >= listLen()-1 {
				//最終項目を描画し終わったので終了
				break
			} else {
				currentDrawLinesIndex++
			}
			currentDrawInLinesIndex = 0 //初期化
		} else {
			currentDrawInLinesIndex++ //項目内の行数を一つ進める
		}
	}
	//デバッグ
	//drawLineNoStatic("stLinesIndex = "+strconv.Itoa(stLinesIndex)+" stInLineIndex = "+strconv.Itoa(stInLineIndex), 80, 0, termbox.ColorRed, defaultBg)
	//drawLineNoStatic("stDrawPos = "+strconv.Itoa(stDrawPos)+" endDrawPos = "+strconv.Itoa(stDrawPos+drawArea.distance-1), 80, 1, termbox.ColorRed, termbox.ColorDefault)
	///drawLineNoStatic("cursorStartPos = "+strconv.Itoa(cursorStartPos)+" cursorEndPos = "+strconv.Itoa(cursorEndPos), 80, 2, termbox.ColorRed, termbox.ColorDefault)
	//drawLineNoStatic("origin = "+strconv.Itoa(drawArea.origin)+" distance = "+strconv.Itoa(drawArea.distance), 80, drawArea.origin, termbox.ColorRed, termbox.ColorDefault)
	//drawLineNoStatic("endLine = "+strconv.Itoa(drawArea.endPoint()), 80, drawArea.endPoint(), termbox.ColorRed, termbox.ColorDefault)
	//静的文字列を一斉描画
	drawScreen()
}
