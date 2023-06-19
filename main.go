package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/cavaliergopher/grab/v3"
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

    ip2asn_info := strings.Split(string(ip2asn_info_raw), "\n")
    for _, info_row := range ip2asn_info {
        info_parts := strings.Split(info_row, "\t")
        if len(info_parts) == 1 {
            continue;
        }
        starting_ip := net.ParseIP(info_parts[0])
        ending_ip := net.ParseIP(info_parts[1])
        fmt.Println(starting_ip, " - ", ending_ip)
    }
}

func panic_on_err(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
        panic("Forcefully panicking for the reason above.");
	}
}