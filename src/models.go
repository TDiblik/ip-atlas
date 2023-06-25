package main

import (
	"net"

	"lukechampine.com/uint128"
)

type Company struct {
	Name                      string
	ASN                       uint32
	TotalNumberOfIPs_v4       uint32
	TotalNumberOfIPs_v6       uint128.Uint128
	TotalNumberOfIPs_combined uint128.Uint128
	OwnedIpRanges_v4          []IPRange
	OwnedIpRanges_v6          []IPRange
	CountryCode               string
}

type IPRange struct {
	FromIP net.IP
	ToIP   net.IP
}

type ASNKeyNameMap struct {
	ASN  uint32
	Name string
}
