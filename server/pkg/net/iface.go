package net

import (
	"errors"
	"net"
	"strings"
)

func FirstIFace() (net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	if len(ifaces) < 1 {
		return net.Interface{}, errors.New("couldn't find any network interfaces")
	}

	return ifaces[0], nil
}

func FindIFace(name string) (net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	if len(ifaces) < 1 {
		return net.Interface{}, errors.New("couldn't find any network interfaces")
	}

	name = strings.ToLower(name)
	for _, iface := range ifaces {
		if strings.Contains(strings.ToLower(iface.Name), name) {
			return iface, nil
		}
	}

	return net.Interface{}, errors.New("found no iface that matches name: " + name)
}