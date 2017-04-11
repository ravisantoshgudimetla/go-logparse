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
)

type host struct {
	kind      string
	resultDir string
	iostat    iostatType
}

type iostatType struct {
	path string
	iops resultType
}

type resultType struct {
	min, max, avg, pct95 float64
}

const iostatFilename string = "disk_IOPS.csv"
const iostatHeader string = "vda-write"

func main() {
	searchDir := "/home/sejug/pbench-user-benchmark_foo_2017-04-11_16:51:07/1/reference-result/tools-default/"
	var masters []host

	dirList, err := ioutil.ReadDir(searchDir)
	if err != nil {
		log.Fatal(err)
	}

	var nodeDir, etcdDir, lbDir []string
	for _, v := range dirList {
		if v.IsDir() {
			r, err := regexp.MatchString("svt_master_", v.Name())
			if r {
				masterNew := host{
					kind:      "master",
					resultDir: searchDir + v.Name(),
					iostat:    iostatType{},
				}
				masters = append(masters, masterNew)
			}
			r, err = regexp.MatchString("svt_node_", v.Name())
			if r {
				nodeDir = append(nodeDir, v.Name())
			}
			r, err = regexp.MatchString("svt_etcd_", v.Name())
			if r {
				etcdDir = append(etcdDir, v.Name())
			}
			r, err = regexp.MatchString("svt_lb__", v.Name())
			if r {
				lbDir = append(lbDir, v.Name())
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for i, master := range masters {
		fmt.Printf("Masters: %+v\n", master)
		fileList := findFile(master.resultDir, iostatFilename)
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

			newResult := newSlice(result, iostatHeader)
			masters[i].iostat.path = file
			masters[i].iostat.iops.min, _ = minimum(newResult)
			masters[i].iostat.iops.max, _ = maximum(newResult)
			masters[i].iostat.iops.avg, _ = mean(newResult)
			masters[i].iostat.iops.pct95, _ = percentile(newResult, 95)

			fmt.Printf("Master: %+v\n", masters[i])

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
