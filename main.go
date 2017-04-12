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
	iostat    iostatType
	pidstat   pidstatType
}

type iostatType struct {
	path string
	iops resultType
}

type pidstatType struct {
	path       string
	cpuPercent resultType
}

type resultType struct {
	min, max, avg, pct95 float64
}

const iostatFilename string = "disk_IOPS.csv"
const iostatHeader string = "vda-write"
const pidstatFilename string = "cpu_usage_percent_cpu.csv"
const pidstatHeader string = "openshift_start_node"

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
						iostat:    iostatType{},
					}
					hosts = append(hosts, newHost)
				}
			}
		}
	}

	for i, host := range hosts {
		fmt.Printf("Host: %+v\n", host)
		fileList := findFile(host.resultDir, iostatFilename)
		for _, file := range fileList {
			fmt.Println(file)
			f, _ := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			r := csv.NewReader(bufio.NewReader(f))
			result, err := r.ReadAll()
			if err != nil {
				log.Fatal(err)
			}

			newResult, err := newSlice(result, iostatHeader)
			if err != nil {
				continue
			}
			hosts[i].iostat.path = file
			hosts[i].iostat.iops.min, _ = minimum(newResult)
			hosts[i].iostat.iops.max, _ = maximum(newResult)
			hosts[i].iostat.iops.avg, _ = mean(newResult)
			hosts[i].iostat.iops.pct95, _ = percentile(newResult, 95)

			fmt.Printf("Host populated: %+v\n", hosts[i])

		}
	}
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
