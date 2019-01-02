package server

import (
	"fmt"
	"net/http"
	"strconv"
)

// JSONContentType can be used for setting the Content-Type header for JSON encoding.
const JSONContentType = "application/json; charset=UTF-8"

// GetInt64Var is a helper to pull gorilla mux Vars.
// If the value is empty, it falls back to the URL
// query string.
// We are ignoring the error here bc we're assuming
// the path had a [0-9]+ descriptor on this var.
func GetInt64Var(r *http.Request, key string) int64 {
	v := Vars(r)[key]
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
	v := Vars(r)[key]
	if len(v) == 0 {
		va := r.URL.Query()[key]
		if len(va) > 0 {
			v = va[0]
		}
	}
	i, _ := strconv.ParseUint(v, 10, 64)
	return i
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
