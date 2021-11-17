package h4

import (
	"fmt"
	//"io"
	//"net"
	//"os"
	//"sync"
	//"time"

	"github.com/google/gousb"
	//"github.com/jacobsa/go-serial/serial"
	//"github.com/pkg/errors"
	//"github.com/sirupsen/logrus"
)

// io.ReadWriteCloser

type usbRWC struct {
	ctx *gousb.Context
}

func (u *usbRWC) Read(p []byte) (int, error) {
	fmt.Printf("usbRWC: read\n")
	return 0, nil
}

func (u *usbRWC) Write(p []byte) (int, error) {
	fmt.Printf("usbRWC: write\n")
	return 0, nil
}

func (u *usbRWC) Close() error {
	fmt.Printf("usbRWC: close\n")
	return nil
}

func NewUsbRWC(ctx *gousb.Context) (*usbRWC, error) {
	return &usbRWC{
		ctx: ctx,
	}, nil
}
