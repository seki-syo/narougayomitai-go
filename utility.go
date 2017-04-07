package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

//描画機能を持つ。基本的にview型は表示画面を形成するのであって、表示機能はこの構造体が持つ
//文字を順次drawSpanによって描画する機能を持たない
var (
	screenBuffer         []drawBuffer //画面にある全文字を持っている
	noStaticScreenBuffer []drawBuffer //動的描画用のバッファ
)

//cell 文字を描画する際にchannelに情報を受け渡すための構造体
type drawBuffer struct {
	x   int
	y   int
	str string
	fg  termbox.Attribute
	bg  termbox.Attribute
}

//initDraw 作成
func initDraw() {
	termbox.Clear(defaultFg, defaultBg)
	screenBuffer = []drawBuffer{}
	noStaticScreenBuffer = []drawBuffer{}
}

//drawScreen 一斉に描写
func drawScreen() {
	Clear()
	allScreenBuffer := append(screenBuffer, noStaticScreenBuffer...)
	for _, b := range allScreenBuffer {
		drawScreenWithBuffer(b)
	}
	Draw()
	noStaticScreenBuffer = []drawBuffer{}
}

//Clear 内部バッファを消去
func Clear() {
	err := termbox.Clear(defaultFg, defaultBg)
	if err != nil {
		panic(err)
	}
}

//Draw 描画を行う（Flush処理）
func Draw() {
	err := termbox.Flush()
	if err != nil {
		panic(err)
	}
}

//drawScreenWithBuffer drawBuffer構造体を描画
func drawScreenWithBuffer(bf drawBuffer) {
	i := 0
	ni := 0
	for _, r := range bf.str {
		i += runewidth.RuneWidth(r)
		if bf.x+i >= width {
			//横幅以上に書き込んだら
			break
		}
		termbox.SetCell(bf.x+ni, bf.y, r, bf.fg, bf.bg)
		ni = i
	}
}

//drawLine 指定位置にStringを表示(文字サイズによって全角半角判別可能)
func drawLine(s string, x, y int, fg, bg termbox.Attribute) {
	i := 0
	ni := 0
	for _, r := range s {
		i += runewidth.RuneWidth(r)
		if x+i >= width {
			//横幅以上に書き込んだら
			break
		}
		termbox.SetCell(x+ni, y, r, fg, bg)
		ni = i
	}
	err := termbox.Flush()
	if err != nil {
		panic(err)
	}
	screenBuffer = append(screenBuffer, drawBuffer{x, y, s, fg, bg})
}

//drawLineNoStatic 指定位置にStringを表示（動的文字とするので記録されない）
func drawLineNoStatic(s string, x, y int, fg, bg termbox.Attribute) {
	noStaticScreenBuffer = append(noStaticScreenBuffer, drawBuffer{x, y, s, fg, bg})
}

//drawRow yに指定した縦座標に横一列を指定の文字列で埋め尽くす（動的文字とするので記録されない）
func drawRowNoStatic(s string, y int, fg, bg termbox.Attribute) {
	i := 0
	row := []string{}
	for {
		i += runewidth.StringWidth(s)
		if i >= width {
			//横幅以上に書き込んだら
			break
		}
		row = append(row, s)
	}
	drawLineNoStatic(strings.Join(row, ""), 0, y, fg, bg)
}

//drawRow yに指定した縦座標に横一列を指定の文字列で埋め尽くす
func drawRow(s string, y int, fg, bg termbox.Attribute) {
	i := 0
	row := []string{}
	for {
		i += runewidth.StringWidth(s)
		if i >= width {
			//横幅以上に書き込んだら
			break
		}
		row = append(row, s)
	}
	drawLine(strings.Join(row, ""), 0, y, fg, bg)
}

//stringJoinRow 画面横一列を埋め尽くすStringを返す
func stringJoinRow(s string, w int) string {
	i := 0
	row := []string{}
	for {
		i += runewidth.StringWidth(s)
		if i >= w {
			//横幅以上に書き込んだら
			break
		}
		row = append(row, s)
	}
	return strings.Join(row, "")
}
