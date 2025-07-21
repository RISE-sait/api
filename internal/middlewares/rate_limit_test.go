package middlewares

import (
	"net/http"
	"testing"
)

func TestGetRealIP_RemoteAddrWithPorts(t *testing.T) {
	r1 := &http.Request{RemoteAddr: "1.2.3.4:1111"}
	r2 := &http.Request{RemoteAddr: "1.2.3.4:2222"}

	ip1 := getRealIP(r1)
	ip2 := getRealIP(r2)

	if ip1 != "1.2.3.4" {
		t.Fatalf("expected host without port, got %s", ip1)
	}
	if ip1 != ip2 {
		t.Fatalf("expected same IP for different ports, got %s and %s", ip1, ip2)
	}
}
