package http

import (
	"strings"

	httppb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/transport/headers/http"
	"github.com/5vnetwork/vx-core/common/dice"
)

func pickString(arr []string) string {
	n := len(arr)
	switch n {
	case 0:
		return ""
	case 1:
		return arr[0]
	default:
		return arr[dice.Roll(n)]
	}
}

func pickRequestURI(v *RequestConfig) string {
	return pickString(v.GetUri())
}

func pickRequestHeaders(v *RequestConfig) []string {
	n := len(v.GetHeader())
	if n == 0 {
		return nil
	}
	headers := make([]string, n)
	for idx, headerConfig := range v.GetHeader() {
		headerName := headerConfig.GetName()
		headerValue := pickString(headerConfig.GetValue())
		headers[idx] = headerName + ": " + headerValue
	}
	return headers
}

func requestVersionValue(v *RequestConfig) string {
	if v == nil || v.GetVersion() == nil {
		return "1.1"
	}
	return v.GetVersion().GetValue()
}

func requestMethodValue(v *RequestConfig) string {
	if v == nil || v.GetMethod() == nil {
		return "GET"
	}
	return v.GetMethod().GetValue()
}

func requestFullVersion(v *RequestConfig) string {
	return "HTTP/" + requestVersionValue(v)
}

func responseHasHeader(v *ResponseConfig, header string) bool {
	cHeader := strings.ToLower(header)
	for _, tHeader := range v.GetHeader() {
		if strings.EqualFold(tHeader.GetName(), cHeader) {
			return true
		}
	}
	return false
}

func pickResponseHeaders(v *ResponseConfig) []string {
	n := len(v.GetHeader())
	if n == 0 {
		return nil
	}
	headers := make([]string, n)
	for idx, headerConfig := range v.GetHeader() {
		headerName := headerConfig.GetName()
		headerValue := pickString(headerConfig.GetValue())
		headers[idx] = headerName + ": " + headerValue
	}
	return headers
}

func responseVersionValue(v *ResponseConfig) string {
	if v == nil || v.GetVersion() == nil {
		return "1.1"
	}
	return v.GetVersion().GetValue()
}

func responseFullVersion(v *ResponseConfig) string {
	return "HTTP/" + responseVersionValue(v)
}

func responseStatusOrDefault(v *ResponseConfig) *Status {
	if v == nil || v.GetStatus() == nil {
		return &httppb.Status{Code: "200", Reason: "OK"}
	}
	return v.GetStatus()
}
