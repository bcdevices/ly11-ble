package h4

import (
	"fmt"
	"io"
	//"net"
	//"os"
	//"sync"
	"context"
	"time"

	"github.com/google/gousb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewUsb(ctx *gousb.Context) (io.ReadWriteCloser, error) {
	logrus.Debugf("opening h4 usb...")

	fast := time.Millisecond * 500
	rwc, err := NewUsbRWC(ctx, fast)
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	eofAsError := true
	if err := resetAndWaitIdle(rwc, time.Second*2, eofAsError); err != nil {
		rwc.Close()
		return nil, errors.Wrap(err, "resetAndWaitIdle")
	}

	fmt.Printf("reset and wait idle complete.\n")

	h := &h4{
		rwc:     rwc,
		done:    make(chan int),
		rxQueue: make(chan []byte, rxQueueSize),
		txQueue: make(chan []byte, txQueueSize),
	}
	h.frame = newFrame(h.rxQueue)

	go h.rxLoop(eofAsError)

	return h, nil
}

type usbRWC struct {
	ctx       *gousb.Context
	usbDev    *gousb.Device
	intf      *gousb.Interface
	inEp      *gousb.InEndpoint
	outEp     *gousb.OutEndpoint
	inStream  *gousb.ReadStream
	outStream *gousb.WriteStream

	timeout time.Duration
}

func (u *usbRWC) Read(p []byte) (int, error) {
	fmt.Printf("usbRWC: read (buf:%d)...\n", len(p))

	opCtx := context.Background()
	opCtx, done := context.WithTimeout(opCtx, u.timeout)
	defer done()

	//n, err := u.inEp.ReadContext(opCtx, p)
	n, err := u.inStream.ReadContext(opCtx, p)
	select {
	case <-opCtx.Done():
		fmt.Printf("usbRWC: read complete. (timeout)\n")

		return 0, nil
	}
	fmt.Printf("usbRWC: read complete.(n=%v, err=%v)\n", n, err)
	return n, err
}

func (u *usbRWC) Write(p []byte) (int, error) {
	fmt.Printf("usbRWC: write(%v)\n", p)
	opCtx := context.Background()
	opCtx, done := context.WithTimeout(opCtx, u.timeout)
	defer done()
	//n, err := u.outEp.WriteContext(opCtx, p)
	n, err := u.outStream.WriteContext(opCtx, p)
	fmt.Printf("usbRWC: write complete.\n")
	return n, err
}

func (u *usbRWC) Close() error {
	fmt.Printf("usbRWC: close\n")
	u.intf.Close()
	if err := u.usbDev.Close(); err != nil {
		return fmt.Errorf("close USB dev: %w", err)
	}
	return nil
}

func NewUsbRWC(ctx *gousb.Context, timeout time.Duration) (*usbRWC, error) {
	const usbVendorId uint16 = 0x2fe3
	const usbProductId uint16 = 0x000c

	usbDev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(usbVendorId),
		gousb.ID(usbProductId))
	if err != nil {
		return nil, fmt.Errorf("open USB device: %w", err)
	}

	if usbDev == nil {
		return nil, fmt.Errorf("device %s not found")
	}

	// Automatically detach any kernel driver and
	// reattach it when releasing the interface.
	err = usbDev.SetAutoDetach(true)
	if err != nil {
		_ = usbDev.Close()

		return nil, fmt.Errorf("set auto-detach: %w", err)
	}

	intf, _, err := usbDev.DefaultInterface()
	if err != nil {
		_ = usbDev.Close()

		return nil, fmt.Errorf("claim intf: %w", err)
	}

	inEp, err := intf.InEndpoint(0x81)
	if err != nil {
		intf.Close()
		_ = usbDev.Close()

		return nil, fmt.Errorf("claim in Ep: %w", err)
	}

	outEp, err := intf.OutEndpoint(0x01)
	if err != nil {
		intf.Close()
		_ = usbDev.Close()

		return nil, fmt.Errorf("claim out Ep: %w", err)
	}

	inStream, err := inEp.NewStream(64, 8)
	if err != nil {
		intf.Close()
		_ = usbDev.Close()

		return nil, fmt.Errorf("prepare input stream: %w", err)
	}

	outStream, err := outEp.NewStream(64, 8)
	if err != nil {
		intf.Close()
		_ = usbDev.Close()

		return nil, fmt.Errorf("prepare output stream: %w", err)
	}

	return &usbRWC{
		ctx:       ctx,
		usbDev:    usbDev,
		intf:      intf,
		inEp:      inEp,
		outEp:     outEp,
		inStream:  inStream,
		outStream: outStream,
		timeout:   timeout,
	}, nil
}
