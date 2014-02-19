// Copyright (c) 2013, 2014 Martin Hilton <martin.hilton@zepler.net>
// 
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package config

import (
	"testing"
)

type testParser struct {
	p *Parser
	t *testing.T
}

func newTestParser(s string, t *testing.T) testParser {
	return testParser{NewParser([]byte(s)), t}
}

func (t testParser) eof() {
	e, err := t.p.Next()
	if err != nil {
		t.t.Fatalf("Error when expecting EOF: %s.", err)
	}
	if e != EOF {
		t.t.Fatalf("Incorrect event, expected EOF got %v.", e)
	}
}

func (t testParser) section(section, parameter string) {
	e, err := t.p.Next()
	if err != nil {
		t.t.Fatalf("Error when expecting Section: %s.", err)
	}
	if e != Section {
		t.t.Fatalf("Incorrect event, expecting Section got %v.", e)
	}
	if t.p.Section != section {
		t.t.Fatalf("Incorrect section value, expecting \"%s\" got \"%s\".", section, t.p.Section)
	}
	if t.p.Parameter != parameter {
		t.t.Fatalf("Incorrect parameter value, expecting \"%s\" got \"%s\".", parameter, t.p.Parameter)
	}
}

func (t testParser) key(section, parameter, key, value string) {
	e, err := t.p.Next()
	if err != nil {
		t.t.Fatalf("Error when expecting Key: %s.", err)
	}
	if e != Key {
		t.t.Fatalf("Incorrect event, expecting Key got %v.", e)
	}
	if t.p.Section != section {
		t.t.Fatalf("Incorrect section value, expecting \"%s\" got \"%s\".", section, t.p.Section)
	}
	if t.p.Parameter != parameter {
		t.t.Fatalf("Incorrect parameter value, expecting \"%s\" got \"%s\".", parameter, t.p.Parameter)
	}
	if t.p.Key != key {
		t.t.Fatalf("Incorrect key value, expecting \"%s\" got \"%s\".", key, t.p.Key)
	}
	if t.p.Value != value {
		t.t.Fatalf("Incorrect parameter value, expecting \"%s\" got \"%s\".", value, t.p.Value)
	}
}

func (t testParser) error() ParseError {
	_, err := t.p.Next()
	if err == nil {
		t.t.Fatalf("No error when expected")
	}

	t.t.Log(err)
	return err.(ParseError)
}

func TestEmptyFile(t *testing.T) {
	p := newTestParser("", t)
	p.eof()
}

func TestHashComment(t *testing.T) {
	p := newTestParser("# This is a Comment", t)
	p.eof()
}

func TestSemicolonComment(t *testing.T) {
	p := newTestParser("; This is also a Comment", t)
	p.eof()
}

func TestBlankLine(t *testing.T) {
	p := newTestParser("\n", t)
	p.eof()
}

func TestWhitespace(t *testing.T) {
	p := newTestParser("    ", t)
	p.eof()
}

func TestSection(t *testing.T) {
	p := newTestParser("[section]", t)
	p.section("section", "")
	p.eof()
}

func TestSectionWithSpaces(t *testing.T) {
	p := newTestParser("[	section   ]", t)
	p.section("section", "")
	p.eof()
}

func TestSectionWithParameter(t *testing.T) {
	p := newTestParser("[section \"parameter\"]\n", t)
	p.section("section", "parameter")
	p.eof()
}

func TestSectionWithParameterAndComment(t *testing.T) {
	p := newTestParser("[section \"parameter\"];The rest of this is a comment", t)
	p.section("section", "parameter")
	p.eof()
}

func TestSectionWithParameterWithAdditionalSpaces(t *testing.T) {
	p := newTestParser("    [		section 	\"parameter\"  ]	;The rest of this is a comment\n\n\n", t)
	p.section("section", "parameter")
	p.eof()
}

func TestKeyOnly(t *testing.T) {
	p := newTestParser("key", t)
	p.key("", "", "key", "")
	p.eof()
}

func TestKeyOnlyWithSpaces(t *testing.T) {
	p := newTestParser("    key\t\t\n\n\n", t)
	p.key("", "", "key", "")
	p.eof()
}

func TestKeyOnlyWithComment(t *testing.T) {
	p := newTestParser("key;comment", t)
	p.key("", "", "key", "")
	p.eof()
}

func TestKeyAndEqualsNoValue(t *testing.T) {
	p := newTestParser("key = ", t)
	p.key("", "", "key", "")
	p.eof()
}

func TestKeyAndEqualsNoValueWithComment(t *testing.T) {
	p := newTestParser("key = ;no value", t)
	p.key("", "", "key", "")
	p.eof()
}

