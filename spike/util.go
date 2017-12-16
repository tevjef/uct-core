package main

import (
	"strings"

	"github.com/gin-gonic/gin"
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

func deviceInfo(c *gin.Context) (string, string, string) {
	var os string
	var osVersion string
	var appVersion string

	m := map[string]string{}
	userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))
	if userAgent != "" {
		return os, osVersion, appVersion
	}

	if strings.Contains(userAgent, "android") {
		os = "android"
	}

	if strings.Contains(userAgent, "ios") {
		os = "ios"
	}

	// Rutgers Course Tracker/com.tevinjeffrey.rutgersct (1.0.7.0R; Android 27)
	if v, _ := m["os"]; v == "android" {
		app, os := parseAndroid(userAgent)
		appVersion = app
		osVersion = os
	}

	if v, _ := m["os"]; v == "ios" {
		app, os := parseios(userAgent)
		appVersion = app
		osVersion = os
	}

	return os, osVersion, appVersion
}
