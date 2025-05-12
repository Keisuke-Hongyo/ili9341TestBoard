package GT20L16J1Y

import (
	"fmt"
	//"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"machine"
	"time"
)

// フォントサイズ
const (
	FontSize2Byte       = 32
	FontSize2ByteWidth  = 16
	FontSize2ByteHeight = 16

	FontSize1Byte       = 16
	FontSize1ByteWidth  = 8
	FontSize1ByteHeight = 16
)

var err error

type Device struct {
	spi       *machine.SPI // Digital Input	bus
	csn       *machine.Pin // Digital Input	SPI Chip Select
	frequency uint32
}

type Font struct {
	FontHeight uint16
	FontWidth  uint16
	FontData   []uint16
}

type Fonts []Font

// New SPIの割り当て
func New(spi *machine.SPI, csn *machine.Pin) *Device {
	return &Device{
		spi:       spi,
		csn:       csn,
		frequency: 16000000,
	}
}

// 初期化
func (d *Device) Initialize() {

	// CSピンの設定
	d.csn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.csn.High() // start position

}

func (d *Device) ReadFonts(str string) Fonts {
	var sjisStr []byte
	var fontsdata Fonts
	var cnt int

	f := d.spi.GetBaudRate()
	defer d.spi.SetBaudRate(f)

	_ = d.spi.SetBaudRate(d.frequency)

	fontsdata = make(Fonts, 0)

	// 文字コードの変換
	t := japanese.ShiftJIS.NewEncoder()
	sjisStr, _, err = transform.Bytes(t, []byte(str))

	ln := len(sjisStr)

	cnt = 0
	for cnt < ln {
		if (sjisStr[cnt] < 0x80) || ((0xa0 < sjisStr[cnt]) && (sjisStr[cnt] <= 0xdF)) {
			// 半角処理
			fontsdata = append(fontsdata, d.readFontAscii(sjisStr[cnt]))
			cnt++
		} else {
			// 全角処理
			fontsdata = append(fontsdata, d.readFontJIS(uint16(sjisStr[cnt])<<8|uint16(sjisStr[cnt+1])))
			cnt += 2
		}
	}
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 24000000,
	})
	return fontsdata
}

// Local Function

// 半角データ読み込み
func (d *Device) readFontAscii(asciicode uint8) Font {

	var address uint32
	var i uint8
	var data Font
	var dt []byte

	data.FontHeight = FontSize1ByteHeight
	data.FontWidth = FontSize1ByteWidth
	data.FontData = make([]uint16, FontSize1ByteWidth)
	dt = make([]byte, FontSize1Byte)

	if asciicode >= 0x20 && asciicode <= 0x7F {
		address = (uint32(asciicode)-0x20)*16 + 255968
	}

	//フォントデータ読み込み
	d.csn.Low()
	time.Sleep(10 * time.Microsecond) // 待ち時間

	// 読み込みアドレス指定
	_, _ = d.spi.Transfer(0x03)
	_, _ = d.spi.Transfer(byte(address >> 16 & 0xff))
	_, _ = d.spi.Transfer(byte(address >> 8 & 0xff))
	_, _ = d.spi.Transfer(byte(address & 0xff))

	// フォントデータ読み込み
	for i = 0; i < FontSize1Byte; i++ {
		dt[i], _ = d.spi.Transfer(0x00)
	}

	d.csn.High()

	// 16bitデータに変換
	for i = 0; i < FontSize1ByteWidth; i++ {
		data.FontData[i] = uint16(dt[i+FontSize1ByteWidth])<<8 + uint16(dt[i])
	}
	return data
}

// 全角データ読み込み
func (d *Device) readFontJIS(code uint16) Font {
	var c1, c2 uint8
	var i uint8
	var msb, lsb uint32
	var address uint32
	var data Font
	var dt []byte

	// UTF-8->SJIS
	c1 = uint8((code & 0xff00) >> 8)
	c2 = uint8(code & 0x00ff)

	if c1 >= 0xe0 {
		c1 = c1 - 0x40
	}
	if c2 >= 0x80 {
		c2 = c2 - 1
	}
	if c2 >= 0x9e {
		c1 = (c1 - 0x70) * 2
		c2 = c2 - 0x7d
	} else {
		c1 = ((c1 - 0x70) * 2) - 1
		c2 = c2 - 0x1f
	}

	/*jisxの区点を求める*/
	msb = uint32(c1) - 0x20 //区
	lsb = uint32(c2) - 0x20 //点

	/*JISの句点番号で分類*/
	address = 0

	/*各種記号・英数字・かな(一部機種依存につき注意,㍍などWindowsと互換性なし)*/
	if msb >= 1 && msb <= 15 && lsb >= 1 && lsb <= 94 {
		address = ((msb-1)*94 + (lsb - 1)) * 32
	}

	/*第一水準*/
	if msb >= 16 && msb <= 47 && lsb >= 1 && lsb <= 94 {
		address = ((msb-16)*94+(lsb-1))*32 + 43584
	}

	/*第二水準*/
	if msb >= 48 && msb <= 84 && lsb >= 1 && lsb <= 94 {
		address = ((msb-48)*94+(lsb-1))*32 + 138464
	}

	/*GT20L16J1Y内部では1区と同等の内容が収録されている*/
	if msb == 85 && lsb >= 0x01 && lsb <= 94 {
		address = ((msb-85)*94+(lsb-1))*32 + 246944
	}

	/*GT20L16J1Y内部では2区、3区と同等の内容が収録されている*/
	if msb >= 88 && msb <= 89 && lsb >= 1 && lsb <= 94 {
		address = ((msb-88)*94+(lsb-1))*32 + 249952
	}

	data.FontWidth = FontSize2ByteWidth
	data.FontHeight = FontSize2ByteHeight
	data.FontData = make([]uint16, FontSize2ByteWidth)
	dt = make([]byte, FontSize2Byte)
	/*漢字ROMにデータを送信*/
	//フォントデータ読み込み
	d.csn.Low()
	time.Sleep(10 * time.Microsecond) // 待ち時間

	// 読み込みアドレス指定
	_, _ = d.spi.Transfer(0x03)
	_, _ = d.spi.Transfer(byte(address >> 16 & 0xff))
	_, _ = d.spi.Transfer(byte(address >> 8 & 0xff))
	_, _ = d.spi.Transfer(byte(address & 0xff))

	// フォントデータ読み込み
	for i = 0; i < FontSize2Byte; i++ {
		dt[i], _ = d.spi.Transfer(0x00)
	}
	d.csn.High()

	// 16bitデータに変換
	for i = 0; i < FontSize2ByteWidth; i++ {
		data.FontData[i] = uint16(dt[i+FontSize2ByteWidth])<<8 + uint16(dt[i])
	}
	return data
}

func (d *Device) PrintTerminal(fontsData Fonts) {
	for i := 0; i < len(fontsData); i++ {
		// Font Data Output
		printfont(fontsData[i])
	}
}

// ターミナル表示用(半角)
func printfont(data Font) {
	var x, y uint16
	for y = 0; y < data.FontHeight; y++ {
		for x = 0; x < data.FontWidth; x++ {
			if data.FontData[x]&(0x01<<y) != 0x00 {
				fmt.Printf("xx")
			} else {
				fmt.Printf("--")
			}
		}
		fmt.Printf("\n")
	}
}
