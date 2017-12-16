package main

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

// UniversityCourseTracker/com.tevinjeffrey.uctios (100; iOS 10.2.1) Course Tracker/1.0.7
func parseios(userAgent string) (string, string) {
	appVersion := ""
	osVersion := ""
	part := strings.Split(userAgent, ";")
	if len(part) == 2 {
		part = strings.Split(part[1], ")")
		if len(part) == 2 {
			osVersion = strings.Trim(part[0], " ")

			part = strings.Split(part[1], "/")
			if len(part) == 2 {
				appVersion = strings.Trim(part[1], " ")
			}
		}
	}

	return appVersion, osVersion
}

func deviceInfo(header http.Header) (string, string, string) {
	var os string
	var osVersion string
	var appVersion string

	userAgent := header.Get("User-Agent")
	if userAgent == "" {
		return os, osVersion, appVersion
	}

	if strings.Contains(strings.ToLower(userAgent), "android") {
		os = "android"
	}

	if strings.Contains(strings.ToLower(userAgent), "ios") {
		os = "ios"
	}

	// Rutgers Course Tracker/com.tevinjeffrey.rutgersct (1.0.7.0R; Android 27)
	if os == "android" {
		appV, osV := parseAndroid(userAgent)
		appVersion = appV
		osVersion = osV
	}

	if os == "ios" {
		appV, osV := parseios(userAgent)
		appVersion = appV
		osVersion = osV
	}

	return os, osVersion, appVersion
}
