package grpcsimpleserver

import (
	"testing"
)

func TestSaveValue(t *testing.T) {
	cont := fibonacciMemcache{}
	err := cont.SaveValue("test", 1)
	if err != nil {
		t.Fatalf("Error during save data: %v", err.Error())
	}
}

func TestGetValue(t *testing.T) {
	val := int64(2)
	cont := fibonacciMemcache{}
	err := cont.SaveValue("test", val)
	if err != nil {
		t.Fatalf("Error during saving data: %v", err.Error())
	}
	readVal, err := cont.GetValue("test")

	if err != nil {
		t.Fatalf("Error during reading data: %v", err.Error())
	}

	if readVal != val {
		t.Fatalf("Error read value isnot qual to original. Expected: %v, Actual: %v", val, readVal)
	}
}
