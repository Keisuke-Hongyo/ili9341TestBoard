package TochPad

import (
	"machine"
)

const (
	cmd_rdx        = 0xD1
	cmd_rdy        = 0x91
	cmd_rdz1       = 0xB1
	cmd_rdz2       = 0xC1
	touch_MIN_X_AD = 400
	touch_MIN_Y_AD = 400
	touch_MAX_Y_AD = 3900
	touch_MAX_X_AD = 3700
)

type Device struct {
	spi   *machine.SPI
	t_cs  machine.Pin
	t_irq machine.Pin

	frequency uint32

	precision uint8

	MinXAd uint16
	MaxXAd uint16
	MinYAd uint16
	MaxYAd uint16
}

type Config struct {
	Precision uint8
}

func New(spi *machine.SPI, t_cs, t_irq machine.Pin) *Device {
	return &Device{
		precision: 10,
		spi:       spi,
		t_cs:      t_cs,
		t_irq:     t_irq,
		frequency: 2000000,
		MaxXAd:    touch_MAX_X_AD,
		MaxYAd:    touch_MAX_Y_AD,
		MinXAd:    touch_MIN_X_AD,
		MinYAd:    touch_MIN_Y_AD,
	}
}

func (d *Device) Configure(config *Config) {

	if config.Precision == 0 {
		d.precision = 10
	} else {
		d.precision = config.Precision
	}

	d.t_cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.t_irq.Configure(machine.PinConfig{Mode: machine.PinInput})

	d.t_cs.High()
}

// タッチセンサ割込ピン出力設定
func (d *Device) SetTouchIqr() {
	f := d.spi.GetBaudRate()
	defer d.spi.SetBaudRate(f)

	d.spi.SetBaudRate(d.frequency)
	d.t_irq.Low()
	d.spi.Transfer(0x80)
	d.t_cs.High()
}

// 割込ピンの出力確認
func (d *Device) GetTouch() bool {
	return d.t_irq.Get()
}

// タッチパネルの座標の取得
func (d *Device) GetPos() (xAd uint16, yAd uint16) {
	var dt []byte
	dt = make([]byte, 2)
	f := d.spi.GetBaudRate()
	defer d.spi.SetBaudRate(f)

	d.spi.SetBaudRate(d.frequency)

	d.t_cs.Low()
	d.spi.Transfer(cmd_rdy)
	dt[0], _ = d.spi.Transfer(0x00)
	dt[1], _ = d.spi.Transfer(0x00)
	yAd = (uint16(dt[0])<<8 | uint16(dt[1])) >> 3

	d.spi.Transfer(cmd_rdx)
	dt[0], _ = d.spi.Transfer(0x00)
	dt[1], _ = d.spi.Transfer(0x00)
	xAd = (uint16(dt[0])<<8 | uint16(dt[1])) >> 3

	d.spi.Transfer(0x80)

	d.t_cs.High()

	if xAd >= d.MaxXAd {
		xAd = d.MaxXAd
	} else if xAd <= d.MinXAd {
		xAd = d.MinXAd
	}

	if yAd >= d.MaxYAd {
		yAd = d.MaxYAd
	} else if yAd <= d.MinYAd {
		yAd = d.MinYAd
	}

	return xAd, yAd
}
