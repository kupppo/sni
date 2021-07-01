package fxpakpro

import (
	"context"
	"encoding/binary"
	"fmt"
)

func (d *Device) put(ctx context.Context, space space, address uint32, data []byte) (err error) {
	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(space)
	sb[6] = byte(FlagNONE)

	// put the data size in:
	size := uint32(len(data))
	binary.BigEndian.PutUint32(sb[252:], size)

	// put the address in:
	binary.BigEndian.PutUint32(sb[256:], address)

	if shouldLock(ctx) {
		defer d.lock.Unlock()
		d.lock.Lock()
	}

	// send the data to the USB port:
	err = sendSerial(d.f, sb)
	if err != nil {
		_ = d.Close()
		return
	}

	dest := sb[0:]
	for len(data) > 0 {
		var n int
		for i := range dest {
			dest[i] = 0
		}
		n = copy(dest, data)
		data = data[n:]

		err = sendSerial(d.f, sb)
		if err != nil {
			_ = d.Close()
			return
		}
	}

	// await single response:
	err = recvSerial(d.f, sb, 512)
	if err != nil {
		_ = d.Close()
		return
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		_ = d.Close()
		return fmt.Errorf("put: fxpakpro response packet does not contain USBA header")
	}
	if ec := sb[5]; ec != 0 {
		return fmt.Errorf("put: %w", fxpakproError(ec))
	}

	return
}