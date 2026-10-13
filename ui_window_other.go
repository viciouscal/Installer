//go:build !cli && !windows

package main

func initialWindowSize() (int, int) {
	return 1200, 700
}
