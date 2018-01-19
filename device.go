package tlc59711

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
	"time"
)

type Max7219Reg byte

type LED byte

const (
	LED0 LED = 0
	LED1     = iota
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
	port    spi.PortCloser
	conn    spi.Conn
	spistr  string
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
	_, err := host.Init()
	if err != nil {
		log.WithError(err).Fatal()
	}

	d.spistr = fmt.Sprintf("/dev/spidev%d.%d", spibus, spidevice)

	port, err := spireg.Open(d.spistr)
	if err != nil {
		return err
	}
	d.port = port

	conn, err := port.Connect(int64(500000), spi.Mode3, 8)
	if err != nil {
		return err
	}
	d.conn = conn

	d.init()

	return nil
}

func (d *Tlc59711) worker() {
	for true {
		select {
		case <-d.abort:
			return
		case buf := <-d.toWrite:
			err := d.conn.Tx(buf, nil)
			if err != nil {
				log.Errorf("Error sending Datapacket to %v. Shutdown worker. : %v", d.spistr, err)
				return
			}
			time.Sleep(4 * time.Microsecond)
		}
	}
}

func (d *Tlc59711) Close() {
	d.abort <- struct{}{}

	if d.port == nil {
		return
	}

	err := d.port.Close()

	log.Warn(err)
}

func (d *Tlc59711) init() {
	d.buffer[0] = 65125
}

func (d *Tlc59711) sendBufferLine(pos int) error {
	BCr := uint32(0x7F)
	BCg := uint32(0x7F)
	BCb := uint32(0x7F)

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
	buf[0] = uint8(command >> 24)
	buf[1] = uint8(command >> 16)
	buf[2] = uint8(command >> 8)
	buf[3] = uint8(command & 0xFF)

	for i := 0; i < 12; i++ {
		bufI := i*2 + 4
		dbufI := pos*12 + i
		buf[bufI] = uint8(d.buffer[dbufI] >> 8)
		buf[bufI+1] = uint8(d.buffer[dbufI] & 0xFF)
	}

	//for i, val := range d.buffer {
	//	buf[i*2+4] = uint8(val >> 8)
	//	buf[i*2+5] = uint8(val & 0xFF)
	//}

	d.toWrite <- buf

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
