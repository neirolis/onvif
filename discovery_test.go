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

		addrs, err := Discovery(&iface)
		if err != nil {
			t.Error(err)
		}
		if len(addrs) > 0 {
			t.Log(addrs)
			return
		}
	}

	t.Error("addresses not found")
}
