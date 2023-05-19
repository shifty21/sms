// SPDX-License-Identifier: MIT
//
// Copyright © 2018 Kent Gibson <warthog618@gmail.com>.

package tpdu_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shifty21/sms/encoding/bcd"
	"github.com/shifty21/sms/encoding/tpdu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type marshalTimestampPattern struct {
	name string
	in   tpdu.Timestamp
	out  []byte
	err  error
}

func TestMarhalBinary(t *testing.T) {
	patterns := []marshalTimestampPattern{
		{
			"19700101",
			tpdu.Timestamp{
				Time: time.Date(1970, time.January, 1, 1, 2, 3, 0, time.UTC),
			},
			[]byte{0x07, 0x10, 0x10, 0x10, 0x20, 0x30, 0x00},
			nil,
		},
		{
			"19991231",
			tpdu.Timestamp{
				Time: time.Date(1999, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", -15*60)),
			},
			[]byte{0x99, 0x21, 0x13, 0x32, 0x95, 0x95, 0x18},
			nil,
		},
		{
			"20001231",
			tpdu.Timestamp{Time: time.Date(2000, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", 15*60))},
			[]byte{0x00, 0x21, 0x13, 0x32, 0x95, 0x95, 0x10},
			nil,
		},
		{
			"20170831",
			tpdu.Timestamp{
				Time: time.Date(2017, time.August, 31, 11, 21, 54, 0, time.FixedZone("any", 8*3600)),
			},
			[]byte{0x71, 0x80, 0x13, 0x11, 0x12, 0x45, 0x23},
			nil,
		},
		{
			"20700101",
			tpdu.Timestamp{
				Time: time.Date(2070, time.January, 1, 1, 2, 3, 0, time.UTC),
			},
			[]byte{0x07, 0x10, 0x10, 0x10, 0x20, 0x30, 0x00},
			nil,
		},
		{
			"21001231",
			tpdu.Timestamp{
				Time: time.Date(2100, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", 15*60)),
			},
			[]byte{0x00, 0x21, 0x13, 0x32, 0x95, 0x95, 0x10},
			nil,
		},
		{
			"20701231",
			tpdu.Timestamp{
				Time: time.Date(2070, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", 24*3600)),
			},
			nil,
			bcd.ErrInvalidInteger(96),
		},
		// how to trigger invalid integer in date (other than tz)??
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			b, err := p.in.MarshalBinary()
			require.Equal(t, p.err, err)
			assert.Equal(t, p.out, b)
		}
		t.Run(p.name, f)
	}
}

func TestUnmarhalBinary(t *testing.T) {
	patterns := []struct {
		name string
		in   []byte
		out  tpdu.Timestamp
		err  error
	}{
		{
			"19700101",
			[]byte{0x07, 0x10, 0x10, 0x10, 0x20, 0x30, 0x00},
			tpdu.Timestamp{
				Time: time.Date(1970, time.January, 1, 1, 2, 3, 0, time.UTC),
			},
			nil,
		},
		{
			"19991231",
			[]byte{0x99, 0x21, 0x13, 0x32, 0x95, 0x95, 0x18},
			tpdu.Timestamp{
				Time: time.Date(1999, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", -15*60)),
			},
			nil,
		},
		{
			"20001231",
			[]byte{0x00, 0x21, 0x13, 0x32, 0x95, 0x95, 0x10},
			tpdu.Timestamp{
				Time: time.Date(2000, time.December, 31, 23, 59, 59, 0, time.FixedZone("SCTS", 15*60)),
			},
			nil,
		},
		{
			"20170831",
			[]byte{0x71, 0x80, 0x13, 0x11, 0x12, 0x45, 0x23},
			tpdu.Timestamp{
				Time: time.Date(2017, time.August, 31, 11, 21, 54, 0, time.FixedZone("SCTS", 8*3600)),
			},
			nil,
		},
		{
			"short",
			[]byte{0x71, 0x80, 0x13, 0x11, 0x12, 0x45},
			tpdu.Timestamp{},
			tpdu.ErrUnderflow,
		},
		{
			"invalid digit",
			[]byte{0xa1, 0x80, 0x13, 0x11, 0x12, 0x45, 0x00},
			tpdu.Timestamp{},
			bcd.ErrInvalidOctet(0xa1),
		},
		{
			"invalid signed digit",
			[]byte{0x71, 0x80, 0x13, 0x11, 0x12, 0x45, 0xa0},
			tpdu.Timestamp{},
			bcd.ErrInvalidOctet(0xa0),
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			s := tpdu.Timestamp{}
			err := s.UnmarshalBinary(p.in)
			if err != p.err {
				t.Fatalf("error unmarshalling %v: %v", p.in, err)
			}
			if !s.Equal(p.out.Time) {
				t.Fatalf("failed to unmarshal %v: expected %v, got %v", p.in, p.out, s)
			}
			szn, szo := s.Zone()
			ozn, ozo := p.out.Zone()
			assert.Equal(t, ozn, szn)
			assert.Equal(t, ozo, szo)
		}
		t.Run(p.name, f)
	}
}

func TestTimestampString(t *testing.T) {
	patterns := []struct {
		in  time.Time
		out string
	}{
		{
			time.Date(2000, 11, 2, 3, 4, 5, 65, time.FixedZone("ABC", 22800)),
			"2000-11-02 03:04:05 +0620",
		},
		{
			time.Date(2000, 11, 2, 3, 4, 5, 0, time.FixedZone("ABC", 22800)),
			"2000-11-02 03:04:05 +0620",
		},
		{
			time.Date(2000, 11, 2, 3, 4, 5, 0, time.FixedZone("TEST", 0)),
			"2000-11-02 03:04:05 +0000",
		},
	}
	for _, p := range patterns {
		f := func(t *testing.T) {
			out := tpdu.Timestamp{p.in}.String()
			assert.Equal(t, p.out, out)
		}
		t.Run(fmt.Sprintf("%02x", p.in), f)
	}
}
