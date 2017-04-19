package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type host struct {
	kind      string
	resultDir string
	results   []resultType
}

type resultType struct {
	kind, path           string
	min, max, avg, pct95 float64
}

var fileHeader = map[string]string{
	"disk_IOPS.csv":             "vda-write",
	"cpu_usage_percent_cpu.csv": "openshift_start_node",
}

func main() {
	var hosts []host
	var hostRegex []*regexp.Regexp
	searchDir := "/home/sejug/pbench-user-benchmark_foo_2017-04-11_16:51:07/1/reference-result/tools-default/"
	hostTags := []string{"svt_master_", "svt_node_", "svt_etcd_", "svt_lb_"}

	for _, tag := range hostTags {
		newRegex := regexp.MustCompile(tag)
		hostRegex = append(hostRegex, newRegex)
	}

	dirList, err := ioutil.ReadDir(searchDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range dirList {
		if item.IsDir() {
			for _, regex := range hostRegex {
				if regex.MatchString(item.Name()) {
					kind := strings.Split(item.Name(), "_")
					newHost := host{
						kind:      kind[1],
						resultDir: searchDir + item.Name(),
					}
					hosts = append(hosts, newHost)
				}
			}
		}
	}

	for i, host := range hosts {
		fmt.Printf("Host: %+v\n", host)
		for fileName, headerName := range fileHeader {
			fileList := findFile(host.resultDir, fileName)
			for _, file := range fileList {
				result, err := readCSV(file)
				if err != nil {
					continue
				}
				newResult, err := newSlice(result, headerName)
				if err != nil {
					continue
				}

				hosts[i].addResult(newResult, file, headerName)
				fmt.Printf("CALLER Host populated: %+v\n", hosts[i])

			}
		}
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

func (h *host) addResult(newResult []float64, file string, kind string) []resultType {
	min, _ := minimum(newResult)
	max, _ := maximum(newResult)
	avg, _ := mean(newResult)
	pct95, _ := percentile(newResult, 95)

	h.results = append(h.results, resultType{
		kind:  kind,
		path:  file,
		min:   min,
		max:   max,
		avg:   avg,
		pct95: pct95,
	})

	fmt.Printf("FUNCTION Host populated: %+v\n", h)
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
