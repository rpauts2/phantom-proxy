package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestValidateListenAddress(t *testing.T) {
	for _, tc := range []struct {
		name string
		addr string
		ok   bool
	}{
		{name: "IPv4 loopback", addr: "127.0.0.1:8088", ok: true},
		{name: "IPv6 loopback", addr: "[::1]:8088", ok: true},
		{name: "public bind denied", addr: "0.0.0.0:8088"},
		{name: "LAN bind denied", addr: "192.168.1.10:8088"},
		{name: "hostname denied", addr: "localhost:8088"},
		{name: "missing port denied", addr: "127.0.0.1"},
		{name: "empty port denied", addr: "127.0.0.1:"},
		{name: "invalid port denied", addr: "127.0.0.1:70000"},
		{name: "non-numeric port denied", addr: "127.0.0.1:http"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateListenAddress(tc.addr)
			if (err == nil) != tc.ok {
				t.Fatalf("validateListenAddress(%q) error = %v, want success %v", tc.addr, err, tc.ok)
			}
		})
	}
}

func TestParseCampaignAllowlist(t *testing.T) {
	got, err := parseCampaignAllowlist(" demo, training ,demo2 ")
	if err != nil {
		t.Fatalf("parseCampaignAllowlist() error = %v", err)
	}
	want := []string{"demo", "training", "demo2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("allowlist = %#v, want %#v", got, want)
	}

	for _, raw := range []string{"", " , , ", "demo,bad id", "demo,demo", strings.Repeat("a", 65)} {
		t.Run(raw, func(t *testing.T) {
			if _, err := parseCampaignAllowlist(raw); err == nil {
				t.Fatalf("parseCampaignAllowlist(%q) unexpectedly succeeded", raw)
			}
		})
	}
}

func TestValidateScopeDuration(t *testing.T) {
	for _, tc := range []struct {
		duration time.Duration
		ok       bool
	}{
		{time.Minute, true},
		{8 * time.Hour, true},
		{24 * time.Hour, true},
		{time.Second, false},
		{24*time.Hour + time.Second, false},
	} {
		err := validateScopeDuration(tc.duration)
		if (err == nil) != tc.ok {
			t.Errorf("validateScopeDuration(%s) error = %v, want success %v", tc.duration, err, tc.ok)
		}
	}
}
