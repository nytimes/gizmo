package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Let's have generic errors for expected conditions. Typically
// these would be http.StatusBadRequest (400)
var (
	JSONContentType = "application/json; charset=UTF-8"
	DateISOFormat   = "2006-01-02"
)

// ParseISODate is a handy function to accept
func ParseISODate(dateStr string) (date time.Time, err error) {
	date, err = time.ParseInLocation(DateISOFormat, dateStr, time.Local)
	return
}

// ParseDateRange will look for and parse 'startDate' and 'endDate' ISO date
// strings in the given vars map.
func ParseDateRange(vars map[string]string) (startDate time.Time, endDate time.Time, err error) {
	startDate, err = ParseISODate(vars["startDate"])
	if err != nil {
		err = errors.New("please use a valid start date with a format of YYYY-MM-DD")
		return
	}

	endDate, err = ParseISODate(vars["endDate"])
	if err != nil {
		err = errors.New("please use a valid end date with a format of YYYY-MM-DD")
		return
	}

	return
}

// GetInt64Var is a helper to pull gorilla mux Vars.
// If the value is empty, it falls back to the URL
// query string.
// We are ignoring the error here bc we're assuming
// the path had a [0-9]+ descriptor on this var.
func GetInt64Var(r *http.Request, key string) int64 {
	v := mux.Vars(r)[key]
	if len(v) == 0 {
		va := r.URL.Query()[key]
		if len(va) > 0 {
			v = va[0]
		}
	}
	i, _ := strconv.ParseInt(v, 10, 64)

	return i
}

// GetUInt64Var is a helper to pull gorilla mux Vars.
// If the value is empty, it falls back to the URL
// query string.
// We are ignoring the error here bc we're assuming
// the path had a [0-9]+ descriptor on this var.
func GetUInt64Var(r *http.Request, key string) uint64 {
	v := mux.Vars(r)[key]
	if len(v) == 0 {
		va := r.URL.Query()[key]
		if len(va) > 0 {
			v = va[0]
		}
	}
	i, _ := strconv.ParseUint(v, 10, 64)
	return i
}

// ParseDateRangeFullDay will look for and parse 'startDate' and 'endDate' ISO
// date strings in the given vars map. It will then set the startDate time to
// midnight and the endDate time to 23:59:59.
func ParseDateRangeFullDay(vars map[string]string) (startDate time.Time, endDate time.Time, err error) {
	startDate, endDate, err = ParseDateRange(vars)
	if err != nil {
		return
	}

	// set time to beginning of day
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0,
		0, 0, time.Local)
	// set the time to the end of day
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59,
		59, 1000, time.Local)
	return
}

// ParseTruthyFalsy is a helper method to attempt to parse booleans in
// APIs that have no set contract on what a boolean should look like.
func ParseTruthyFalsy(flag interface{}) (result bool, err error) {
	s := fmt.Sprint(flag)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}
