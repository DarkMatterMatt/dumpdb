/**
 * Source: https://gist.github.com/regeda/969a067ff4ed6ffa8ed6
 */

package pkg

import (
	"unicode"
	"unicode/utf8"
)

type buffer struct {
	r         []byte
	runeBytes [utf8.UTFMax]byte
}

func (b *buffer) write(r rune) {
	if r < utf8.RuneSelf {
		b.r = append(b.r, byte(r))
		return
	}
	n := utf8.EncodeRune(b.runeBytes[0:], r)
	b.r = append(b.r, b.runeBytes[0:n]...)
}

func (b *buffer) indent() {
	if len(b.r) > 0 {
		b.r = append(b.r, '_')
	}
}

// CamelcaseToUnderscore converts a camelcase style string to an underscore style string.
func CamelcaseToUnderscore(s string) string {
	b := buffer{
		r: make([]byte, 0, len(s)),
	}
	var m rune
	var w bool
	for _, ch := range s {
		if unicode.IsUpper(ch) {
			if m != 0 {
				if !w {
					b.indent()
					w = true
				}
				b.write(m)
			}
			m = unicode.ToLower(ch)
		} else {
			if m != 0 {
				b.indent()
				b.write(m)
				m = 0
				w = false
			}
			b.write(ch)
		}
	}
	if m != 0 {
		if !w {
			b.indent()
		}
		b.write(m)
	}
	return string(b.r)
}

type camelcaseToUnderscoreReplacer struct{}

// Replace returns a copy of s with all replacements performed.
func (*camelcaseToUnderscoreReplacer) Replace(s string) string {
	return CamelcaseToUnderscore(s)
}

// NewCamelcaseToUnderscoreReplacer creates a string replacer.
func NewCamelcaseToUnderscoreReplacer() *camelcaseToUnderscoreReplacer {
	return &camelcaseToUnderscoreReplacer{}
}
