package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/gousb"
	"github.com/pkg/errors"
	"github.com/rigado/ble"
	"github.com/rigado/ble/examples/lib/dev"
)

var (
	du  = flag.Duration("du", 5*time.Second, "scanning duration")
	dup = flag.Bool("dup", true, "allow duplicate reported")
)

// Bus 001 Device 014: ID 2fe3:000b ZEPHYR USB-DEV

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := gousb.NewContext()
	defer ctx.Close()

	opt := ble.OptTransportH4Usb(ctx)

	d, err := dev.NewDevice("", opt)
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// Scan for specified durantion, or until interrupted by user.
	fmt.Printf("Scanning for %s...\n", *du)
	bgCtx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	chkErr(ble.Scan(bgCtx, *dup, advHandler, nil))
}

func advHandler(a ble.Advertisement) {
	if a.Connectable() {
		fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}
	fmt.Printf("\n")
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}
