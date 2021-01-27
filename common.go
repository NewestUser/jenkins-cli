package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func TouchFile(name string) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			return err
		}
		defer file.Close()
		return nil
	}
	return nil
}

type PrettyFormat struct {
	template strings.Builder
}

func (w *PrettyFormat) PadField(vals interface{}, extractor func(int) interface{}) *PrettyFormat {
	s := InterfaceSlice(vals)

	fields := make([]interface{}, len(s), len(s))
	for i, _ := range s {
		fields[i] = extractor(i)
	}
	return w.Pad(fields)
}

func (w *PrettyFormat) Pad(vals interface{}) *PrettyFormat {
	var pad = 0
	for _, val := range InterfaceSlice(vals) {
		strVal := fmt.Sprintf("%v", val)
		if pad < len(strVal) {
			pad = len(strVal)
		}
	}

	w.template.WriteString(fmt.Sprintf("%%-%dv", pad))
	return w
}

func (w *PrettyFormat) Append(format string) *PrettyFormat {
	w.template.WriteString(format)
	return w
}

func (w *PrettyFormat) Format() string {
	return w.template.String()
}

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

type arrayFlag []string

func (f *arrayFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *arrayFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
