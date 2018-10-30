package tlc59711

import (
	"fmt"
	"github.com/pkg/errors"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
	"sync"
	"time"
)

const (
	LEDCOUNT = 12
	MAXVALUE = uint16(65535)
	MINVALUE = uint16(0)
)

type Tlc59711 struct {
	count            int
	buffer           []uint16
	port             spi.PortCloser
	conn             spi.Conn
	spistr           string
	toWrite          chan []byte
	abort            chan struct{}
	autoflushOnce    sync.Once
	autoflushEnabled bool
	isDirtyVar       bool
	isDirtySync      *sync.Mutex
	BCr, BCg, BCb    uint32
}

func NewDevice(cascaded int) *Tlc59711 {
	buf := make([]uint16, LEDCOUNT*cascaded)
	dev := &Tlc59711{
		count:       cascaded,
		buffer:      buf,
		toWrite:     make(chan []byte),
		abort:       make(chan struct{}),
		isDirtySync: &sync.Mutex{},
		isDirtyVar:  true}

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

	conn, err := port.Connect(physic.Frequency(8000000), spi.Mode3, 8)
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
	if d.autoflushEnabled {
		return errors.New("autoflushing is enabled, so this call will do nothing")
	}

	return d.flush()
}

func (d *Tlc59711) flush() error {
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

	d.isDirtySync.Lock()
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
	d.isDirtyVar = false
	d.isDirtySync.Unlock()

	err := d.conn.Tx(buf, nil)
	if err != nil {
		return errors.Wrap(err, "transmitting to tlc59711 failed")
	}
	time.Sleep(4 * time.Nanosecond)

	return nil
}

func (d *Tlc59711) EnableAutoflush() error {
	d.autoflushOnce.Do(func() {
		d.autoflushEnabled = true
		go d.autoflush()
	})

	return nil
}

func (d *Tlc59711) autoflush() {
	for {
		if d.isDirty() {
			d.flush()
		} else {
			select {
			case <-time.After(100 * time.Millisecond):
				continue
			case <-d.abort:
				return
			}
		}
	}
}

func (d *Tlc59711) isDirty() bool {
	d.isDirtySync.Lock()
	defer d.isDirtySync.Unlock()

	return d.isDirtyVar
}

func (d *Tlc59711) SetBuffer(id int, value uint16) {
	d.isDirtySync.Lock()
	d.buffer[id] = value
	d.isDirtyVar = true
	d.isDirtySync.Unlock()
}
