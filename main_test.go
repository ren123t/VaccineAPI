package main

import "testing"

func TestFakeFunction(t *testing.T) {
	want := 43.445232
	if got := 43.445232; got != want {
		t.Errorf("Hello() = %v, want %v", got, want)
	}
}
