package serverconf

import (
	"bufio"
	"bytes"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type jsoncfg struct {
	top map[string]string
	sub map[int64]map[string]string
}

func newjsoncfg(path string) (*jsoncfg, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	jpp := newjsonpp(f)
	defer jpp.Close()

	dec := json.NewDecoder(jpp)
	dec.UseNumber()
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('{') {
		return nil, errors.New("json config: top level must be object")
	}
	top := make(map[string]string)
	sub := make(map[int64]map[string]string)
	err = parseobj(dec, top, sub)
	if err != nil {
		return nil, err
	}
	return &jsoncfg{top: top, sub: sub}, nil
}

func parseobj(dec *json.Decoder, top map[string]string, sub map[int64]map[string]string) error {
	for {
		k, err := dec.Token()
		if err != nil {
			return err
		}
		t, err := dec.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case string:
			top[k.(string)] = v
		case json.Number:
			n, err := v.Int64()
			if err != nil {
				return err
			}
			// The back-and-forth conversion enforces integers only, at
			// full precision, but strconv does not understand E-notation.
			top[k.(string)] = strconv.FormatInt(n, 10)
		case bool:
			if v {
				top[k.(string)] = "true"
			} else {
				top[k.(string)] = "false"
			}
		case nil:
			top[k.(string)] = ""
		case json.Delim:
			if sub == nil {
				return errors.New(fmt.Sprintf("json config: nested more than once, at %v", k))
			}
			if v != json.Delim('{') {
				return errors.New(fmt.Sprintf("json config: can only nest objects, at %v", k))
			}
			i, err := strconv.ParseInt(k.(string), 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("json config: nested object key must be int, at %v", k))
			}
			sub[i] = make(map[string]string)
			err = parseobj(dec, sub[i], nil)
			if err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("json config: unknown token type, at %v", k))
		}
		if !dec.More() {
			break
		}
	}
	return nil
}

func (f *jsoncfg) GlobalMap() map[string]string {
	m := make(map[string]string, len(f.top))
	for k, v := range f.top {
		m[k] = v
	}
	return m
}

func (f *jsoncfg) SubMap(sub int64) map[string]string {
	if _, ok := f.sub[sub]; !ok {
		return nil
	}
	m := make(map[string]string, len(f.sub[sub]))
	for k, v := range f.sub[sub] {
		m[k] = v
	}
	return m
}

type jsonpp struct {
	f    io.ReadCloser
	buf  bytes.Buffer
	scan *bufio.Scanner
	// Note: bufio.Scanner will not allocate more than 65536 bytes per line.
}

func newjsonpp(f io.ReadCloser) *jsonpp {
	return &jsonpp{f: f, scan: bufio.NewScanner(f)}
}

func (j *jsonpp) Read(p []byte) (n int, err error) {
	for j.buf.Len() < len(p) {
		// This JSON-with-comments preprocessor is simple, but will break
		// some valid JSON by splitting on the // sequence inside a string.
		// Fortunately, the escape /\/ is valid and parsed without problems.
		if j.scan.Scan() {
			if strings.Contains(j.scan.Text(), "//") {
				j.buf.WriteString(strings.SplitN(j.scan.Text(), "//", 2)[0])
			} else {
				j.buf.Write(j.scan.Bytes())
			}
		} else if j.scan.Err() != nil {
			return 0, j.scan.Err()
		} else {
			break
		}
	}
	return j.buf.Read(p)
}

func (j *jsonpp) Close() error {
	return j.f.Close()
}
