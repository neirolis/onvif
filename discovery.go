package onvif

import (
	"net"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/uuid"
)

func Discovery(iface *net.Interface) (xaddrs []string, err error) {
	msg := buildProbeMessage(uuid.New().String(), nil, []string{"dn:NetworkVideoTransmitter"}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})

	resp, err := sendUDPMulticast(iface, []byte(msg.String()))
	if err != nil {
		return xaddrs, err
	}
	for _, j := range resp {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			continue
		}

		endpoints := doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch/XAddrs")
		for _, xaddr := range endpoints {
			xaddr := strings.Split(strings.Split(xaddr.Text(), " ")[0], "/")[2]
			xaddrs = append(xaddrs, xaddr)
		}
	}
	return
}
