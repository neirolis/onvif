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
	IPv4     net.IP
	IPv6     net.IP
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

		ipv4, ipv6 := lookupXaddrs(doc)
		hardware, name, location := lookupScopes(doc)

		devices = append(devices, Device{
			IPv4:     ipv4,
			IPv6:     ipv6,
			Hardware: hardware,
			Name:     name,
			Location: location,
		})
	}

	return
}

// lookup xaddrs by path ./Body/ProbeMatches/ProbeMatch/XAddrs
// ex: <d:XAddrs>http://${IPv4-addr}/onvif/device_service http://[${IPv6-addr}]/onvif/device_service</d:XAddrs>
func lookupXaddrs(doc *etree.Document) (ipv4, ipv6 net.IP) {
	elem := doc.Root().FindElement("./Body/ProbeMatches/ProbeMatch/XAddrs")

	for _, xaddr := range strings.Split(elem.Text(), " ") {
		u, err := url.Parse(xaddr)
		if err != nil {
			continue
		}

		ip := net.ParseIP(u.Hostname())
		switch {
		case ip.To4() != nil:
			ipv4 = ip
		case ip.To16() != nil:
			ipv6 = ip
		}
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
