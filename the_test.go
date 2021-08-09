package main

import (
	"math"
	"testing"
)

func TestCalcArty(t *testing.T) {

	for _, testCase := range []struct {
		name         string
		fromCoord    string
		toCoord      string
		expectedDist int
		expectedAz   float64
	}{
		{
			name:         "up",
			fromCoord:    "h-8-2",
			toCoord:      "h-7-5",
			expectedDist: 168,
			expectedAz:   0,
		},
		{
			name:         "right",
			fromCoord:    "h-8-5",
			toCoord:      "i-8-6",
			expectedDist: 168,
			expectedAz:   90,
		},
		{
			name:         "left",
			fromCoord:    "h-8-5",
			toCoord:      "g-8-6",
			expectedDist: 84,
			expectedAz:   270,
		},
		{
			name:         "upleft",
			fromCoord:    "h-8-5",
			toCoord:      "g-7-5",
			expectedDist: 178,
			expectedAz:   315,
		},
	} {
		from, err := NewCoord(testCase.fromCoord)
		if err != nil {
			t.Fatalf("err: %v, name: %s", err, testCase.name)
		}

		to, err := NewCoord(testCase.toCoord)
		if err != nil {
			t.Fatalf("err: %v, name: %s", err, testCase.name)
		}

		az, dist, err := calcArty(from, to)

		if err != nil {
			t.Fatalf("err: %v, name: %s", err, testCase.name)
		}
		if dist != testCase.expectedDist {
			t.Fatalf("Bad distance: %d, name: %s", dist, testCase.name)
		}
		if az != testCase.expectedAz {
			t.Fatalf("Bad az: %f, name: %s", az, testCase.name)
		}
	}

	c := offsetCoord(coord{0, 0}, 90*(math.Pi/180), 10)
	if c.y != -10 {
		t.Fatalf("Bad wind calc for south: %v", c)
	}
	c = offsetCoord(coord{0, 0}, 180*(math.Pi/180), 10)
	if c.x != -10 {
		t.Fatalf("Bad wind calc for left: %v", c)
	}
	//testing.Init()

	/*
		az, dist, err = calcArty("h-8-5", "h-8-6", "")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if dist != 42 {
			t.Fatalf("Bad distance: %d", dist)
		}
		if az != 90 {
			t.Fatalf("Bad az: %d", az)
		}

		az, dist, err = calcArty("f-11-3-3", "f-10-9-9", "")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if dist != 238 {
			t.Fatalf("Bad distance: %d", dist)
		}
		if az != 0 {
			t.Fatalf("Bad az: %d", az)
		}*/
}