func TestKeyAndSimpleValue(t *testing.T) {
	p := newTestParser("key = value\n", t)
	p.key("", "", "key", "value")
	p.eof()
}

func TestKeyAndSimpleValueWithSpaces(t *testing.T) {
	p := newTestParser("key = value with spaces\t\n", t)
	p.key("", "", "key", "value with spaces")
	p.eof()
}

func TestKeyAndSimpleValueWithQuotes(t *testing.T) {
	p := newTestParser("key = value with \"quotes\"\t\n", t)
	p.key("", "", "key", "value with \"quotes\"")
	p.eof()
}

func TestKeyAndStringValue(t *testing.T) {
	p := newTestParser("key = \"value in quotes\"\n", t)
	p.key("", "", "key", "value in quotes")
	p.eof()
}

func TestKeyAndRawStringValue(t *testing.T) {
	p := newTestParser("key = `value\nin\nraw\nquotes`\n", t)
	p.key("", "", "key", "value\nin\nraw\nquotes")
	p.eof()
}

func TestShortFile(t *testing.T) {
	file := `# Test config file
[section]
key1=value1 ; simple kv
key2        ; no value (boolean?)

[section2 "parameter"]
key1 = value2`

	p := newTestParser(file, t)
	p.section("section", "")
	p.key("section", "", "key1", "value1")
	p.key("section", "", "key2", "")
	p.section("section2", "parameter")
	p.key("section2", "parameter", "key1", "value2")
	p.eof()
}

func TestUnterminatedRawString(t *testing.T) {
	p := newTestParser("key=`\n\n", t)
	p.error()
}

func TestSectionOpenOnly(t *testing.T) {
	p := newTestParser("[", t)
	p.error()
}

func TestSectionNoClose(t *testing.T) {
	p := newTestParser("[section", t)
	p.error()
}

func TestSectionWithParameterNoClose(t *testing.T) {
	p := newTestParser("[section \"parameter\"", t)
	p.error()
}

func TestErrorLine(t *testing.T) {
	p := newTestParser("\n\n\n    ]", t)
	err := p.error()

	if err.Line != 4 {
		t.Fatalf("Expected error on line 4, got %d", err.Line)
	}

	if err.Col != 5 {
		t.Fatalf("Expected error in column 5, go %d", err.Col)
	}
}

func TestEOFErrorLine(t *testing.T) {
	p := newTestParser("\nkey=`raw value\n", t)
	err := p.error()

	if err.Line != 3 {
		t.Fatalf("Expected error on line 3, got %d", err.Line)
	}

	if err.Col != 1 {
		t.Fatalf("Expected error in column 1, go %d", err.Col)
	}
}

func TestCommentInSection(t *testing.T) {
	p := newTestParser("[section # This is a comment ]\n", t)
	err := p.error()
	if err.Line != 1 {
		t.Fatalf("Expected error on line 1, got %d", err.Line)
	}
	if err.Col != 10 {
		t.Fatalf("Expected error in column 10, go %d", err.Col)
	}
}

func TestInvalidEncoding(t *testing.T) {
	p := newTestParser("[section \"parameter\xE0\x80\xA2]", t)
	err := p.error()
	if err.Line != 1 {
		t.Fatalf("Expected error on line 1, got %d", err.Line)
	}
	if err.Col != 20 {
		t.Fatalf("Expected error in column 20, go %d", err.Col)
	}
}

func TestUnterminatedString(t *testing.T) {
	p := newTestParser("key=\"unterminated string value\n\"", t)
	err := p.error()
	if err.Line != 1 {
		t.Fatalf("Expected error on line 1, got %d", err.Line)
	}
	if err.Col != 31 {
		t.Fatalf("Expected error in column 31, got %d", err.Col)
	}
}

func TestSectionParameterWithEscapes(t *testing.T) {
	p := newTestParser("[section \"\\\\\\t\\r\\n\\\"\"]", t)
	p.section("section", "\\\t\r\n\"")
	p.eof()
}

func TestInvalidEscape(t *testing.T) {
	p := newTestParser("[section \"\\b\"]", t)
	err := p.error()
	if err.Line != 1 {
		t.Fatalf("Expected error on line 1, got %d", err.Line)
	}
	if err.Col != 12 {
		t.Fatalf("Expected error in column 12, got %d", err.Col)
	}
}

func TestJavaPropertiesFile(t *testing.T) {
	file := `# Test File
com.sun.foo = Value 1
com.sun.bar = Value 2
com.sun.baz = Value 3
`
	p := newTestParser(file, t)
	p.key("", "", "com.sun.foo", "Value 1")
	p.key("", "", "com.sun.bar", "Value 2")
	p.key("", "", "com.sun.baz", "Value 3")
	p.eof()
}
