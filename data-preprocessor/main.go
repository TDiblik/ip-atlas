package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"lukechampine.com/uint128"
)

func main() {
    IS_PRODUCTION := os.Getenv("IP_ATLAS_PRODUCTION") == "TRUE"
    if IS_PRODUCTION {
        fmt.Println("Running in production mode.")

        fmt.Println("Started downloading ip2asn-combined.")
        _, err := grab.Get(".", "https://iptoasn.com/data/ip2asn-combined.tsv.gz")
        if err != nil {
            panic(err)
        }
        fmt.Println("Finished downloading ip2asn-combined.")
    } else {
        fmt.Println("Running in development mode. Make sure you downloaded ip2asn-combined.tsv.gz beforehand.")
    }
    
    fmt.Println("Starting to unzip.")
    gzippedFile, err := os.Open("./ip2asn-combined.tsv.gz")
    panic_on_err("Unable to open file ip2asn-combined.tsv.gz: ", err)
	defer gzippedFile.Close()

	gzipReader, err := gzip.NewReader(gzippedFile)
    panic_on_err("Error creating gzip reader: ", err)
	defer gzipReader.Close()

    ip2asn_info_raw, err := io.ReadAll(gzipReader)
    panic_on_err("Unable to read from gzip reader: ", err)
    
    if !IS_PRODUCTION {
        err = os.WriteFile("./ip2asn-combined.tsv", ip2asn_info_raw, 0644)
        panic_on_err("Unable to write output to a file: ", err)
    }
    fmt.Println("Succesfully unziped.")

    fmt.Println("Starting preprocessing data.")
    asn_map := make(map[uint32]*Company)
    ip2asn_info := strings.Split(string(ip2asn_info_raw), "\n")
    for _, info_row := range ip2asn_info {
        info_parts := strings.Split(info_row, "\t")
        if len(info_parts) == 1 {
            continue;
        }
        from_ip := net.ParseIP(info_parts[0]).To16()
        to_ip := net.ParseIP(info_parts[1]).To16()
        
        company_asn_raw, err := strconv.ParseUint(info_parts[2], 10, 32);
        panic_on_err("Unable to parse company asn: ", err)
        
        company_asn := uint32(company_asn_raw)
        _, exists := asn_map[company_asn]
        if !exists {
            asn_map[company_asn] = &Company{
                Name:             strings.TrimSpace(info_parts[4]),
                ASN:              uint32(company_asn),
                TotalNumberOfIPs: 0,
                OwnedIpRanges:    make([]IPRange, 0),
                CountryCode:      strings.TrimSpace(info_parts[3]),
            }
        }
        company := asn_map[company_asn]
        company.OwnedIpRanges = append(company.OwnedIpRanges, IPRange {
        	FromIP: from_ip,
        	ToIP:   to_ip,
        })
        company.TotalNumberOfIPs += Calc_number_of_ips_between(from_ip, to_ip).Big().Uint64()
    }
    
    // TODO: go over each key and create .json and .html for each company and fill in html templates (also, todo, create html templates lol xd).
    var keys [] uint32
    for k, _ := range asn_map {
        keys = append(keys, k)
        // fmt.Println(k)
    }
    fmt.Println(len(keys))

    fmt.Println("Done preprocessing data.")
}

type Company struct {
    Name string
    ASN uint32
    TotalNumberOfIPs uint64 // Let's hope no single company owns more than 50% :D
    OwnedIpRanges []IPRange
    CountryCode string
}

type IPRange struct {
    FromIP net.IP
    ToIP net.IP
}

// Both have to be in 16 byte representation!
func Calc_number_of_ips_between(starting, ending net.IP) (uint128.Uint128) {
	from_ip := uint128.FromBytesBE(starting)
	to_ip := uint128.FromBytesBE(ending)

	return to_ip.Sub(from_ip).Add64(1)
}

func panic_on_err(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
        panic("Forcefully panicking for the reason above.");
	}
}