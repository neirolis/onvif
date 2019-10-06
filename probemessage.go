/*******************************************************
 * Copyright (C) 2018 Palanjyan Zhorzhik
 *
 * This file is part of WS-Discovery project.
 *
 * WS-Discovery can be copied and/or distributed without the express
 * permission of Palanjyan Zhorzhik
 *******************************************************/

package onvif

import (
	"net"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/yakovlevdmv/gosoap"
	"golang.org/x/net/ipv4"
)

const bufSize = 8192

func buildProbeMessage(uuidV4 string, scopes, types []string, nmsp map[string]string) gosoap.SoapMessage {
	namespaces := make(map[string]string)
	namespaces["a"] = "http://schemas.xmlsoap.org/ws/2004/08/addressing"

	probeMessage := gosoap.NewEmptySOAP()

	probeMessage.AddRootNamespaces(namespaces)

	// header
	var headerContent []*etree.Element

	action := etree.NewElement("a:Action")
	action.SetText("http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe")
	action.CreateAttr("mustUnderstand", "1")

	msgID := etree.NewElement("a:MessageID")
	msgID.SetText("uuid:" + uuidV4)

	replyTo := etree.NewElement("a:ReplyTo")
	replyTo.CreateElement("a:Address").SetText("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous")

	to := etree.NewElement("a:To")
	to.SetText("urn:schemas-xmlsoap-org:ws:2005:04:discovery")
	to.CreateAttr("mustUnderstand", "1")

	headerContent = append(headerContent, action, msgID, replyTo, to)
	probeMessage.AddHeaderContents(headerContent)

	// body
	probe := etree.NewElement("Probe")
	probe.CreateAttr("xmlns", "http://schemas.xmlsoap.org/ws/2005/04/discovery")

	if len(types) != 0 {
		typesTag := etree.NewElement("d:Types")
		if len(nmsp) != 0 {
			for key, value := range nmsp {
				typesTag.CreateAttr("xmlns:"+key, value)
			}
		}
		typesTag.CreateAttr("xmlns:d", "http://schemas.xmlsoap.org/ws/2005/04/discovery")
		var typesString string
		for _, j := range types {
			typesString += j
			typesString += " "
		}

		typesTag.SetText(strings.TrimSpace(typesString))

		probe.AddChild(typesTag)
	}

	if len(scopes) != 0 {
		scopesTag := etree.NewElement("d:Scopes")
		var scopesString string
		for _, j := range scopes {
			scopesString += j
			scopesString += " "
		}
		scopesTag.SetText(strings.TrimSpace(scopesString))

		probe.AddChild(scopesTag)
	}

	probeMessage.AddBodyContent(probe)

	return probeMessage
}

func sendUDPMulticast(iface *net.Interface, msg []byte) (result []string, err error) {
	group := net.IPv4(239, 255, 255, 250)

	c, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)

	if err := p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
		return result, err
	}

	if err := p.SetMulticastInterface(iface); err != nil {
		return result, err
	}
	p.SetMulticastTTL(2)
	if _, err := p.WriteTo(msg, nil, &net.UDPAddr{IP: group, Port: 3702}); err != nil {
		return result, err
	}

	if err := p.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return result, err
	}

	for {
		b := make([]byte, bufSize)
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			break
		}
		result = append(result, string(b[0:n]))
	}

	return result, nil
}
