package onvif

import (
	"net"
	"testing"
)

func TestDiscovery(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, iface := range ifaces {
		switch {
		case iface.Flags&net.FlagUp == 0: // skip down interface
			continue
		case iface.Flags&net.FlagLoopback != 0: // skip loopback
			continue
		case iface.Flags&net.FlagPointToPoint != 0: // skip point-to-point
			continue
		}

		devices, err := Discovery(&iface)
		if err != nil {
			t.Error(err)
		}
		if len(devices) > 0 {
			t.Logf("%+v", devices)
			return
		}
	}

	t.Error("addresses not found")
}
