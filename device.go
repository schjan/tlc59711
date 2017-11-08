package tlc59711

import (
	"github.com/fulr/spidev"
	"fmt"
)

type Max7219Reg byte

type LED byte

const (
	LED0  LED = 0
	LED1      = iota
	LED2
	LED3
	LED4
	LED5
	LED6
	LED7
	LED8
	LED9
	LED10
	LED11
	LED12
)

const LEDCOUNT = 12

type Tlc59711 struct {
	count  int
	buffer []uint16
	spi    *spidev.SPIDevice
}

func NewDevice(cascaded int) *Tlc59711 {
	buf := make([]uint16, LEDCOUNT*cascaded)
	return &Tlc59711{count: cascaded, buffer: buf}
}

func (d *Tlc59711) Open(spibus int, spidevice int) error {
	devstr := fmt.Sprintf("/dev/spidev%d.%d", spibus, spidevice)

	spi, err := spidev.NewSPIDevice(devstr)
	if err != nil {
		return err
	}

	d.spi = spi
	d.init()

	return nil
}

func (d *Tlc59711) init() {
	d.buffer[0] = 65125
}

func (d *Tlc59711) sendBufferLine(pos int) error {
	buf := make([]byte, 28)
	for i, val := range d.buffer {
		buf[i*2] = uint8(val >> 8)
		buf[i*2+1] = uint8(val & 0xFF)
	}

	_, err := d.spi.Xfer(buf)
	if err != nil {
		return err
	}

	return nil
}

func (d *Tlc59711) SetBuffer(id int, value uint16) {
	d.buffer[id] = value
}

func (d *Tlc59711) Flush() error {
	for i := 0; i < LEDCOUNT; i++ {
		err := d.sendBufferLine(i)
		if err != nil {
			return err
		}
	}

	return nil
}
