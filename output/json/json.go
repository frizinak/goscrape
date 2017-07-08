package json

import (
	"encoding/json"
	"fmt"
	"io"
)

var nl = []byte{'\n'}

type JSON struct {
	w    *json.Encoder
	keys []string
}

func New(w io.Writer, keys []string) *JSON {
	return &JSON{json.NewEncoder(w), keys}
}

func (c *JSON) Write(f []fmt.Stringer) error {
	strs := make(map[string]string, len(f))
	for i := range f {
		strs[c.keys[i]] = f[i].String()
	}

	return c.w.Encode(strs)
}
