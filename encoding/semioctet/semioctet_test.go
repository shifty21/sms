// SPDX-License-Identifier: MIT
//
// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.

package semioctet_test

import (
	"fmt"
	"testing"

	"github.com/shifty21/sms/encoding/semioctet"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	patterns := []struct {
		name         string
		inDst        []byte
		inSrc        []byte
		out          []byte
		outReadCount int
		err          error
	}{
		{
			"nil",
			nil,
			nil,
			nil,
			0,
			nil,
		},
		{
			"nil dst",
			nil,
			[]byte{0x10, 0x32, 0x54, 0x76},
			nil,
			0,
			nil,
		},
		{
			"nil src",
			make([]byte, 3),
			nil,
			[]byte{},
			0,
			nil,
		},
		{
			"empty dst",
			[]byte{},
			[]byte{0x10, 0x32, 0x54, 0x76},
			[]byte{},
			0,
			nil,
		},
		{
			"empty src",
			make([]byte, 3),
			[]byte{},
			[]byte{},
			0,
			nil,
		},
		{
			"limit src",
			make([]byte, 8),
			[]byte{0x10, 0x32, 0x54, 0x76},
			[]byte("01234567"),
			4,
			nil,
		},
		{
			"fill limit src",
			make([]byte, 8),
			[]byte{0x10, 0x32, 0x54, 0xf6},
			[]byte("0123456"),
			4,
			nil,
		},
		{
			"fill limit even dst",
			make([]byte, 6),
			[]byte{0x10, 0x32, 0x54, 0xf6, 0x98},
			[]byte("012345"),
			3,
			nil,
		},
		{
			"no fill limit even dst",
			make([]byte, 6),
			[]byte{0x10, 0x32, 0x54, 0x76, 0x98},
			[]byte("012345"),
			3,
			nil,
		},
		{
			"fill limit odd dst",
			make([]byte, 5),
			[]byte{0x10, 0x32, 0xf4, 0x76},
			[]byte("01234"),
			3,
			nil,
		},
		{
			"alphabet",
			make([]byte, 15),
			[]byte{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe},
			[]byte("0123456789*#abc"),
			8,
			nil,
		},
		{
			"no fill limit even dst",
			make([]byte, 6),
			[]byte{0x10, 0x32, 0x54, 0x76, 0x98},
			[]byte("012345"),
			3,
			nil,
		},
		{
			"no fill limit odd dst",
			make([]byte, 5),
			[]byte{0x10, 0x32, 0x54, 0x76},
			nil,
			3,
			semioctet.ErrMissingFill,
		},
		{
			"skip inter fill",
			make([]byte, 10),
			[]byte{0x10, 0x32, 0xF4, 0x76, 0xF8},
			[]byte("01234678"),
			5,
			nil,
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			dst, count, err := semioctet.Decode(p.inDst, p.inSrc)
			assert.Equal(t, p.err, err)
			assert.Equal(t, p.outReadCount, count)
			assert.Equal(t, p.out, dst)
		}
		t.Run(p.name, f)
	}
}

func TestEncode(t *testing.T) {
	patterns := []struct {
		name string
		in   []byte
		out  []byte
		err  error
	}{
		{
			"nil",
			nil,
			nil,
			nil,
		},
		{
			"empty src",
			[]byte{},
			[]byte{},
			nil,
		},
		{
			"even src",
			[]byte("01234567"),
			[]byte{0x10, 0x32, 0x54, 0x76},
			nil,
		},
		{
			"odd src",
			[]byte("0123456"),
			[]byte{0x10, 0x32, 0x54, 0xf6},
			nil,
		},
		{
			"alphabet",
			[]byte("0123456789*#abc"),
			[]byte{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe},
			nil,
		},
		{
			"invalid digit",
			[]byte("012345D6789"),
			nil,
			semioctet.ErrInvalidDigit('D'),
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			dst, err := semioctet.Encode(p.in)
			assert.Equal(t, p.err, err)
			assert.Equal(t, p.out, dst)
		}
		t.Run(p.name, f)
	}
}

// TestErrInvalidDigit tests that the errors can be stringified.
// It is fragile, as it compares the strings exactly, but its main purpose is
// to confirm the Error function doesn't recurse, as that is bad.
func TestErrInvalidOctet(t *testing.T) {
	patterns := []byte{0x00, 0xa0, 0x0a, 0x9a, 0xa9, 0xff}
	for _, p := range patterns {
		f := func(t *testing.T) {
			err := semioctet.ErrInvalidDigit(p)
			expected := fmt.Sprintf("semioctet: invalid digit: '%c' - 0x%x", byte(p), int(p))
			s := err.Error()
			assert.Equal(t, expected, s)
		}
		t.Run(fmt.Sprintf("%x", p), f)
	}
}
