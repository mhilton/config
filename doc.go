// Copyright 2013 Martin Hilton

// Package config provides support for reading simple configuration files.
// The format of the configuration files is similar to those used by git.
//
// Configuration files parsed by this package must be UTF-8 encoded.
// Configuration files consist of sections and keys.
//
// A section starts with its name, and an optional string parameter in
// square brackets and stops at the start of the next section.
//
// Keys are specified by a name, optionally (but usually) followed an '='
// and a value. Values are either specified as plain text, a string or
// a raw string. Plain text values are taken to be the characters between
// the '=' and the end of the line (or start of comment) with any leading
// or trailing whitespace stripped.
//
// Comments start with either the ';' or '#' characters and continue to
// the end of the line. Blank lines, or those consisting only of comments
// are ignored.
//
// Strings are used for string values and parameters, they must not contain
// an embedded new line. Strings support the following escape sequences.
//   \" "
//   \\ \
//   \n newline
//   \r carriage
//   \t tab
//
// Raw strings start and end with '`', they are a literal copy of all the
// characters, including newlines, in between.
//
// An example configuration file is:
//   # Example file supported by config
//
//   ; Global options
//   debug = true
//
//   [host "example.org"]
//   port = 8080
//   user-name = example
package config
