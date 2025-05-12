package lcdProc

import (
	"ili9341/GT20L16J1Y"
	"ili9341/TochPad"
	"image/color"
	"machine"
	"tinygo.org/x/drivers/ili9341"
)

type LcdBoard struct {
	Lcd   *ili9341.Device
	Touch *TochPad.Device
	Font  *GT20L16J1Y.Device
	Xpos  uint16
	Ypos  uint16
}

func New(
	spi *machine.SPI, lcdDc machine.Pin, lcdCsn machine.Pin, lcdRst machine.Pin,
	touchCsn machine.Pin, touchIrq machine.Pin, fntCsn machine.Pin) *LcdBoard {
	d := ili9341.NewSPI(spi, lcdDc, lcdCsn, lcdRst)
	d.Configure(ili9341.Config{})

	t := TochPad.New(spi, touchCsn, touchIrq)
	t.Configure(&TochPad.Config{Precision: 10})

	f := GT20L16J1Y.New(spi, &fntCsn)
	f.Initialize()

	return &LcdBoard{
		Lcd:   d,
		Touch: t,
		Font:  f,
	}
}

func (d *LcdBoard) GetTouch() bool {
	return d.Touch.GetTouch()
}

func (d *LcdBoard) GetPos() (x uint16, y uint16) {

	width, height := d.Lcd.Size()

	xAd, yAd := d.Touch.GetPos()

	switch d.Lcd.Rotation() {
	case ili9341.Rotation0:
		x = uint16(int(width) - int(width)*(int(d.Touch.MaxXAd)-int(xAd))/(int(d.Touch.MaxXAd)-int(d.Touch.MinXAd)))

		y = uint16(int(height) * (int(d.Touch.MaxYAd) - int(yAd)) / (int(d.Touch.MaxYAd) - int(d.Touch.MinYAd)))
		break

	case ili9341.Rotation90:
		x = uint16(int(width) * (int(d.Touch.MaxYAd) - int(yAd)) / (int(d.Touch.MaxYAd) - int(d.Touch.MinYAd)))
		y = uint16(int(height) * (int(d.Touch.MaxXAd) - int(xAd)) / (int(d.Touch.MaxXAd) - int(d.Touch.MinXAd)))
		break

	case ili9341.Rotation270:
		x = uint16(int(width) - (int(width) * (int(d.Touch.MaxYAd) - int(yAd)) / (int(d.Touch.MaxYAd) - int(d.Touch.MinYAd))))
		y = uint16(int(height) - (int(height) * (int(d.Touch.MaxXAd) - int(xAd)) / (int(d.Touch.MaxXAd) - int(d.Touch.MinXAd))))
		break
	case ili9341.Rotation180:
		x = uint16(int(width) * (int(d.Touch.MaxXAd) - int(xAd)) / (int(d.Touch.MaxXAd) - int(d.Touch.MinXAd)))
		y = uint16(int(height) - (int(height) * (int(d.Touch.MaxYAd) - int(yAd)) / (int(d.Touch.MaxYAd) - int(d.Touch.MinYAd))))
		break
	}

	return x, y
}

func (d *LcdBoard) LcdPrint(x uint16, y uint16, str string, fg color.RGBA, bg color.RGBA) {
	d.Xpos = x // set position X
	d.Ypos = y // set position Y
	d.printText(str, fg, bg)
}

func (d *LcdBoard) printText(str string, fg color.RGBA, bg color.RGBA) {
	var f GT20L16J1Y.Fonts
	tmp := d.Xpos
	f = d.Font.ReadFonts(str)
	for i := 0; i < len(f); i++ {
		// Font Data Output
		d.printChar(f[i], fg, bg)
		d.Xpos += f[i].FontWidth
	}
	d.Xpos = tmp
}

func (d *LcdBoard) printChar(font GT20L16J1Y.Font, fg color.RGBA, bg color.RGBA) {
	var x, y uint16
	for y = 0; y < font.FontHeight; y++ {
		for x = 0; x < font.FontWidth; x++ {
			if font.FontData[x]&(0x01<<y) != 0x00 {
				d.Lcd.SetPixel(int16(x+d.Xpos), int16(y+d.Ypos), fg)
			} else {
				d.Lcd.SetPixel(int16(x+d.Xpos), int16(y+d.Ypos), bg)
			}
		}
	}
}
