package main

import (
	"fmt"
	"ili9341/lcdProc"
	"machine"
	"time"
	"tinygo.org/x/drivers/ili9341"
)

func timer1ms(ch chan<- bool) {
	for {
		time.Sleep(1 * time.Millisecond)
		ch <- true
	}
}

func main() {
	var cnt uint8 = 0
	var str string

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 24000000,
	})

	display := lcdProc.New(
		machine.SPI0, // SPI Bus
		machine.GP15, // DC
		machine.GP17, // CS
		machine.GP14, // Reset
		machine.GP12, // Touch CSN
		machine.GP13, // Tocuh IRQ
		machine.GP11, // KanjiFont CSN
	)

	display.Touch.SetTouchIqr()

	display.Lcd.SetRotation(ili9341.Rotation90)

	width, height := display.Lcd.Size()
	display.Lcd.FillRectangle(0, 0, width/2, height/2, lcdProc.White)
	display.Lcd.FillRectangle(width/2, 0, width/2, height/2, lcdProc.Red)
	display.Lcd.FillRectangle(0, height/2, width/2, height/2, lcdProc.Green)
	display.Lcd.FillRectangle(width/2, height/2, width/2, height/2, lcdProc.Blue)
	display.Lcd.FillRectangle(width/4, height/4, width/2, height/2, lcdProc.Black)
	ch := make(chan bool)
	go timer1ms(ch)

	for {
		select {
		case <-ch:
			cnt++
			break

		}
		str = fmt.Sprintf("w = %3d h=%3d", width, height)
		display.LcdPrint(100, 100, str, lcdProc.Orange, lcdProc.Black)
		str := fmt.Sprintf("カウント = %3d", cnt)
		display.LcdPrint(100, 130, str, lcdProc.White, lcdProc.Black)

		if display.GetTouch() {
			str = fmt.Sprintf("Not tocuched ")
			display.LcdPrint(100, 150, str, lcdProc.White, lcdProc.Black)
			display.LcdPrint(120, 200, "ここをタッチ", lcdProc.Red, lcdProc.Black)
		} else {
			x, y := display.GetPos()
			str = fmt.Sprintf("x=%4d y=%4d", x, y)
			display.LcdPrint(100, 150, str, lcdProc.White, lcdProc.Black)
			if (x > 120 && x < 220) && (y > 200 && x < 216) {
				// 文字の範囲内
				display.LcdPrint(120, 200, "タッチ！！！", lcdProc.White, lcdProc.Black)
			} else {
				// 文字の範囲外
				display.LcdPrint(120, 200, "ここをタッチ", lcdProc.Red, lcdProc.Black)
			}
		}

	}

}
