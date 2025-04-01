package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"sync"
)

type EpicManifestJsonFile struct {
	Name       string `json:"Filename"`
	HashString string `json:"FileHash"`
}

func (emjf *EpicManifestJsonFile) Hash() []byte {
	var ret []byte
	for i := 0; i < 20; i++ {
		num, err := strconv.Atoi(emjf.HashString[i*3 : (i*3)+3])
		check(err)
		ret = append(ret, byte(num))
	}
	return ret
}

type EpicManifestJson struct {
	BuildVersionString string                 `json:"BuildVersionString"`
	Files              []EpicManifestJsonFile `json:"FileManifestList"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkfile(basepath string, file EpicManifestJsonFile, wg *sync.WaitGroup, errors *[]string) {
	defer wg.Done()

	filepath := path.Join(basepath, file.Name)

	var checkfile = true
	f, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("\033[31m%s does not exist!\033[37m\n", file.Name)
			*errors = append(*errors, file.Name)
			checkfile = false
		} else {
			panic(err)
		}
	}
	if checkfile {
		defer f.Close()

		h := sha1.New()
		buf := make([]byte, 32768)

		for {
			n, err := f.Read(buf)
			if err != nil && err.Error() != "EOF" {
				panic(err)
			}
			if n == 0 {
				break
			}
			h.Write(buf[:n])
		}

		if reflect.DeepEqual(h.Sum(nil), file.Hash()) {
			fmt.Printf("\033[32m%s is fine!\033[37m\n", file.Name)
		} else {
			fmt.Printf("\033[31m%s has wrong hash!\033[37m\n", file.Name)
			*errors = append(*errors, file.Name)
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s [path to manifest file] [path to game folder]\n", os.Args[0])
		return
	}

	basepath := os.Args[2]

	manifestdata, err := os.ReadFile(os.Args[1])
	check(err)

	manifest := EpicManifestJson{}
	json.Unmarshal(manifestdata, &manifest)

	wg := &sync.WaitGroup{}
	var errors []string
	for i := 0; i < len(manifest.Files); i++ {
		wg.Add(1)

		file := manifest.Files[i]
		go checkfile(basepath, file, wg, &errors)
	}

	wg.Wait()

	fmt.Println()
	if len(errors) > 0 {
		fmt.Printf("Found %d error(s) with the following files:\n", len(errors))
		for i := 0; i < len(errors); i++ {
			fmt.Printf(" - %s\n", errors[i])
		}
	} else {
		fmt.Println("No errors were found!")
	}
}
