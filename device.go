package tlc59711

import (
	"fmt"
	"golang.org/x/exp/io/spi"
	log "github.com/Sirupsen/logrus"
	"time"
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
	count   int
	buffer  []uint16
	spi     *spi.Device
	toWrite chan []byte
	abort   chan struct{}
}

func NewDevice(cascaded int) *Tlc59711 {
	buf := make([]uint16, LEDCOUNT*cascaded)
	dev := &Tlc59711{count: cascaded, buffer: buf, toWrite: make(chan []byte, 1000), abort: make(chan struct{}, 1)}

	go dev.worker()

	return dev
}

func (d *Tlc59711) Open(spibus int, spidevice int) error {

	devstr := fmt.Sprintf("/dev/spidev%d.%d", spibus, spidevice)
	dev, err := spi.Open(&spi.Devfs{
		Dev:      devstr,
		Mode:     spi.Mode3,
		MaxSpeed: 500000,
	})
	if err != nil {
		return err
	}
	err = dev.SetBitOrder(spi.MSBFirst)
	if err != nil {
		return err
	}

	d.spi = dev
	d.init()

	return nil
}

func (d *Tlc59711) worker() {
	for true {
		select {
		case <-d.abort:
			return
		case buf := <-d.toWrite:
			d.spi.Tx(buf, nil)
			time.Sleep(4 * time.Microsecond)
		}
	}
}

func (d *Tlc59711) Close() {
	d.abort <- struct{}{}

	if d.spi == nil {
		return
	}

	err := d.spi.Close()

	log.Warn(err)
}

func (d *Tlc59711) init() {
	d.buffer[0] = 65125
}

func (d *Tlc59711) sendBufferLine(pos int) error {
	BCr := uint32(0x7F)
	BCg := BCr
	BCb := BCg

	command := uint32(0x25)
	command <<= 5
	//OUTTMG = 1, EXTGCK = 0, TMGRST = 1, DSPRPT = 1, BLANK = 0 -> 0x16
	command |= 0x16
	command <<= 7
	command |= BCr
	command <<= 7
	command |= BCg
	command <<= 7
	command |= BCb

	buf := make([]byte, 28)
	for i, val := range d.buffer {
		buf[i*2+4] = uint8(val >> 8)
		buf[i*2+5] = uint8(val & 0xFF)
	}

	buf[0] = uint8(command >> 24)
	buf[1] = uint8(command >> 16)
	buf[2] = uint8(command >> 8)
	buf[3] = uint8(command & 0xFF)

	d.toWrite <- buf
	//err := d.spi.Tx(buf, nil)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (d *Tlc59711) SetBuffer(id int, value uint16) {
	d.buffer[id] = value
}

func (d *Tlc59711) Flush() error {
	err := d.sendBufferLine(0)
	if err != nil {
		return err
	}

	//for i := 0; i < LEDCOUNT; i++ {
	//	err := d.sendBufferLine(i)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}
