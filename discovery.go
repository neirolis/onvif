package onvif

import (
	"net"
	"net/url"
	"path"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/uuid"
)

type Device struct {
	XAddr    string
	Hardware string
	Name     string
	Location string
}

//Discovery onvif devices by type NetworkVideoTransmitter
func Discovery(iface *net.Interface) (devices []Device, err error) {
	msg := buildProbeMessage(uuid.New().String(), nil, []string{"dn:NetworkVideoTransmitter"}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})

	resp, err := sendUDPMulticast(iface, []byte(msg.String()))
	if err != nil {
		return nil, err
	}

	for _, j := range resp {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			continue
		}

		addr, ok := lookupXaddr(doc)
		if !ok {
			continue
		}

		hardware, name, location := lookupScopes(doc)

		devices = append(devices, Device{
			XAddr:    addr,
			Hardware: hardware,
			Name:     name,
			Location: location,
		})
	}

	return
}

// lookup xaddrs by path ./Body/ProbeMatches/ProbeMatch/XAddrs
// ex: <d:XAddrs>http://${IPv4-addr}/onvif/device_service http://[${IPv6-addr}]/onvif/device_service  ... </d:XAddrs>
func lookupXaddr(doc *etree.Document) (xaddr string, ok bool) {
	elem := doc.Root().FindElement("./Body/ProbeMatches/ProbeMatch/XAddrs")

	for _, onvifurl := range strings.Split(elem.Text(), " ") {
		u, err := url.Parse(onvifurl)
		if err != nil {
			continue
		}

		xaddr = u.Hostname()
		ok = true
		break
	}

	return
}

// lookup scopes by path ./Body/ProbeMatches/ProbeMatch/Scopes
// ex: <d:Scopes>onvif://www.onvif.org/type/video_encoder onvif://www.onvif.org/hardware/DS-2CD2042WD-I onvif://www.onvif.org/name/HIKVISION%20DS-2CD2042WD-I</d:Scopes>
func lookupScopes(doc *etree.Document) (hardware, name, location string) {
	elem := doc.Root().FindElement("./Body/ProbeMatches/ProbeMatch/Scopes")

	for _, scope := range strings.Split(elem.Text(), " ") {
		u, err := url.Parse(scope)
		if err != nil {
			continue
		}

		switch {
		case strings.Contains(u.Path, "hardware"):
			_, hardware = path.Split(u.Path)
		case strings.Contains(u.Path, "name"):
			_, name = path.Split(u.Path)
		case strings.Contains(u.Path, "location"):
			_, location = path.Split(u.Path)
		}
	}

	return
}
