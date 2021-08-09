package main

import (
	"testing"
)

func TestCalcArty(t *testing.T) {
	testing.Init()

	az, dist, err := calcArty("h-8-2", "h-8-5", "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if dist != 42 {
		t.Fatalf("Bad distance: %d", dist)
	}
	if az != 0 {
		t.Fatalf("Bad az: %d", az)
	}

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
	}
}