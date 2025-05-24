package mdns

import (
	"golang.org/x/net/dns/dnsmessage"
)

// BuildPTRRecord returns a PTR record for service discovery
func BuildPTRRecord(serviceType, instanceName dnsmessage.Name) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  serviceType,
			Type:  dnsmessage.TypePTR,
			Class: dnsmessage.ClassINET,
			TTL:   120,
		},
		Body: &dnsmessage.PTRResource{PTR: instanceName},
	}
}

// BuildSRVRecord returns an SRV record for the service instance
func BuildSRVRecord(instanceName, hostname dnsmessage.Name, port uint16) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  instanceName,
			Type:  dnsmessage.TypeSRV,
			Class: dnsmessage.ClassINET,
			TTL:   120,
		},
		Body: &dnsmessage.SRVResource{
			Priority: 0,
			Weight:   0,
			Port:     port,
			Target:   hostname,
		},
	}
}

// BuildTXTRecord returns a TXT record for the service instance
func BuildTXTRecord(instanceName dnsmessage.Name, txt []string) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  instanceName,
			Type:  dnsmessage.TypeTXT,
			Class: dnsmessage.ClassINET,
			TTL:   120,
		},
		Body: &dnsmessage.TXTResource{TXT: txt},
	}
}

// BuildARecord returns an A record for the hostname
func BuildARecord(hostname dnsmessage.Name, ip [4]byte) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  hostname,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   120,
		},
		Body: &dnsmessage.AResource{A: ip},
	}
}
