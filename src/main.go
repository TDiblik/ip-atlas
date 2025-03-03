package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"lukechampine.com/uint128"
)

const DIR_PERMISSIONS fs.FileMode = 0755
const FILE_PERMISSIONS fs.FileMode = 0655

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	IS_PRODUCTION := os.Getenv("IP_ATLAS_PRODUCTION") == "TRUE"
	if IS_PRODUCTION {
		log.Println("Running in production mode.")

		log.Println("Started downloading ip2asn-combined.")
		_, err := grab.Get(".", "https://iptoasn.com/data/ip2asn-combined.tsv.gz")
		panic_on_err("Unable to download file ip2asn-combined.tsv.gz: ", err)
		log.Println("Finished downloading ip2asn-combined.")
	} else {
		log.Println("Running in development mode. Make sure you downloaded ip2asn-combined.tsv.gz beforehand.")
	}

	log.Println("Starting to unzip.")
	gzipped_file, err := os.Open("./ip2asn-combined.tsv.gz")
	panic_on_err("Unable to open file ip2asn-combined.tsv.gz: ", err)
	defer gzipped_file.Close()

	gzip_reader, err := gzip.NewReader(gzipped_file)
	panic_on_err("Error creating gzip reader: ", err)
	defer gzip_reader.Close()

	ip2asn_info_raw, err := io.ReadAll(gzip_reader)
	panic_on_err("Unable to read from gzip reader: ", err)

	if !IS_PRODUCTION {
		err = os.WriteFile("./ip2asn-combined.tsv", ip2asn_info_raw, FILE_PERMISSIONS)
		panic_on_err("Unable to write output to a file: ", err)
	}
	log.Println("Succesfully unziped.")

	log.Println("Starting preprocessing data.")
	asn_map := make(map[uint32]*Company)
	ip2asn_info := strings.SplitSeq(string(ip2asn_info_raw), "\n")
	for info_row := range ip2asn_info {
		info_parts := strings.Split(info_row, "\t")
		if len(info_parts) == 1 {
			continue
		}
		from_ip := net.ParseIP(info_parts[0]).To16()
		to_ip := net.ParseIP(info_parts[1]).To16()
		if from_ip.IsPrivate() || to_ip.IsPrivate() {
			continue
		}

		company_asn_raw, err := strconv.ParseUint(info_parts[2], 10, 32)
		panic_on_err("Unable to parse company asn: ", err)

		company_asn := uint32(company_asn_raw)
		_, exists := asn_map[company_asn]
		if !exists {
			asn_map[company_asn] = &Company{
				Name:                      strings.TrimSpace(info_parts[4]),
				ASN:                       company_asn,
				TotalNumberOfIPs_v4:       0,
				TotalNumberOfIPs_v6:       uint128.Zero,
				TotalNumberOfIPs_combined: uint128.Zero,
				OwnedIpRanges_v4:          make([]IPRange, 0),
				OwnedIpRanges_v6:          make([]IPRange, 0),
				CountryCode:               strings.TrimSpace(info_parts[3]),
			}
		}
		company := asn_map[company_asn]
		new_ip_range := IPRange{
			FromIP: from_ip,
			ToIP:   to_ip,
		}
		number_of_ips_in_range := NumberOfIPsInRange(from_ip, to_ip)
		if from_ip.To4() != nil {
			company.OwnedIpRanges_v4 = append(company.OwnedIpRanges_v4, new_ip_range)
			company.TotalNumberOfIPs_v4 += uint32(number_of_ips_in_range.Big().Uint64()) // Impossible to overflow
		} else {
			company.OwnedIpRanges_v6 = append(company.OwnedIpRanges_v6, new_ip_range)
			company.TotalNumberOfIPs_v6 = company.TotalNumberOfIPs_v6.Add(number_of_ips_in_range)
		}
		company.TotalNumberOfIPs_combined = company.TotalNumberOfIPs_combined.Add(number_of_ips_in_range)
	}
	log.Println("Done preprocessing data.")

	log.Println("Start creating output files.")

	// Prepare directories
	os.RemoveAll("./dist")
	os.Mkdir("./dist", DIR_PERMISSIONS)
	os.Mkdir("./dist/company", DIR_PERMISSIONS)
	copy_file("./templates/globals.css", "./dist/globals.css")
	copy_file("./templates/api.html", "./dist/api.html")

	// Write company files (json + html page)
	company_file_raw, err := os.ReadFile("./templates/company.html")
	panic_on_err("Unable to read ./templates/company.html: ", err)
	company_file_contents := string(company_file_raw)
	for _, value_raw := range asn_map {
		json_value, err := json.Marshal(value_raw)
		panic_on_err(fmt.Sprint("Unable to parse value into json : ", value_raw), err)
		err = os.WriteFile(fmt.Sprint("./dist/company/", value_raw.ASN, ".json"), json_value, FILE_PERMISSIONS)
		panic_on_err("Unable to write value into a file: ", err)
		write_company_file(value_raw, company_file_contents)
	}

	// Write ipv4, ipv6 and combined chart
	asn_name_arr := make([]ASNKeyNameMap, 0, len(asn_map))
	sorted_asn := make([]*Company, 0, len(asn_map))
	for key, value := range asn_map {
		sorted_asn = append(sorted_asn, value)
		asn_name_arr = append(asn_name_arr, ASNKeyNameMap{
			ASN:  key,
			Name: value.Name,
		})
	}
	asn_key_name_map_json, err := json.Marshal(asn_name_arr)
	panic_on_err("Unable to parse asn_key_name_map into a json file: ", err)
	err = os.WriteFile("./dist/company/key_name_map.json", asn_key_name_map_json, FILE_PERMISSIONS)
	panic_on_err("Unable to write asn_key_name_map_json value into a file: ", err)
	chart_rows_ip_v4 := create_chart_string(sorted_asn, 0)
	chart_rows_ip_v6 := create_chart_string(sorted_asn, 1)
	chart_rows_ip_combined := create_chart_string(sorted_asn, 2)

	index_file_raw, err := os.ReadFile("./templates/index.html")
	panic_on_err("Unable to read ./templates/index.html: ", err)
	index_file_contents := string(index_file_raw)

	write_index_file("index", index_file_contents, chart_rows_ip_v4)
	write_index_file("ipv6", index_file_contents, chart_rows_ip_v6)
	write_index_file("combined", index_file_contents, chart_rows_ip_combined)

	log.Println("Done creating output files.")
}

