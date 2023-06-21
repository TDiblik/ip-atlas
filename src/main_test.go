package main

import (
	"net"
	"testing"

	"lukechampine.com/uint128"
)

func TestNumberOfIPsBetweenRangeCalculation(t *testing.T) {
    var tests = [] struct {from net.IP; to net.IP; expected uint64} {
		{net.ParseIP("0.0.0.0").To16(), net.ParseIP("0.0.0.255").To16(), 256},
		{net.ParseIP("0.0.0.0").To16(), net.ParseIP("0.0.1.255").To16(), 512},
		{net.IPv4zero.To16(), net.ParseIP("0.0.2.255").To16(), 768},
	}
	
	for _, test := range tests {
		if test.from == nil || test.to == nil {
			t.Fatal("Unable to parse ip -- from: ", test.from, " to: ", test.to)
		}
		ips_between := NumberOfIPsInRange(test.from, test.to)
		if ips_between.Cmp(uint128.From64(test.expected)) != 0 {
			t.Fatal("Number of IPs between ", test.from ," and ", test.to, " is supposed to be ", test.expected, " but is ", ips_between)
		}
	}
}