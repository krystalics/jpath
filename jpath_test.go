package jpath

import (
	"testing"
)

const (
	simple = "{\"simpleArray\":[1,2,3],\"labels\":\"{\\\"level_1\\\":{\\\"tag_id\\\":\\\"example-1\\\",\\\"tag_name\\\":\\\"school\\\",\\\"prob\\\":1,\\\"level\\\":1},\\\"level_2\\\":{\\\"tag_id\\\":\\\"example-2\\\",\\\"tag_name\\\":\\\"class\\\",\\\"prob\\\":1,\\\"level\\\":2}}\"}"
)

func TestSimple(t *testing.T) {
	jPath, _ := New(simple)
	demo(t, jPath)
}

func TestNewConcurrencySafe(t *testing.T) {
	jPath, _ := NewConcurrencySafe(simple)
	for i := 0; i < 100; i++ {
		go func() {
			demo(t, jPath)
		}()
	}
}

func demo(t *testing.T, jPath *JPath) {
	expectedValue := "school"
	realTag1 := jPath.Find("labels.level_1.tag_name")
	if realTag1 != expectedValue {
		t.Errorf("Expected: %v\n Got: %v", expectedValue, realTag1)
	}

	expectedValue = "class"
	realTag2 := jPath.Find("labels.level_2.tag_name")
	if realTag2 != expectedValue {
		t.Errorf("Expected: %v\n Got: %v", expectedValue, realTag2)
	}

	val := jPath.Find("simpleArray[1]")
	//from  str convert to number, it will be float64
	if val != 2.0 {
		t.Errorf("Expected: %v\n Got: %v", 2, val)
	}
}
