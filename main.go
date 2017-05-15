package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sjug/go-logparse/stats"
)

var searchDir, resultDir string

type host struct {
	kind      string
	resultDir string
	results   []resultType
}

type resultType struct {
	kind, path           string
	min, max, avg, pct95 float64
}

var fileHeader = map[string][]string{
	"disk_IOPS.csv":                      {"vda-write"},
	"cpu_usage_percent_cpu.csv":          {"openshift_start_master_api", "openshift_start_master_controll", "openshift_start_node"},
	"memory_usage_resident_set_size.csv": {"openshift_start_master_api", "openshift_start_master_controll", "openshift_start_node"},
	"network_l2_network_packets_sec.csv": {"eth0-rx", "eth0-tx"},
	"network_l2_network_Mbits_sec.csv":   {"eth0-rx", "eth0-tx"},
}

func initFlags() {
	flag.StringVar(&searchDir, "i", "/home/jugs/work/20170330_2Knodes_5Kprojects_nvme_etcd_a/tools-default/", "pbench run result directory to parse")
	flag.StringVar(&resultDir, "o", "/tmp/", "directory to output parsed CSV result data")
	flag.Parse()
}

func main() {
	var hosts []host
	hostRegex := regexp.MustCompile(`svt[_-][elmn]\w*[_-]\d`)
	//searchDir := "/home/jugs/work/pbench-user-benchmark_foo_2017-04-11_16:51:07/1/reference-result/tools-default/"
	initFlags()

	// Return director listing of searchDir
	dirList, err := ioutil.ReadDir(searchDir)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over directory contents
	for _, item := range dirList {
		// Match subdirectory that follows our pattern
		if hostRegex.MatchString(item.Name()) && item.IsDir() {
			kind := strings.Split(item.Name(), ":")
			newHost := host{
				kind:      kind[0],
				resultDir: searchDir + item.Name(),
			}
			hosts = append(hosts, newHost)
		}
	}

	// Maps are not ordered, create ordered slice of keys
	keys := []string{}
	for k, _ := range fileHeader {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Iterate over all known hosts
	for i, host := range hosts {
		fmt.Printf("Host: %+v\n", host)
		// Find each raw data CSV
		for _, key := range keys {
			fileList := findFile(host.resultDir, key)
			// findFile returns slice, though there should only be one file
			for _, file := range fileList {
				// Parse file into 2d-string slice
				result, err := readCSV(file)
				if err != nil {
					continue
				}
				// In a single file we have multiple headers to extract
				for _, header := range fileHeader[key] {
					// Extract single column of data that we want
					newResult, err := newSlice(result, header)
					if err != nil {
						//need to keep list of columns same for all types
						//continue
					}

					// Mutate host to add calcuated stats to object
					hosts[i].addResult(newResult, file, header)

				}
				fmt.Printf("CALLER Host populated: %+v\n", hosts[i])

			}
		}
	}

	csvFile, err := os.Create(resultDir + "out.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	// Write test CSV data to stdout
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()
	for i := range hosts {
		writer.Write(hosts[i].toSlice())
	}
}

func readCSV(file string) ([][]string, error) {
	fmt.Println(file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	result, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *host) toSlice() (row []string) {
	row = append(row, h.kind)
	for _, result := range h.results {
		row = append(row, strconv.FormatFloat(result.avg, 'f', 2, 64))
	}
	return
}

func (h *host) addResult(newResult []float64, file string, kind string) []resultType {
	min, _ := stats.Minimum(newResult)
	max, _ := stats.Maximum(newResult)
	avg, _ := stats.Mean(newResult)
	pct95, _ := stats.Percentile(newResult, 95)

	h.results = append(h.results, resultType{
		kind:  kind,
		path:  file,
		min:   min,
		max:   max,
		avg:   avg,
		pct95: pct95,
	})

	return h.results

}

func findFile(dir string, ext string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal(err)
	}
	var fileList []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		r, err := regexp.MatchString(ext, f.Name())
		if err == nil && r {
			fileList = append(fileList, path)
		}
		return nil
	})
	return fileList
}

func newSlice(bigSlice [][]string, title string) ([]float64, error) {
	floatValues := make([]float64, len(bigSlice)-1)
	var column int
	for i, v := range bigSlice {
		if i == 0 {
			var err error
			column, err = stringPositionInSlice(title, v)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			continue
		}
		value, _ := strconv.ParseFloat(bigSlice[i][column], 64)
		floatValues[i-1] = value
	}
	return floatValues, nil
}

// TODO: handle duplicates or none
func stringPositionInSlice(a string, list []string) (int, error) {
	for i, v := range list {
		match, _ := regexp.MatchString(a, v)
		if match {
			return i, nil
		}
	}
	return 0, fmt.Errorf("No matching headers")
}
