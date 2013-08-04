package config

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const (
	eof rune = -(iota + 1)
	invalid
)

// An Event is a type of event that can be detected by the parser.
type Event int

const (
	Section Event = iota // A new section in the config file.
	Key                  // A new key in the config file.
	EOF                  // The end of the config file.
)

// ParseError is returned from Next when the file cannot be parsed
// correctly.
type ParseError struct {
	Line int    // The line where the error occurred.
	Col  int    // The column in the line where the error occured.
	Msg  string // A description of the error.
}

func (p ParseError) Error() string {
	return fmt.Sprintf("[%d, %d] %s", p.Line, p.Col, p.Msg)
}

// Parser reads the configuration file finding sections and keys
type Parser struct {
	Section   string // Current section name, "" if there is no section
	Parameter string // Current section parameter\, "" if there is no section, or the current seciton has no parameter
	Key       string // Current name, "" if this is a Section event
	Value     string // Current value, "" if there is no value or we have a Section event

	b    []byte
	r    rune  // current rune
	pos  int   // position of current rune
	npos int   // position of next rune
	line int   // line of current rune
	col  int   // column of current rune
	err  error // current active error
}

// NewParser creates a parser to parse the given input bytes
func NewParser(b []byte) *Parser {
	return &Parser{b: b}
}

// Next finds the next Section or Key in the file.
func (p *Parser) Next() (e Event, err error) {
	for p.err == nil {
		p.read()

		switch {
		case p.r == eof:
			e = EOF
			return
		case p.r == '\n':
			continue
		case p.r == ';' || p.r == '#':
			p.scanComment()
			continue
		case p.r == '[':
			p.read()
			e = Section
			p.Section, p.Parameter = p.parseSection()
			err = p.err
			return
		case isSpace(p.r):
			continue
		case isName(p.r):
			e = Key
			p.Key, p.Value = p.parseKey()
			err = p.err
			return
		default:
			p.unexpected("'[', NAME, ';' or '#'")
		}
	}

	err = p.err
	return
}

func (p *Parser) read() {
	var l int

	if p.r == eof || p.r == invalid {
		return
	}

	if p.r == '\n' || p.npos == 0 {
		p.line++
		p.col = 1
	} else {
		p.col++
	}

	if p.npos >= len(p.b) {
		p.r = eof
		l = 0
	} else {
		p.r, l = utf8.DecodeRune(p.b[p.npos:])
		if p.r == utf8.RuneError && l < 2 {
			p.error("UTF-8 encoding error")
			p.r = invalid
		}
	}

	p.pos, p.npos = p.npos, p.npos+l

}

func (p *Parser) unexpected(msg string) {
	if p.r == eof {
		p.error("Unexpected EOF, expecting " + msg + ".")
	} else {
		p.error(fmt.Sprintf("Unexpected %q, expecting %s.", p.r, msg))
	}
}

func (p *Parser) error(msg string) {
	if p.err == nil {
		p.err = ParseError{p.line, p.col, msg}
	}
}

func (p *Parser) parseKey() (key, value string) {
	key = p.parseName()
	p.scanSpace()
	if p.r == ';' || p.r == '#' || p.r == '\n' || p.r == eof {
		goto afterValue
	}

	if p.r != '=' {
		p.unexpected("'=', ';' or '#'")
	}

	p.read()
	p.scanSpace()
	switch p.r {
	case ';', '#', '\n':
		// continue
	case '"':
		p.read()
		value = p.parseString()
	case '`':
		p.read()
		value = p.parseRawString()
	default:
		value = p.parseValue()
	}

afterValue:
	p.scanSpace()
	if p.r == ';' || p.r == '#' {
		p.scanComment()
	}

	if p.r != '\n' && p.r != eof {
		p.unexpected("'\\n' or EOF")
	}

	return
}

func (p *Parser) parseSection() (name, parameter string) {
	p.scanSpace()
	name = p.parseName()
	if p.r == ']' {
		goto closed
	} else if !isSpace(p.r) {
		p.unexpected("']' or space")
	}

	p.scanSpace()
	if p.r == '"' {
		p.read()
		parameter = p.parseString()
		p.scanSpace()
	}

	if p.r != ']' {
		p.unexpected("']'")
	}

closed:
	p.read()
	p.scanSpace()
	if p.r == ';' || p.r == '#' {
		p.scanComment()
	}

	if p.r != '\n' && p.r != eof {
		p.unexpected("'\\n' or EOF")
	}

	return
}

func (p *Parser) parseValue() string {
	var buf []rune
	end := 0

	for {
		switch {
		case isSpace(p.r):
			buf = append(buf, p.r)
		case p.r == ';' || p.r == '#' || p.r == '\n' || p.r == invalid || p.r == eof:
			return string(buf[:end])
		default:
			buf = append(buf, p.r)
			end = len(buf)
		}

		p.read()
	}

	panic("unreachable!")
}

func (p *Parser) parseString() string {
	var buf []rune
	for {
		switch p.r {
		case '"':
			p.read()
			return string(buf)
		case '\n', eof:
			p.error("unterminated string")
			return ""
		case invalid:
			return ""
		case '\\':
			p.read()
			buf = p.parseStringEscape(buf)
		default:
			buf = append(buf, p.r)
			p.read()
		}
	}

	panic("unreachable!")
}

func (p *Parser) parseStringEscape(buf []rune) []rune {
	switch p.r {
	case '\\', '"':
		buf = append(buf, p.r)
	case 'n':
		buf = append(buf, '\n')
	case 'r':
		buf = append(buf, '\r')
	case 't':
		buf = append(buf, '\t')
	default:
		p.unexpected("'\\', '\"', 'n', 'r', 't'")
		return buf
	}

	p.read()
	return buf
}

func (p *Parser) parseRawString() string {
	start := p.pos
	end := start

	for {
		switch p.r {
		case '`':
			end = p.pos
			p.read()
			return string(p.b[start:end])
		case eof:
			p.error("Unterminated raw string")
			return ""
		case invalid:
			return ""
		default:
			p.read()
		}
	}

	panic("unreachable!")
}

func (p *Parser) parseName() string {
	if !isName(p.r) {
		p.unexpected("NAME")
	}

	start := p.pos
	p.read()
	for isName(p.r) {
		p.read()
	}

	return string(p.b[start:p.pos])
}

func (p *Parser) scanComment() {
	for p.r != '\n' && p.r != eof {
		p.read()
	}
}

func (p *Parser) scanName() {
	for isName(p.r) {
		p.read()
	}
}

func (p *Parser) scanSpace() {
	for isSpace(p.r) {
		p.read()
	}
}

func isSpace(r rune) bool {
	return r != '\n' && unicode.IsSpace(r)
}

func isName(r rune) bool {
	return r == '.' || r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
