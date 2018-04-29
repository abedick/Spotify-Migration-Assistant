package common

import (
	"fmt"
	"reflect"
)

func SectionTitle(title string) {

	width := 55

	fmt.Print("\u256D")
	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u256E\n\u2502 \u25B6 ", title, " ")

	for i := 0; i < width-(len(title)+4); i++ {
		fmt.Print(" ")
	}
	fmt.Print("\u2502\n\u2570")

	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u256F\n")
}

func SubsectionTitle(title string) {

	width := 55

	fmt.Print("   \u256D")
	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u256E\n   \u2502 \u2301 ", title, " ")

	for i := 0; i < width-(len(title)+4); i++ {
		fmt.Print(" ")
	}
	fmt.Print("\u2502\n   \u2570")

	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u256F\n")
}

func Contains(slice []interface{}, item interface{}) (bool, error) {

	if len(slice) == 0 {
		return false, nil
	}

	slice_t := reflect.TypeOf(slice[0])
	item_t := reflect.TypeOf(item)

	if slice_t != item_t {
		return false, &errorString{"Slice type != item type"}
	}

	for _, index := range slice {
		if index == item {
			return true, nil
		}
	}

	return false, nil
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