// Both have to be in 16 byte representation!
func NumberOfIPsInRange(starting, ending net.IP) uint128.Uint128 {
	from_ip := uint128.FromBytesBE(starting)
	to_ip := uint128.FromBytesBE(ending)

	return to_ip.Sub(from_ip).Add64(1)
}

func copy_file(file_path, dest_path string) {
	file, err := os.ReadFile(file_path)
	panic_on_err(fmt.Sprint("Unable to read file ", file_path), err)
	err = os.WriteFile(dest_path, file, FILE_PERMISSIONS)
	panic_on_err(fmt.Sprint("Unable to write file ", file_path, " to ", dest_path), err)
}

// sort_by == 0 => ipv4
// sort_by == 1 => ipv6
// sort_by == 2 => combined
// sort_by == default => panic
func create_chart_string(asns []*Company, sort_by uint) string {
	sort.Slice(asns, func(i, j int) bool {
		return get_total_based_on_sort(asns[i], sort_by).Cmp(get_total_based_on_sort(asns[j], sort_by)) == 1
	})

	number_of_all_ips := uint128.Max
	if sort_by == 0 {
		number_of_all_ips = uint128.From64(uint64(4294967296)) // 2^32
	}

	is_max_percentage_not_routed := asns[0].ASN != 0
	max_percentage := calc_percentage(get_total_based_on_sort(asns[0], sort_by), number_of_all_ips, is_max_percentage_not_routed)

	var chart_rows_to_append strings.Builder
	for i, value := range asns {
		total := get_total_based_on_sort(value, sort_by)
		if total.Cmp(uint128.Zero) == 0 {
			continue
		}

		is_not_routed := value.ASN != 0
		percentage_boosted := calc_percentage(total, number_of_all_ips, is_not_routed)
		percentage_weighted := float64(percentage_boosted.Div(max_percentage).Big().Uint64())
		percentage_normalized := float64(percentage_boosted.Big().Uint64())
		if i == 0 {
			percentage_weighted = percentage_weighted * 100
			percentage_normalized = percentage_normalized / 10_000_000
		} else {
			percentage_weighted = percentage_weighted / 100
			percentage_normalized = percentage_normalized / 100_000_000_000
		}
		percentage_formatted := fmt.Sprint(percentage_normalized)
		if percentage_normalized == 0.0 {
			percentage_formatted = "< 1e-11"
		}
		chart_rows_to_append.WriteString(fmt.Sprint(
			"<div class=\"chart-row\">",
			"<div class=\"chart-label\"> <a href=\"/company/", value.ASN, ".html\" target=\"_blank\">", value.Name, "</a></div>",
			"<div class=\"chart-bar\">",
			"<div class=\"chart-bar-internal\" style=\"width: ", percentage_weighted, "%\"></div>",
			"</div>",
			"<div class=\"chart-percentage\">", percentage_formatted, "% (", total, ")", "</div>",
			"</div>",
		))
	}
	return chart_rows_to_append.String()
}

