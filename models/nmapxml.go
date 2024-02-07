package models

import "encoding/xml"

// NmapXML represents the top-level structure of the Nmap XML output
type NmapXML struct {
	XMLName xml.Name      `xml:"nmaprun"`
	Hosts   []NmapXMLHost `xml:"host"`
}

// NmapXMLHost represents a host in the Nmap XML output
type NmapXMLHost struct {
	Ports []NmapXMLPort `xml:"ports>port"`
}

// NmapXMLPort represents a port in the Nmap XML output
type NmapXMLPort struct {
	Protocol string         `xml:"protocol,attr"`
	PortID   string         `xml:"portid,attr"`
	State    NmapXMLState   `xml:"state"`
	Service  NmapXMLService `xml:"service"`
}

// NmapXMLState represents the state of a port in the Nmap XML output
type NmapXMLState struct {
	State string `xml:"state,attr"`
}

// NmapXMLService represents the service of a port in the Nmap XML output
type NmapXMLService struct {
	Name string `xml:"name,attr"`
}
