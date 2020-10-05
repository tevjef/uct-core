package spike

import (
	"net/http"
	"strings"
)

// Rutgers Course Tracker/com.tevinjeffrey.rutgersct (1.0.7.0R; Android 27)
func parseAndroid(userAgent string) (string, string) {
	appVersion := ""
	osVersion := ""
	part := strings.Split(userAgent, "(")
	if len(part) == 2 {
		part = strings.Split(part[1], ";")
		if len(part) == 2 {
			appVersion = part[0]
			osVersion = strings.Trim(part[1], " )")
		}
	}

	return appVersion, osVersion
}

// UniversityCourseTracker/1.0.8 (com.tevinjeffrey.uctios; build:100; iOS 10.3.1) Alamofire/4.5.1
func parseios(userAgent string) (string, string) {
	appVersion := ""
	osVersion := ""
	part := strings.Split(userAgent, "(")
	if len(part) == 2 {
		part1 := strings.Split(part[0], "/")
		if len(part1) == 2 {
			appVersion = strings.Trim(part1[1], " ")
		}
		part1 = strings.Split(part[1], ")")
		if len(part1) == 2 {
			part1 = strings.Split(part1[0], ";")
			osVersion = strings.Trim(part1[2], " ")
		}
	}

	return appVersion, osVersion
}

func deviceInfo(header http.Header) (string, string, string) {
	var os = "unknown"
	var osVersion string
	var appVersion string

	userAgent := header.Get("User-Agent")
	if userAgent == "" {
		return os, osVersion, appVersion
	}

	if strings.Contains(strings.ToLower(userAgent), "android") {
		os = "android"
		appV, osV := parseAndroid(userAgent)
		appVersion = appV
		osVersion = osV
	}

	if strings.Contains(strings.ToLower(userAgent), "ios") {
		os = "ios"
		appV, osV := parseios(userAgent)
		appVersion = appV
		osVersion = osV
	}

	return os, osVersion, appVersion
}
