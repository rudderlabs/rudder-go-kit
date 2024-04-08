package ip

import (
	"net/http"
	"strings"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
)

func FromReq(req *http.Request) string {
	addresses := strings.Split(req.Header.Get(headerXForwardedFor), ",")
	if addresses[0] == "" {
		splits := strings.Split(req.RemoteAddr, ":")
		return strings.Join(splits[:len(splits)-1], ":") // When there is no load-balancer
	}

	return strings.ReplaceAll(addresses[0], " ", "")
}
