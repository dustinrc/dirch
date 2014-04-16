package main

import (
	"fmt"
	"github.com/daviddengcn/go-villa"
	"github.com/dustin/go-humanize"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CountSize struct {
	count int
	size  int64
}

func (cs *CountSize) String() string {
	return fmt.Sprintf("%d total, %s", cs.count, humanize.Bytes(uint64(cs.size)))
}

type FileDirCount struct {
	numFiles, numDirs int
}

func (fdc *FileDirCount) Count(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	if info.IsDir() {
		fdc.numDirs++
	} else {
		fdc.numFiles++
	}
	return nil
}

func (fdc FileDirCount) String() string {
	var f, d string
	if fdc.numFiles == 1 {
		f = "file"
	} else {
		f = "files"
	}
	if fdc.numDirs == 1 {
		d = "directory"
	} else {
		d = "directories"
	}
	return fmt.Sprintf("%d %s, %d %s", fdc.numFiles, f, fdc.numDirs, d)
}

type ExtensionCountSize map[string]*CountSize

func (ecs ExtensionCountSize) Count(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	ext := filepath.Ext(path)
	if v, ok := ecs[ext]; ok {
		v.count++
		v.size += info.Size()
	} else {
		ecs[ext] = &CountSize{1, 1}
	}
	return nil
}

func (ecs ExtensionCountSize) SortBy(value string) []string {
	var keys []string
	var sortFn func([]string)
	for k := range ecs {
		keys = append(keys, k)
	}
	switch value {
	default:
		sortFn = sort.Strings
	case "count":
		sortFn = func(keys []string) {
			villa.SortF(
				len(keys),
				// need to order secondly by key
				func(i, j int) bool { return ecs[keys[i]].count > ecs[keys[j]].count },
				func(i, j int) { keys[i], keys[j] = keys[j], keys[i] },
			)
		}
	case "count<":
		sortFn = func(keys []string) {
			villa.SortF(
				len(keys),
				// need to order secondly by key
				func(i, j int) bool { return ecs[keys[i]].count < ecs[keys[j]].count },
				func(i, j int) { keys[i], keys[j] = keys[j], keys[i] },
			)
		}
	case "size":
		sortFn = func(keys []string) {
			villa.SortF(
				len(keys),
				// need to order secondly by key
				func(i, j int) bool { return ecs[keys[i]].size > ecs[keys[j]].size },
				func(i, j int) { keys[i], keys[j] = keys[j], keys[i] },
			)
		}
	case "size<":
		sortFn = func(keys []string) {
			villa.SortF(
				len(keys),
				// need to order secondly by key
				func(i, j int) bool { return ecs[keys[i]].size < ecs[keys[j]].size },
				func(i, j int) { keys[i], keys[j] = keys[j], keys[i] },
			)
		}
	}
	sortFn(keys)
	return keys
}

func (ecs ExtensionCountSize) String() string {
	keys := ecs.SortBy("key")
	var out []string
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s: %s", k, ecs[k]))
	}
	return strings.Join(out, "\n")
}

type MultiFuncDispatch struct {
	fns []filepath.WalkFunc
}

func (mfd *MultiFuncDispatch) Dispatch(path string, info os.FileInfo, e error) error {
	for _, fn := range mfd.fns {
		//		if err := fn(path, info, e); err != nil {
		//			fmt.Fprintln(os.Stderr, err)
		//			return err
		//		}
		fn(path, info, e)
	}
	return nil
}

func main() {
	fdc := FileDirCount{}
	ecs := ExtensionCountSize{}

	sought := []filepath.WalkFunc{}
	sought = append(sought, fdc.Count, ecs.Count)
	mfd := MultiFuncDispatch{sought}

	filepath.Walk(os.Args[1], mfd.Dispatch)
	fmt.Println(fdc)
	fmt.Println(ecs)
}
