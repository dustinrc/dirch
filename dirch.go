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

const (
	otherFile = iota
	regularFile
	directory
)

func fileType(info os.FileInfo) int {
	if info.IsDir() {
		return directory
	} else if info.Mode().IsRegular() {
		return regularFile
	} else {
		return otherFile
	}
}

type CountSize struct {
	count int
	size  int64
}

func (cs *CountSize) String() string {
	return fmt.Sprintf("%d total, %s", cs.count, humanize.Bytes(uint64(cs.size)))
}

type FileDirCount struct {
	regularFiles, otherFiles, directories int
}

func (fdc *FileDirCount) Count(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	switch fileType(info) {
	case regularFile:
		fdc.regularFiles++
	case directory:
		fdc.directories++
	case otherFile:
		fdc.otherFiles++
	}

	return nil
}

func (fdc FileDirCount) String() string {
	var f, d, o string
	if fdc.regularFiles == 1 {
		f = "file"
	} else {
		f = "files"
	}
	if fdc.directories == 1 {
		d = "directory"
	} else {
		d = "directories"
	}
	if fdc.otherFiles == 1 {
		o = "other"
	} else {
		o = "others"
	}
	return fmt.Sprintf("%d %s, %d %s, %d %s", fdc.regularFiles, f,
		fdc.directories, d,
		fdc.otherFiles, o)
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

type extInDir struct {
	ext string
	dir string
}

type ExtensionLocation map[extInDir]*CountSize

func (el ExtensionLocation) Count(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	eid := extInDir{filepath.Ext(path), filepath.Dir(path)}
	if v, ok := el[eid]; ok {
		v.count++
		v.size += info.Size()
	} else {
		el[eid] = &CountSize{1, info.Size()}
	}
	return nil
}

func (el ExtensionLocation) String() string {
	var out []string
	for k := range el {
		out = append(out, fmt.Sprintf("%s: %s", k, el[k]))
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
	el := ExtensionLocation{}

	sought := []filepath.WalkFunc{}
	sought = append(sought, fdc.Count, ecs.Count, el.Count)
	mfd := MultiFuncDispatch{sought}

	filepath.Walk(os.Args[1], mfd.Dispatch)
	fmt.Println(fdc)
	fmt.Println(ecs)
	fmt.Println(el)
}
