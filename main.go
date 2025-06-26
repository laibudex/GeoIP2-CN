package main

import (
	"bufio"
	"flag"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

var (
	srcFile      string
	dstFile      string
	databaseType string

	// 中国记录
	cnRecord = mmdbtype.Map{
		"country": mmdbtype.Map{
			"geoname_id":           mmdbtype.Uint32(1814991),
			"is_in_european_union": mmdbtype.Bool(false),
			"iso_code":             mmdbtype.String("CN"),
			"names": mmdbtype.Map{
				"de":    mmdbtype.String("China"),
				"en":    mmdbtype.String("China"),
				"es":    mmdbtype.String("China"),
				"fr":    mmdbtype.String("Chine"),
				"ja":    mmdbtype.String("中国"),
				"pt-BR": mmdbtype.String("China"),
				"ru":    mmdbtype.String("Китай"),
				"zh-CN": mmdbtype.String("中国"),
			},
		},
	}

	// 中非共和国（CF）记录
	cfRecord = mmdbtype.Map{
		"country": mmdbtype.Map{
			"geoname_id":           mmdbtype.Uint32(1814989),
			"is_in_european_union": mmdbtype.Bool(false),
			"iso_code":             mmdbtype.String("CF"),
			"names": mmdbtype.Map{
				"de":    mmdbtype.String("Centralafrikanische Republik"),
				"en":    mmdbtype.String("Central African Republic"),
				"es":    mmdbtype.String("República Centroafricana"),
				"fr":    mmdbtype.String("République centrafricaine"),
				"ja":    mmdbtype.String("中央アフリカ共和国"),
				"pt-BR": mmdbtype.String("República Centro-Africana"),
				"ru":    mmdbtype.String("Центральноафриканская Республика"),
				"zh-CN": mmdbtype.String("中非共和国"),
			},
		},
	}
)

func init() {
	flag.StringVar(&srcFile, "s", "ip_list.txt", "specify source IP list file (supports [CN] / [CF] sections)")
	flag.StringVar(&dstFile, "d", "Country.mmdb", "specify destination mmdb file")
	flag.StringVar(&databaseType, "t", "GeoIP2-Country", "specify MaxMind database type")
	flag.Parse()
}

func main() {
	writer, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: databaseType,
			RecordSize:   24,
		},
	)
	if err != nil {
		log.Fatalf("failed to create writer: %v", err)
	}

	fh, err := os.Open(srcFile)
	if err != nil {
		log.Fatalf("failed to open %s: %v", srcFile, err)
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	currentRecord := cnRecord // 默认使用 CN

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 区段头，例如 [CN] 或 [CF]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sec := strings.ToUpper(line[1 : len(line)-1])
			switch sec {
			case "CN":
				currentRecord = cnRecord
			case "CF":
				currentRecord = cfRecord
			default:
				log.Warnf("unknown section %s, skip", sec)
			}
			continue
		}

		// 其余行，视为 CIDR，解析并插入
		_, network, err := net.ParseCIDR(line)
		if err != nil {
			log.Fatalf("invalid CIDR %s: %v", line, err)
		}
		if err := writer.Insert(network, currentRecord); err != nil {
			log.Fatalf("failed to insert %s: %v", line, err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading %s: %v", srcFile, err)
	}

	outFh, err := os.Create(dstFile)
	if err != nil {
		log.Fatalf("failed to create output file %v", err)
	}
	defer outFh.Close()

	if _, err := writer.WriteTo(outFh); err != nil {
		log.Fatalf("failed to write to %s: %v", dstFile, err)
	}

	log.Infof("successfully wrote mmdb to %s", dstFile)
}
