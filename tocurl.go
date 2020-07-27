package tocurl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

func FromRequest(r *http.Request) string {
	return fromRequest(r, false)
}

func FromRequestComplete(r *http.Request) string {
	return fromRequest(r, true)
}

func fromRequest(r *http.Request, more bool) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("curl -v")
	if r.Method != "" {
		buf.WriteString(fmt.Sprintf(" -X %s", showData(r.Method, more)))
	}
	setHeaders(buf, r.Header, more)
	if r.Host != "" && r.Host != r.URL.Host {
		buf.WriteString(fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %v", "Host", showData(r.Host, more))))
	}

	if ua := r.UserAgent(); ua != "" {
		buf.WriteString(fmt.Sprintf(" -A %q", showData(ua, more)))
	}

	if ref := r.Referer(); ref != "" {
		buf.WriteString(fmt.Sprintf(" -e %q", showData(ref, more)))
	}

	setRequestBody(buf, r, more)

	buf.WriteString(fmt.Sprintf(" %q", showData(r.URL.String(), more)))
	return buf.String()
}

func setRequestBody(buf *bytes.Buffer, r *http.Request, more bool) {
	if r == nil || r.Body == nil {
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewReader(data))

	if len(data) == 0 {
		return
	}

	buf.WriteString(fmt.Sprintf(" -d %q", showData(string(data), more)))
}

func setHeaders(buf *bytes.Buffer, h http.Header, more bool) {
	if len(h) == 0 {
		return
	}
	headerKey := make([]string, 0, len(h))

	for key := range h {
		headerKey = append(headerKey, key)
	}

	sort.Strings(headerKey)

	for _, key := range headerKey {
		values := h[key]
		for _, value := range values {
			if strings.ToLower(key) != "host" {
				buf.WriteString(fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", showData(key, more), showData(value, more))))
			}
		}
	}
}

func showData(data string, more bool) string {
	if more {
		return data
	}
	const max = 64
	if len(data) > max {
		var tmpArray [max * 2]byte
		tmp := tmpArray[:0]
		tmp = append(tmp, data[:max/2]...)
		tmp = append(tmp, []byte(fmt.Sprintf(" ... %d bytes hide ... ", len(data)-max))...)
		tmp = append(tmp, data[len(data)-1-max/2:]...)
		return string(tmp)
	}
	return data
}
