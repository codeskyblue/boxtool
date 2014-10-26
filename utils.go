package main

import (
	"fmt"
	"strconv"
)

func genHelp(title string, items map[string]string) string {
	maxlen := 0
	keys := make([]string, 0, len(items))
	for key, _ := range items {
		keys = append(keys, key)
		if maxlen < len(key) {
			maxlen = len(key)
		}
	}
	help := title + "\n"
	for _, key := range keys {
		help += fmt.Sprintf("%-"+strconv.Itoa(maxlen+8)+"s%s\n", key, items[key])
	}
	return help
}
