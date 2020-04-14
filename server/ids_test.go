package server

import (
	"regexp"
	"testing"
)

const (
	b64Regex = "\\w{10,20}"
	// thanks to https://adamscheller.com/regular-expressions/uuid-regex/ for saving me from writing this
	uuidRegex = "([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}"
)

func TestAppIDers(t *testing.T) {
	tests := []struct {
		desc, regex string
		iDFunc func() IDer
	} {
		{"AppUUIDNoName", uuidRegex, func() IDer {return NewAppUUID("")}},
		{"AppUUIDWithName", "blapp-"+uuidRegex, func() IDer {return NewAppUUID("blapp")}},
		{"RandB64NoName", b64Regex, func() IDer {return NewRandB64ID("")}},
		{"RandB64WithName", "fungus-"+b64Regex, func() IDer {return NewRandB64ID("fungus")}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ider := test.iDFunc()
			id, err := ider.ID()
			if err != nil {
				t.Error("failed to get mock app ID", "err", err)
			}

			match, err := regexp.MatchString(test.regex, id)
			if err != nil {
				t.Error("err matching generated ID", "err", err)
			}
			if !match {
				t.Error("ID did not match", "got", id)
			}
		})
	}
}

type MockIDer struct {
	sendThis string
}

func (m *MockIDer) ID() (string, error) {
	return m.sendThis, nil
}

func TestPipelineID_ID(t *testing.T) {
	first := "antleeb"
	second := "babananab"
	third := "canola"

	mockIDer := &MockIDer{}
	pipeIDer := &PipelineID{AppIDer: mockIDer}

	mockIDer.sendThis = first
	id, err := pipeIDer.ID("")
	if err != nil {
		t.Error("failed to get pipeline ID", "err", err)
	}
	if first != id {
		t.Error("frist ID call did not match", "got", id, "expected", first)
	}

	mockIDer.sendThis = second
	id, err = pipeIDer.ID(id)
	exp := first + fullIDerSep + second
	if err != nil {
		t.Error("failed to get pipeline ID", "err", err)
	}
	if exp != id {
		t.Error("second ID call did not match", "got", id, "expected", exp)
	}

	mockIDer.sendThis = third
	id, err = pipeIDer.ID(id)
	exp = first + fullIDerSep + second + fullIDerSep + third
	if err != nil {
		t.Error("failed to get pipeline ID", "err", err)
	}
	if exp != id {
		t.Error("third ID call did not match", "got", id, "expected", exp)
	}
}

func BenchmarkAppUUID_ID(b *testing.B) {
	iDer := NewAppUUID("something")
	for i := 0; i < b.N; i++ {
		_, _ = iDer.ID()
	}
}

func BenchmarkRandB64_ID(b *testing.B) {
	iDer := NewRandB64ID("something")
	for i := 0; i < b.N; i++ {
		_, _ = iDer.ID()
	}
}
