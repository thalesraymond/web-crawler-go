package indexer

import (
	"fmt"
	"io"
)

func Placeholder(w io.Writer) error {
	_, err := fmt.Fprintln(w, "Indexer Placeholder")
	return err
}
