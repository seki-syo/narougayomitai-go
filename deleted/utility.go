package main

import (
	"fmt"
	"math"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

//描画機能を持つ。基本的にview型は表示画面を形成するのであって、表示機能はこの構造体が持つ
var (
	drawSpan    int       //文字の表示間隔
	nowDrawSpan int       //文字の表示間隔(現在)
	drawerChan  chan int  //文字の表示間隔を制御するチャンネル
	runebaff    chan cell //文字を流し込むチャンネル
	screenBaff  []cell    //画面にある全文字を持っている
)

//cell 文字を描画する際にchannelに情報を受け渡すための構造体
type cell struct {
	x  int
	y  int
	c  rune
	fg termbox.Attribute
	bg termbox.Attribute
}

//iinitDraw 作成
func initDraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawSpan = 0
	nowDrawSpan = drawSpan
	drawerChan = make(chan int, 1)
	runebaff = make(chan cell, 1)
	screenBaff = []cell{}
}

//drawLoop 表示機構をスタートさせる
func drawLoop() {
	//チャンネルに流し込まれた文字を設定された間隔で表示するGoルーチンを作成
	for {
		select {
		case c, ok := <-runebaff:
			if ok {
				termbox.SetCell(c.x, c.y, c.c, c.fg, c.bg)
				screenBaff = append(screenBaff, c)
				err := termbox.Flush()
				if err != nil {
					fmt.Println(err)
				}
				if nowDrawSpan != 0 {
					time.Sleep(time.Duration(nowDrawSpan) * time.Millisecond)
				}
			}
		case nowDrawSpan = <-drawerChan: //文字の表示間隔を変更

		default:
		}
	}
}

//drawScreenPrompt 一斉に描写
func drawScreenPrompt() {
	for _, c := range screenBaff {
		termbox.SetCell(c.x, c.y, c.c, c.fg, c.bg)
	}

	err := termbox.Flush()
	if err != nil {
		fmt.Println(err)
	}
}

//zeroDrawSpan 文字の表示間隔を0にして一斉描画
func zeroDrawSpan() {
	mainMutex.Lock()
	defer mainMutex.Unlock()
	drawerChan <- 0
}

//recoveryDrawSpan 文字の表示間隔を元に戻す
func recoveryDrawSpan() {
	mainMutex.Lock()
	mainMutex.Unlock()
	drawerChan <- drawSpan
}

//changeDrawSpan 文字の表示間隔を変更
func changeDrawSpan(c int) {
	drawSpan = c
	recoveryDrawSpan()
}

//setCell 文字をセットする
func setCell(x, y int, c rune, fg, bg termbox.Attribute) {
	runebaff <- cell{
		x,
		y,
		c,
		fg,
		bg,
	}
}

//drawLine 指定位置にStringを表示(文字サイズによって全角半角判別可能)
func drawLine(s string, x, y int, fg, bg termbox.Attribute) {
	mainMutex.Lock()
	defer mainMutex.Unlock()
	i := 0
	for _, r := range s {
		setCell(x+i, y, r, fg, bg)
		i += runewidth.RuneWidth(r)
	}
}

//drawRow yに指定した縦座標に横一列を指定の文字列で埋め尽くす
func drawRow(s string, y int, fg, bg termbox.Attribute) {
	mainMutex.Lock()
	mainMutex.Unlock()
	i := 0
	for {
		drawLine(s, i, y, fg, bg)

		if i >= width {
			//横幅以上に書き込んだら
			break
		}
		i += runewidth.StringWidth(s)
	}
}

//drawLineWithPosition stringを指定した表示方法で表示
func drawLineWithPosition(s string, x, y int, fg, bg termbox.Attribute, p DrawPosition) {

	w := width                     //横幅取得
	sw := runewidth.StringWidth(s) //文字列長さ取得
	if sw > w {
		//文字列のほうが長いので一部のみを表示
		diff := sw - w - 1
		i := 0
		for _, r := range s[diff:] {
			setCell(x+i, y, r, fg, bg)
			i += runewidth.RuneWidth(r)
		}

		switch p {
		case LeftDrawn:
			//左字詰めで表示
			i := 0
			for _, r := range s {
				setCell(x+i, y, r, fg, bg)
				i += runewidth.RuneWidth(r)
			}
		case RightDrawn:
			//右字詰めで表示
			i := w - sw
			for _, r := range s {
				setCell(x+i, y, r, fg, bg)
				i += runewidth.RuneWidth(r)
			}
		case CenterDrawn:
			//中央に表示
			i := int(math.Trunc(float64((w - sw) / 2))) //切り捨て
			for _, r := range s {
				setCell(x+i, y, r, fg, bg)
			}

		}

	}
}

//drawLinePrompt drawLineを瞬時に描画する
func drawLinePrompt(s string, x, y int, fg, bg termbox.Attribute) {
	zeroDrawSpan()
	drawLine(s, x, y, fg, bg)
	recoveryDrawSpan()
}

//drawRowPrompt drawRowを瞬時に描画する
func drawRowPrompt(s string, y int, fg, bg termbox.Attribute) {
	zeroDrawSpan()
	drawRow(s, y, fg, bg)
	recoveryDrawSpan()
}
