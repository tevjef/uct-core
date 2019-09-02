package cache

import (
	"testing"
	"time"
)

func TestPolicy_CacheControl(t *testing.T) {
	type fields struct {
		ServerMaxAge time.Duration
		ClientMaxAge time.Duration
		Directive    Directive
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "time.Second, time.Second, Public", fields: fields{time.Second, time.Second, Public}, want: "public, max-age=1"},
		{name: "time.Second, time.Second, Private", fields: fields{time.Second, time.Second, Private}, want: "private, max-age=1"},
		{name: "time.Second, time.Second, NoStore", fields: fields{time.Second, time.Second, NoStore}, want: "no-store"},
		{name: "time.Second, time.Second, NoCache", fields: fields{time.Second, time.Second, NoCache}, want: "no-cache"},
		{name: "time.Second, 0, Public", fields: fields{time.Second, 0, Public}, want: "public, max-age=1"},
		{name: "time.Second, 0, Private", fields: fields{time.Second, 0, Private}, want: "private, max-age=1"},
		{name: "time.Second, 0, NoStore", fields: fields{time.Second, 0, NoStore}, want: "no-store"},
		{name: "time.Second, 0, NoCache", fields: fields{time.Second, 0, NoCache}, want: "no-cache"},
		{name: "time.Second, 0, 0", fields: fields{time.Second, 0, 0}, want: "no-store"},
		{name: "time.Second, 0, 0", fields: fields{time.Second, 0, Private}, want: "private, max-age=1"},
		{name: "time.Second, 0, NoStore", fields: fields{time.Second, 0, NoStore}, want: "no-store"},
		{name: "time.Second, 0, NoCache", fields: fields{time.Second, 0, NoCache}, want: "no-cache"},
		{name: "0, time.Second, Public", fields: fields{0, time.Second, Public}, want: "public, max-age=1"},
		{name: "0, time.Second, Private", fields: fields{0, time.Second, Private}, want: "private, max-age=1"},
		{name: "0, time.Second, NoStore", fields: fields{0, time.Second, NoStore}, want: "no-store"},
		{name: "0, time.Second, NoCache", fields: fields{0, time.Second, NoCache}, want: "no-cache"},
		{name: "0, 0, Public", fields: fields{0, 0, Public}, want: ""},
		{name: "0, 0, Private", fields: fields{0, 0, Private}, want: ""},
		{name: "0, 0, NoStore", fields: fields{0, 0, NoStore}, want: "no-store"},
		{name: "0, 0, NoCache", fields: fields{0, 0, NoCache}, want: "no-cache"},
	}
	for _, tt := range tests {
		p := &Policy{
			ServerMaxAge: tt.fields.ServerMaxAge,
			ClientMaxAge: tt.fields.ClientMaxAge,
			Directive:    tt.fields.Directive,
		}
		if got := p.CacheHeader(); got != tt.want {
			t.Errorf("%q. Policy.CacheHeader() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
