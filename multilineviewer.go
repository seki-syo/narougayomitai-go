package main

//MultiLine Viewer
import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

//MultiLineViewer 複数行の文字を画面に表示するための構造体。termboxによる文字送りも可能
type MultiLineViewer struct {
	foldedArray []string
	currentLine int    //現在表示中のLine
	cancelFunc  func() //Escキーが押された時に実行されるキャンセル処理
	height      int
	width       int
	leftFunc    func() //左キーを押したときの関数
	rightFunc   func() //右キーを押したときの関数
}

//NewMultiLineViewer 作成
func NewMultiLineViewer() *MultiLineViewer {
	return &MultiLineViewer{}
}

//Init 初期化
func (v *MultiLineViewer) Init() {
	v.foldedArray = []string{}
	v.currentLine = 0 //最上部の行から描画
	v.cancelFunc = func() {}
	v.leftFunc = func() {}
	v.rightFunc = func() {}
	v.width, v.height = termbox.Size()
	//キー押下時の動作を設定
	SetInputFunction(v.moveUp, v.moveDown, v.leftFunc, v.rightFunc, v.cancelFunc, func() {}, v.moveTop, v.moveBottom)
}

//Draw 描画
func (v *MultiLineViewer) Draw() {
	//一行ずつ描画
	var drawLineCon int
	if v.height < len(v.foldedArray) {
		drawLineCon = v.height
	} else {
		drawLineCon = len(v.foldedArray)
	}
	for di := 0; di < drawLineCon; di++ {
		drawLineNoStatic(v.foldedArray[di+v.currentLine], 0, di, defaultFg, defaultBg)
	}
	//デバッグ用
	//drawLineNoStatic("drawLineCon = "+strconv.Itoa(drawLineCon), 60, 5, termbox.ColorRed, defaultBg)
	//drawLineNoStatic("currentLine = "+strconv.Itoa(v.currentLine), 60, 6, termbox.ColorRed, defaultBg)
	//drawLineNoStatic("foldedArrayLen = "+strconv.Itoa(len(v.foldedArray)), 60, 7, termbox.ColorRed, defaultBg)
	//drawLineNoStatic("height = "+strconv.Itoa(v.height), 60, v.height-1, termbox.ColorRed, defaultBg)
	drawScreen()
}

func (v *MultiLineViewer) moveUp() {
	if v.currentLine > 0 {
		v.currentLine--
	}
	v.Draw()
}

func (v *MultiLineViewer) moveDown() {
	if v.currentLine < len(v.foldedArray)-v.height {
		v.currentLine++
	}
	v.Draw()
}

func (v *MultiLineViewer) moveTop() {
	v.currentLine = 0
}

func (v *MultiLineViewer) moveBottom() {
	v.currentLine = len(v.foldedArray) - v.height
}

//SetLeftRightFunc 関数を設定する
func (v *MultiLineViewer) SetLeftRightFunc(left, right func()) {
	v.leftFunc = left
	v.rightFunc = right
	//キー押下時の動作を設定
	SetInputFunction(v.moveUp, v.moveDown, v.leftFunc, v.rightFunc, v.cancelFunc, func() {}, v.moveTop, v.moveBottom)
}

//CancelSetting MultiLineViewerにおけるEscキー押下時の動作を設定
func (v *MultiLineViewer) CancelSetting(f func()) {
	v.cancelFunc = f
	//キー押下時の動作を設定
	SetInputFunction(v.moveUp, v.moveDown, v.leftFunc, v.rightFunc, v.cancelFunc, func() {}, v.moveTop, v.moveBottom)
}

//SetStrings ビュワーに表示する文字列を設定
func (v *MultiLineViewer) SetStrings(str []string) {
	//折り返し行をスライスに追加していく
	for _, l := range str {
		fl := stringFold(l, v.width-8)
		v.foldedArray = append(v.foldedArray, fl...)
	}
}

//stringFold 文字列をwidthに合わせて折り返したものを配列にして返す
func stringFold(str string, w int) []string {
	var runecon int    //文字数カウンター
	var lineStr string //一行あたりの文字列
	var lines []string
	lcon := runewidth.StringWidth(str)
	if lcon <= w {
		//文字列の長さが画面幅より短いのならそのまま帰す
		return []string{str}
	}
	runecon = 0
	lineStr = ""
	lines = []string{}
	for _, r := range str {
		runecon += runewidth.RuneWidth(r)
		if runecon <= w {
			//文字列は画面の幅を超えなかった
			lineStr += string(r)
		} else {
			//文字列は画面の幅を超えてしまっている
			//lineStrをlines配列に加えて初期化
			lines = append(lines, lineStr)
			lineStr = string(r)
			runecon = runewidth.RuneWidth(r)
		}
	}
	lines = append(lines, lineStr)
	return lines
}
