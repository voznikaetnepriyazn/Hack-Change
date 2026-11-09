package repeat

import "bytes"

// RepetitionDetector detects repetition in a stream of text.
type RepetitionDetector struct {
	buf       []byte
	bufSize   int // 256
	triggered bool
}

// Constructor
func NewRepetitionDetector(bufSize int) *RepetitionDetector {
	if bufSize <= 0 {
		bufSize = 256
	}
	return &RepetitionDetector{
		buf:     make([]byte, 0, bufSize),
		bufSize: bufSize,
	}
}

// Feed a new chunk of text to the detector.
func (d *RepetitionDetector) Feed(delta string) bool {
	d.buf = append(d.buf, delta...)
	if len(d.buf) < d.bufSize {
		return false
	}

	window := d.buf[len(d.buf)-d.bufSize:]
	for patLen := 1; patLen <= 20; patLen++ {
		pattern := window[:patLen]
		covers := true
		for i := patLen; i < len(window); i += patLen {
			end := i + patLen
			if end > len(window) {
				end = len(window)
			}
			if !bytes.Equal(window[i:end], pattern[:end-i]) {
				covers = false
				break
			}
		}
		if covers {
			d.triggered = true
			return true
		}
	}
	return false
}
