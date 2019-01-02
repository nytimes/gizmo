// +build go1.7

package server

import (
	"fmt"
	"net/http"
)

// AddIPToContext will attempt to pull an IP address out of the request and
// set it into a gorilla context.
func AddIPToContext(r *http.Request) {
	ip, err := GetIP(r)
	if err != nil {
		LogWithFields(r).Warningf("unable to get IP: %s", err)
	} else {
		vars := Vars(r)
		vars["ip"] = ip
		SetRouteVars(r, vars)
	}

	if ip = GetForwardedIP(r); len(ip) > 0 {
		vars := Vars(r)
		vars["forward-for-ip"] = ip
		SetRouteVars(r, vars)
	}
}

// ContextFields will take a request and convert a context map to logrus Fields.
func ContextFields(r *http.Request) map[string]interface{} {
	fields := map[string]interface{}{}
	for k, v := range Vars(r) {
		strK := fmt.Sprintf("%+v", k)
		typeK := fmt.Sprintf("%T-%+v", k, k)
		// gorilla.mux adds the route to context.
		// we want to remove it for now
		if typeK == "mux.contextKey-1" || typeK == "mux.contextKey-0" {
			continue
		}
		// web.varsKey for _all_ mux variables (gorilla or httprouter)
		if typeK == "web.contextKey-2" {
			strK = "muxvars"
		}
		fields[strK] = fmt.Sprintf("%#+v", v)
	}
	fields["path"] = r.URL.Path
	fields["rawquery"] = r.URL.RawQuery

	return fields
}
