package tlc59711

import (
	"fmt"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
	"time"
)

const LEDCOUNT = 12

type Tlc59711 struct {
	count         int
	buffer        []uint16
	port          spi.PortCloser
	conn          spi.Conn
	spistr        string
	toWrite       chan []byte
	abort         chan struct{}
	BCr, BCg, BCb uint32
}

func NewDevice(cascaded int) *Tlc59711 {
	buf := make([]uint16, LEDCOUNT*cascaded)
	dev := &Tlc59711{count: cascaded, buffer: buf, toWrite: make(chan []byte), abort: make(chan struct{})}

	dev.BCr = 0x7F
	dev.BCg = 0x7F
	dev.BCb = 0x7F

	return dev
}

func (d *Tlc59711) Open(spibus int, spidevice int) error {
	_, err := host.Init()
	if err != nil {
		return err
	}

	d.spistr = fmt.Sprintf("/dev/spidev%d.%d", spibus, spidevice)

	port, err := spireg.Open(d.spistr)
	if err != nil {
		return err
	}
	d.port = port

	conn, err := port.Connect(int64(8000000), spi.Mode3, 8)
	if err != nil {
		return err
	}
	d.conn = conn

	d.init()

	return nil
}

func (d *Tlc59711) Close() error {
	d.abort <- struct{}{}

	if d.port == nil {
		return nil
	}

	err := d.port.Close()

	return err
}

func (d *Tlc59711) init() {
	d.buffer[0] = 65125
}

func (d *Tlc59711) Flush() error {
	command := uint32(0x25)
	command <<= 5
	//OUTTMG = 1, EXTGCK = 0, TMGRST = 1, DSPRPT = 1, BLANK = 0 -> 0x16
	command |= 0x16
	command <<= 7
	command |= d.BCr
	command <<= 7
	command |= d.BCg
	command <<= 7
	command |= d.BCb

	buf := make([]byte, 28*d.count)
	for dev := 0; dev < d.count; dev++ {
		buf[0+dev*28] = uint8(command >> 24)
		buf[1+dev*28] = uint8(command >> 16)
		buf[2+dev*28] = uint8(command >> 8)
		buf[3+dev*28] = uint8(command & 0xFF)

		for i := 0; i < 12; i++ {
			bufI := i*2 + 4
			dbufI := dev*12 + i
			buf[bufI+dev*28] = uint8(d.buffer[dbufI] >> 8)
			buf[bufI+1+dev*28] = uint8(d.buffer[dbufI] & 0xFF)
		}
	}

	d.conn.Tx(buf, nil)
	time.Sleep(4 * time.Nanosecond)

	return nil
}

func (d *Tlc59711) SetBuffer(id int, value uint16) {
	d.buffer[id] = value
}
