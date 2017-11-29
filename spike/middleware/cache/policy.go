package cache

import (
	"bytes"
	"strconv"
	"time"
)

type Directive int

const (
	NoStore Directive = iota
	NoCache
	Public
	Private
)

var directives = [...]string{
	"no-store",
	"no-cache",
	"public",
	"private",
}

func (d Directive) String() string { return directives[d] }

type Policy struct {
	ServerMaxAge time.Duration
	ClientMaxAge time.Duration
	Directive    Directive
}

var DefaultPolicy *Policy = &Policy{
	ServerMaxAge: 10 * time.Second,
	ClientMaxAge: 10 * time.Second,
	Directive:    Public,
}

func PolicyWithExpiration(expiration time.Duration) *Policy {
	return &Policy{
		ServerMaxAge: expiration,
		ClientMaxAge: expiration,
		Directive:    Public,
	}
}

func (p *Policy) CacheHeader() string {
	var buffer bytes.Buffer

	writeMaxAge := func() {
		buffer.WriteString("public, max-age=")
		buffer.WriteString(strconv.Itoa(int(p.ClientMaxAge.Seconds())))
	}

	if p.ServerMaxAge > 0 && p.ClientMaxAge == 0 {
		p.ClientMaxAge = p.ServerMaxAge
	}

	if p.ClientMaxAge > 0 {
		if p.Directive == Private {
			buffer.WriteString(p.Directive.String())
			buffer.WriteString(", ")
			writeMaxAge()
		} else {
			writeMaxAge()
		}
	} else if p.Directive == NoStore || p.Directive == NoCache {
		buffer.WriteString(p.Directive.String())
	}

	return buffer.String()
}

func (p *Policy) CacheControl() (string, string) {
	return "Cache-Control", p.CacheHeader()
}
