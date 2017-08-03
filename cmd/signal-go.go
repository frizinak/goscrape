// +build !js

package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
)

func trap(max *int, stderr io.Writer) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		fmt.Fprintln(stderr, "quitting...")
		*max = 1
	}()
}
