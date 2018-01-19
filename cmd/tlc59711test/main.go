package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/schjan/tlc59711"
	"time"
)

func main() {
	dev := tlc59711.NewDevice(1)

	err := dev.Open(0, 0)
	defer dev.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	for true {
		bla := 0
		start := time.Now()
		log.Info("start")
		for i := uint16(0); i < 65505; i += 10 {
			SetAllValue(dev, i)
			bla++
		}
		elapsed := time.Since(start)

		log.Infof("%v steps in %v", bla, elapsed)
	}
}

func SetAllValue(dev *tlc59711.Tlc59711, value uint16) {
	for i := 0; i < 12; i++ {
		dev.SetBuffer(i, value)
	}
	dev.Flush()
}
