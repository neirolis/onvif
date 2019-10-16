package onvif

import (
	"bytes"
	"net"
	"net/url"
	"path"
	"sort"
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

// Discovery onvif devices by type NetworkVideoTransmitter
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

		// remove from name hardware value
		name = strings.Replace(name, hardware, "", -1)
		name = strings.TrimSpace(name)

		devices = append(devices, Device{
			XAddr:    addr,
			Hardware: hardware,
			Name:     name,
			Location: location,
		})
	}

	sort.Slice(devices, func(i, j int) bool {
		ip1 := net.ParseIP(devices[i].XAddr)
		ip2 := net.ParseIP(devices[j].XAddr)

		return bytes.Compare(ip1, ip2) < 0
	})

	return
}

// lookup xaddrs by path ./Body/ProbeMatches/ProbeMatch/XAddrs
// ex: <d:XAddrs>http://${IPv4-addr}/onvif/device_service http://[${IPv6-addr}]/onvif/device_service  ... </d:XAddrs>
func lookupXaddr(doc *etree.Document) (xaddr string, ok bool) {
	elem := doc.Root().FindElement("./Body/ProbeMatches/ProbeMatch/XAddrs")
	if elem == nil {
		return
	}

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
	if elem == nil {
		return
	}

	for _, scope := range strings.Split(elem.Text(), " ") {
		u, err := url.Parse(scope)
		if err != nil {
			continue
		}

		upath := strings.ToLower(u.Path)

		switch {
		case strings.Contains(upath, "hardware"):
			_, hardware = path.Split(u.Path)
		case strings.Contains(upath, "name"):
			_, name = path.Split(u.Path)
		case strings.Contains(upath, "location"):
			_, location = path.Split(u.Path)
		}
	}

	return
}
