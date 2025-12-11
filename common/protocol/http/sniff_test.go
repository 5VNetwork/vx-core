package http_test

import (
	"testing"

	. "github.com/5vnetwork/vx-core/common/protocol/http"
)

func TestHTTPHeaders(t *testing.T) {
	cases := []struct {
		input  string
		domain string
		err    bool
	}{
		{
			input: `GET /tutorials/other/top-20-mysql-best-practices/ HTTP/1.1
Host: net.tutsplus.com
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Cookie: PHPSESSID=r2t5uvjq435r4q7ib3vtdjq120
Pragma: no-cache
Cache-Control: no-cache`,
			domain: "net.tutsplus.com",
		},
		{
			input: `POST /foo.php HTTP/1.1
Host: localhost
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Referer: http://localhost/test.php
Content-Type: application/x-www-form-urlencoded
Content-Length: 43
 
first_name=John&last_name=Doe&action=Submit`,
			domain: "localhost",
		},
		{
			input: `X /foo.php HTTP/1.1
Host: localhost
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Referer: http://localhost/test.php
Content-Type: application/x-www-form-urlencoded
Content-Length: 43
 
first_name=John&last_name=Doe&action=Submit`,
			domain: "",
			err:    true,
		},
		{
			input: `GET /foo.php HTTP/1.1
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Referer: http://localhost/test.php
Content-Type: application/x-www-form-urlencoded
Content-Length: 43

Host: localhost
first_name=John&last_name=Doe&action=Submit`,
			domain: "",
			err:    true,
		},
		{
			input:  `GET /tutorials/other/top-20-mysql-best-practices/ HTTP/1.1`,
			domain: "",
			err:    true,
		},
	}

	for _, test := range cases {
		header, err := SniffHTTP1Host([]byte(test.input))
		if test.err {
			if err == nil {
				t.Errorf("Expect error but nil, in test: %v", test)
			}
		} else {
			if err != nil {
				t.Errorf("Expect no error but actually %s in test %v", err.Error(), test)
			}
			if header.Domain() != test.domain {
				t.Error("expected domain ", test.domain, " but got ", header.Domain())
			}
		}
	}
}

func TestHTTP1(t *testing.T) {
	cases := []struct {
		input   string
		domain  string
		method  string
		path    string
		query   string
		version string
		err     bool
	}{
		{
			input: `GET /tutorials/other/top-20-mysql-best-practices/ HTTP/1.1
Host: net.tutsplus.com
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Cookie: PHPSESSID=r2t5uvjq435r4q7ib3vtdjq120
Pragma: no-cache
Cache-Control: no-cache`,
			domain:  "net.tutsplus.com",
			method:  "GET",
			path:    "/tutorials/other/top-20-mysql-best-practices/",
			query:   "",
			version: "HTTP/1.1",
		},
		{
			input: `POST /foo.php HTTP/1.1
Host: localhost
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Referer: http://localhost/test.php
Content-Type: application/x-www-form-urlencoded
Content-Length: 43

first_name=John&last_name=Doe&action=Submit`,
			domain:  "localhost",
			method:  "POST",
			path:    "/foo.php",
			query:   "",
			version: "HTTP/1.1",
		},
		{
			input: `X /foo.php HTTP/1.1
Host: localhost
User-Agent: Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.5) Gecko/20091102 Firefox/3.5.5 (.NET CLR 3.5.30729)
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-us,en;q=0.5
Accept-Encoding: gzip,deflate
Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7
Keep-Alive: 300
Connection: keep-alive
Referer: http://localhost/test.php
Content-Type: application/x-www-form-urlencoded
Content-Length: 43

first_name=John&last_name=Doe&action=Submit`,
			domain:  "",
			method:  "",
			path:    "",
			query:   "",
			version: "",
			err:     true,
		},
		{
			input: `GET /api/search?q=test&limit=10 HTTP/1.1
Host: example.com
User-Agent: Mozilla/5.0
Accept: application/json`,
			domain:  "example.com",
			method:  "GET",
			path:    "/api/search",
			query:   "q=test&limit=10",
			version: "HTTP/1.1",
		},
		{
			input: `PUT /api/users/123 HTTP/1.0
Host: api.example.com
Content-Type: application/json`,
			domain:  "api.example.com",
			method:  "PUT",
			path:    "/api/users/123",
			query:   "",
			version: "HTTP/1.0",
		},
	}

	for _, test := range cases {
		header, err := SniffHttp1([]byte(test.input))
		if test.err {
			if err == nil {
				t.Errorf("Expect error but nil, in test: %v", test)
			}
		} else {
			if err != nil {
				t.Errorf("Expect no error but actually %s in test %v", err.Error(), test)
			}
			if header.Host() != test.domain {
				t.Error("expected domain ", test.domain, " but got ", header.Host())
			}
			if header.Method() != test.method {
				t.Error("expected method ", test.method, " but got ", header.Method())
			}
			if header.Path() != test.path {
				t.Error("expected path ", test.path, " but got ", header.Path())
			}
			if header.Query() != test.query {
				t.Error("expected query ", test.query, " but got ", header.Query())
			}
			if header.Version() != test.version {
				t.Error("expected version ", test.version, " but got ", header.Version())
			}
		}
	}
}
