package routerdriver

import (
	//"bytes"
	"strings"
)

func SlashPath(p string) string {
	if p == "" {
		return "/"
	}

	pl := len(p)
	//var buffer bytes.Buffer

	switch {
	case p[0] != '/' && p[pl-1] != '/':
		//buffer.WriteString("/")
		//buffer.WriteString(p)
		//buffer.WriteString("/")
		return "/" + p + "/"
	case p[0] != '/':
		//buffer.WriteString("/")
		//buffer.WriteString(p)
		return "/" + p
	case p[pl-1] != '/':
		//buffer.WriteString(p)
		//buffer.WriteString("/")
		return p + "/"
	}

	return p
}

func CleanPath(p string) string {

	if p == "" {
		return "/"
	}
	n := len(p)
	var buf []byte

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1
	w := 1

	if p[n-1] != '/' {
		p += "/"
		n++
	}

	if p[0] != '/' {
		r = 0
		buf = make([]byte, n+1)
		buf[0] = '/'
	}

	trailing := n > 2 && p[n-1] == '/'

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp). So in contrast to the path package this
	// loop has no expensive function calls (except 1x make)

	for r < n {
		switch {
		case p[r] == '/':
			// empty path element, trailing slash is added after the end
			r++

		case p[r] == '.' && r+1 == n:
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/':
			// . element
			r++

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// .. element: remove to last /
			r += 2

			if w > 1 {
				// can backtrack
				w--

				if buf == nil {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}

		default:
			// real path element.
			// add slash if needed
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}

			// copy element
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	// re-append trailing slash
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}

	if buf == nil {
		return p[:w]
	}
	return string(buf[:w])
}

// internal helper to lazily create a buffer if necessary
func bufApp(buf *[]byte, s string, w int, c byte) {
	if *buf == nil {
		if s[w] == c {
			return
		}

		*buf = make([]byte, len(s))
		copy(*buf, s[:w])
	}
	(*buf)[w] = c
}

//parse thr url our defined.
func ParseDefinedUrl(path string) (string, []string) {
	morePath := strings.Split(path, "*")

	anPath := ""
	switch len(morePath) {
	case 1:
		anPath = path
	case 2:
		anPath = morePath[0]
	default:
		panic("not invlide path! split *")
	}

	manyPath := strings.Split(anPath, ":")

	realPath := ""
	var realMap = make([]string, 0)
	switch len(manyPath) {
	case 1:
		realPath = anPath
	case 0:
		panic("not invlide path! split:")
	default:
		realPath = manyPath[0]
		realMap = manyPath[1:]
	}

	if len(morePath) == 2 {
		realMap = append(realMap, morePath[1:]...)
	}

	//strip the last charact if the char is /
	for inx, value := range realMap {
		if l := len(value); value[l-1] == '/' && l > 2 {
			realMap[inx] = string(value[:l-1])
		}
	}

	return realPath, realMap
}
