package network

import (
	"fmt"
	"io"
)

func Placeholder(w io.Writer) error {
	_, err := fmt.Fprintln(w, "Network Placeholder")
	return err
}