func get_total_based_on_sort(company *Company, sort_by uint) uint128.Uint128 {
	var total uint128.Uint128
	switch sort_by {
	case 0:
		total = uint128.From64(uint64(company.TotalNumberOfIPs_v4))
	case 1:
		total = company.TotalNumberOfIPs_v6
	case 2:
		total = company.TotalNumberOfIPs_combined
	}
	return total
}

func calc_percentage(total, max uint128.Uint128, is_not_routed bool) uint128.Uint128 {
	if is_not_routed {
		total = total.Mul64(10_000)
	}
	return total.Div(max.Div64(1_000_000_000))
}

func write_index_file(filename, original_index_file_contents, chart_rows string) {
	err := os.WriteFile(
		fmt.Sprint("./dist/", filename, ".html"),
		[]byte(
			strings.Replace(
				original_index_file_contents,
				"{{ INSERT_NEW_ROWS }}",
				chart_rows,
				1,
			),
		),
		FILE_PERMISSIONS,
	)
	panic_on_err("Unable to write index file: ", err)
}

func write_company_file(company *Company, original_company_file_contents string) {
	out := original_company_file_contents
	out = strings.Replace(out, "{{ INSERT_NAME }}", company.Name, 1)
	out = strings.Replace(out, "{{ INSERT_ASN }}", fmt.Sprint(company.ASN), 1)
	out = strings.Replace(out, "{{ INSERT_COUNTRY_CODE }}", company.CountryCode, 1)
	out = strings.Replace(out, "{{ INSERT_TOTAL_NUMBER_OF_IP4s }}", fmt.Sprint(company.TotalNumberOfIPs_v4), 1)
	out = strings.Replace(out, "{{ INSERT_TOTAL_NUMBER_OF_IP6s }}", company.TotalNumberOfIPs_v6.String(), 1)
	out = strings.Replace(out, "{{ INSERT_TOTAL_NUMBER_OF_IPs_COMBINED }}", company.TotalNumberOfIPs_combined.String(), 1)

	var ipv4sToAppend strings.Builder
	for _, r := range company.OwnedIpRanges_v4 {
		ipv4sToAppend.WriteString(fmt.Sprint("<li>", r.FromIP, " - ", r.ToIP, "</li>"))
	}
	out = strings.Replace(out, "{{ INSERT_IPV4_RANGES }}", ipv4sToAppend.String(), 1)

	var ipv6sToAppend strings.Builder
	for _, r := range company.OwnedIpRanges_v6 {
		ipv6sToAppend.WriteString(fmt.Sprint("<li>", r.FromIP, " - ", r.ToIP, "</li>"))
	}
	out = strings.Replace(out, "{{ INSERT_IPV6_RANGES }}", ipv6sToAppend.String(), 1)

	err := os.WriteFile(fmt.Sprint("./dist/company/", company.ASN, ".html"), []byte(out), FILE_PERMISSIONS)
	panic_on_err("Unable to write company file: ", err)
}

func panic_on_err(msg string, err error) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
