package bloom

import (
	"crypto/rand"
	"testing"
)

func TestCreate(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Error(err)
	}
}

func TestAddFailsForBadArraySize(t *testing.T) {
	filter, _ := New()

	for i := 0; i < 32; i++ {
		err := filter.Add(make([]byte, i))
		if i == 16 {
			if err != nil {
				t.Error("should not have failed for i=", i)
			}

		} else {
			if err == nil {
				t.Error("should have failed for i=", i)
			}
		}
	}
}

func TestContainsFailsForBadArraySize(t *testing.T) {
	filter, _ := New()

	for i := 0; i < 32; i++ {
		_, err := filter.Contains(make([]byte, i))
		if i == 16 {
			if err != nil {
				t.Error("should not have failed for i=", i)
			}

		} else {
			if err == nil {
				t.Error("should have failed for i=", i)
			}
		}
	}
}

func TestContainsFailsForEmptyFilter(t *testing.T) {
	filter, _ := New()
	val := make([]byte, 16)
	rc, err := filter.Contains(val)
	if err != nil {
		t.Error(err)
	}
	if rc {
		t.Error("empty filter should not contain anything")
	}
}

func makeTestEntry() []byte {
	ar := make([]byte, 16)
	rand.Read(ar)
	return ar
}

func TestHappyPath(t *testing.T) {
	filter, _ := New()
	val := makeTestEntry()
	err1 := filter.Add(val)
	if err1 != nil {
		t.Error(err1)
	}

	rc, err2 := filter.Contains(val)
	if err2 != nil {
		t.Error(err2)
	}
	if !rc {
		t.Error("filter should contain value now")
	}
}

func TestDiscrimination(t *testing.T) {
	for repeat := 0; repeat < 128; repeat++ {
		filter, _ := New()
		for i := 0; i < 1024; i++ {
			filter.Add(makeTestEntry())
		}

		n := 0
		for i := 0; i < 1024; i++ {
			rc, _ := filter.Contains(makeTestEntry())
			if rc {
				n++
			}
		}
		if n < 16 {
			t.Error("We expected to get SOMETHING")
		}
		if n > 200 {
			t.Error("We got too much, i.e.", n)
		}
	}
}
