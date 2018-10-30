package main

import (
	"github.com/schjan/tlc59711"
	"log"
	"time"
)

func main() {
	dev := tlc59711.NewDevice(2)

	err := dev.Open(0, 0)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dev.Close()
	defer SetAllValue(dev, 0)

	dev.EnableAutoflush()

	// dim up and down
	for {
		steps := 0
		start := time.Now()
		log.Println("start")
		for i := uint16(0); i < 65505; i += 20 {
			SetAllValue(dev, i)
			time.Sleep(10 * time.Millisecond)
			steps++
		}
		for i := uint16(65500); i > 0; i -= 20 {
			SetAllValue(dev, i)
			time.Sleep(10 * time.Millisecond)
			steps++
		}
		elapsed := time.Since(start)

		log.Printf("%v steps in %v/n", steps, elapsed)
	}

	//// blink
	//for {
	//	start := time.Now()
	//	log.Println("start")
	//	SetAllValue(dev, 0)
	//
	//	time.Sleep(500 * time.Millisecond)
	//	SetAllValue(dev, 65505)
	//	time.Sleep(500 * time.Millisecond)
	//	elapsed := time.Since(start)
	//	log.Printf("in %v/n", elapsed)
	//}
}

func SetAllValue(dev *tlc59711.Tlc59711, value uint16) {
	for i := 0; i < 12; i++ {
		dev.SetBuffer(i, value)
	}
	for i := 12; i < 24; i++ {
		dev.SetBuffer(i, 65505-value)
	}
}
