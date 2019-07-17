# ONVIF discovery

It contains only one method `Discovery` for search devices by ONVIF protocol in the local network. 

## Install

```go
go get github.com/neirolis/onvif
```


## Usage

```go
  iface, err := net.InterfaceByName(name)
  // if err ...

	addrs, err := onvif.Discovery(iface)
	// if err ...
	// ex: addrs = [192.168.1.10, 192.168.1.12, 192.168.1.33, 192.168.1.123, ...] 
```