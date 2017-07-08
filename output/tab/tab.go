package tab

import (
	"fmt"
	"io"
)

var (
	tab = []byte{'\t'}
	nl  = []byte{'\n'}
)

type Tab struct {
	w io.Writer
}

func New(w io.Writer) *Tab {
	return &Tab{w}
}

func (c *Tab) Write(f []fmt.Stringer) error {
	for i := range f {
		b := []byte(f[i].String())
		if _, err := c.w.Write(b); err != nil {
			return err
		}
		if _, err := c.w.Write(tab); err != nil {
			return err
		}
	}

	_, err := c.w.Write(nl)
	return err

}
