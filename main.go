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
	iostat
}

type iostat struct {
	path string
	iops resultType
}

type resultType struct {
	min, max, avg, pct95 float64
}

const iostatFilename string = "disk_IOPS.csv"

func main() {
	searchDir := "/home/sejug/pbench-user-benchmark_foo_2017-04-11_16:51:07/1/reference-result/tools-default/"

	dirList, err := ioutil.ReadDir(searchDir)
	if err != nil {
		log.Fatal(err)
	}

	var masterDir, nodeDir, etcdDir, lbDir []string
	for _, v := range dirList {
		if v.IsDir() {
			r, err := regexp.MatchString("svt_master_", v.Name())
			if r {
				masterDir = append(masterDir, v.Name())
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

	for _, path := range masterDir {
		fmt.Println("Masters: ", path)
		fileList := findFile(searchDir+path, iostatFilename)
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

			fmt.Printf("Lines %v\n", len(result))
			newResult := newSlice(result, 2)
			total := sum(newResult)
			fmt.Printf("The sum of the values in column 2 is: %v\n", total)
			avg, _ := mean(newResult)
			fmt.Printf("The average of column 2 is: %v\n", avg)
			min, _ := minimum(newResult)
			fmt.Printf("The minimum of column 2 is: %v\n", min)
			max, _ := maximum(newResult)
			fmt.Printf("The maximum of column 2 is: %v\n", max)
			perc, _ := percentile(newResult, 95)
			fmt.Printf("The 95th percentile of column 2 is: %v\n", perc)
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
