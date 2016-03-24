package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/sqliface"
)

// TestScanItems will test our repo's logic for scanning data
// out of the DB and into structs.
func TestScanItems(t *testing.T) {
	testTime := time.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)
	testTime2 := time.Date(2015, 1, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		given *sqliface.MockRows

		want    []*SavedItem
		wantErr error
	}{
		// test to verify we get an empty slice when
		// the DB has no data.
		{
			sqliface.NewMockRows(),

			[]*SavedItem{},
			nil,
		},
		// normal success test
		{
			sqliface.NewMockRows(
				sqliface.MockRow{
					uint64(123),
					"http://nytimes.com/awesome-article",
					testTime,
				},
				sqliface.MockRow{
					uint64(456),
					"http://nytimes.com/awesome-article-2",
					testTime2,
				},
			),

			[]*SavedItem{
				&SavedItem{
					uint64(123),
					"http://nytimes.com/awesome-article",
					testTime,
				},
				&SavedItem{
					uint64(456),
					"http://nytimes.com/awesome-article-2",
					testTime2,
				},
			},
			nil,
		},
		// testing with Scan returning unexpected error.
		// using the wrong type in MockRow (a uint64 instead of a string in
		// this case) to trigger the error.
		{
			sqliface.NewMockRows(
				sqliface.MockRow{
					uint64(123),
					uint64(123),
					testTime,
				},
			),

			[]*SavedItem(nil),
			sqliface.NewTypeError("string", uint64(123)),
		},
	}

	for _, test := range tests {

		// run the test, passing in the MockRows implementation.
		got, err := scanItems(test.given)

		// verify the test's results

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("expected \n%#v\ngot,\n%#v", test.want, got)
		}

		if !reflect.DeepEqual(err, test.wantErr) {
			t.Errorf("expected error of \n%#v\ngot,\n%#v", test.wantErr, err)
		}
	}

}
