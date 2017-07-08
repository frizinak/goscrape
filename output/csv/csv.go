package csv

import (
	"encoding/csv"
	"fmt"
	"io"
)

type CSV struct {
	w      *csv.Writer
	header []string
	first  bool
}

func New(w io.Writer, header []string) *CSV {
	return &CSV{csv.NewWriter(w), header, true}
}

func (c *CSV) Write(f []fmt.Stringer) error {
	if c.first && c.header != nil {
		c.first = false
		if err := c.w.Write(c.header); err != nil {
			return err
		}
	}

	strs := make([]string, len(f))
	for i := range f {
		strs[i] = f[i].String()
	}

	if err := c.w.Write(strs); err != nil {
		return err
	}

	c.w.Flush()

	return nil
}
