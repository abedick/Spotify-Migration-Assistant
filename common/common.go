package common

import "fmt"

func SectionTitle(title string) {

	width := 55

	fmt.Print("\u250C")
	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u2510\n\u2502 \u25B6 ", title, " ")

	for i := 0; i < width-(len(title)+4); i++ {
		fmt.Print(" ")
	}
	fmt.Print("\u2502\n\u2514")

	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u2518\n")
}

func SubsectionTitle(title string) {
	width := 55

	fmt.Print("\u250C")
	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u2510\n\u2502 \u25B6 ", title, " ")

	for i := 0; i < width-(len(title)+4); i++ {
		fmt.Print(" ")
	}
	fmt.Print("\u2502\n\u2514")

	for i := 0; i < width; i++ {
		fmt.Print("\u2500")
	}
	fmt.Print("\u2518\n")
}
