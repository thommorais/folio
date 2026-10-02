package main

import "fmt"

func written(record any, id string, lines ...string) error {
	if flagJSON {
		return encode(record)
	}
	fmt.Println(id)
	for _, line := range lines {
		if line != "" {
			fmt.Println(line)
		}
	}
	return nil
}

func slugLine(slug string) string {
	if slug == "" {
		return ""
	}
	return "slug: " + slug
}

func blockedLine(blocked bool) string {
	if blocked {
		return "blocked"
	}
	return ""
}
