package main

import (
	"github.com/schjan/tlc59711"
	"time"
	log "github.com/Sirupsen/logrus"
)

func main() {
	dev := tlc59711.NewDevice(1)

	dev.Open(0, 0)

	for true {
		dev.SetBuffer(0, 65535)
		dev.Flush()
		log.Info("Hell")
		time.Sleep(1 * time.Second)
		dev.SetBuffer(0, 0)
		dev.Flush()
		log.Info("Dunkel")
		time.Sleep(1 * time.Second)
	}
}
