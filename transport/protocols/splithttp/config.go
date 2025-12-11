package splithttp

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/5vnetwork/vx-core/common/crypto"
)

func (m *XmuxConfig) Init() {
	if m.MaxConcurrency == nil {
		m.MaxConcurrency = &RangeConfig{}
	}
	if m.HMaxRequestTimes == nil {
		m.HMaxRequestTimes = &RangeConfig{}
	}
	if m.HMaxReusableSecs == nil {
		m.HMaxReusableSecs = &RangeConfig{}
	}
	if m.MaxConcurrency.From == 0 && m.MaxConcurrency.To == 0 {
		m.MaxConcurrency.From = 16
		m.MaxConcurrency.To = 32
	}
	if m.HMaxRequestTimes.From == 0 && m.HMaxRequestTimes.To == 0 {
		m.HMaxRequestTimes.From = 600
		m.HMaxRequestTimes.To = 900
	}
	if m.HMaxReusableSecs.From == 0 && m.HMaxReusableSecs.To == 0 {
		m.HMaxReusableSecs.From = 1800
		m.HMaxReusableSecs.To = 3000
	}
}

func (c *SplitHttpConfig) GetNormalizedPath() string {
	pathAndQuery := strings.SplitN(c.Path, "?", 2)
	path := pathAndQuery[0]

	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	return path
}

func (c *SplitHttpConfig) GetNormalizedQuery() string {
	pathAndQuery := strings.SplitN(c.Path, "?", 2)
	query := ""

	if len(pathAndQuery) > 1 {
		query = pathAndQuery[1]
	}

	/*
		if query != "" {
			query += "&"
		}
		query += "x_version=" + core.Version()
	*/

	return query
}

func (c *SplitHttpConfig) GetRequestHeader(rawURL string) (http.Header, error) {
	header := http.Header{}
	for k, v := range c.Headers {
		header.Add(k, v)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	// https://www.rfc-editor.org/rfc/rfc7541.html#appendix-B
	// h2's HPACK Header Compression feature employs a huffman encoding using a static table.
	// 'X' is assigned an 8 bit code, so HPACK compression won't change actual padding length on the wire.
	// https://www.rfc-editor.org/rfc/rfc9204.html#section-4.1.2-2
	// h3's similar QPACK feature uses the same huffman table.
	u.RawQuery = "x_padding=" + strings.Repeat("X", int(c.GetNormalizedXPaddingBytes().rand()))
	header.Set("Referer", u.String())

	return header, nil
}

func (c *SplitHttpConfig) WriteResponseHeader(writer http.ResponseWriter) {
	// CORS headers for the browser dialer
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	// writer.Header().Set("X-Version", core.Version())
	writer.Header().Set("X-Padding", strings.Repeat("X", int(c.GetNormalizedXPaddingBytes().rand())))
}

func (c *SplitHttpConfig) GetNormalizedXPaddingBytes() RangeConfig {
	if c.XPaddingBytes == nil || c.XPaddingBytes.To == 0 {
		return RangeConfig{
			From: 100,
			To:   1000,
		}
	}

	return *c.XPaddingBytes
}

func (c *SplitHttpConfig) GetNormalizedScMaxEachPostBytes() RangeConfig {
	if c.ScMaxEachPostBytes == nil || c.ScMaxEachPostBytes.To == 0 {
		return RangeConfig{
			From: 1000000,
			To:   1000000,
		}
	}

	return *c.ScMaxEachPostBytes
}

func (c *SplitHttpConfig) GetNormalizedScMinPostsIntervalMs() RangeConfig {
	if c.ScMinPostsIntervalMs == nil || c.ScMinPostsIntervalMs.To == 0 {
		return RangeConfig{
			From: 30,
			To:   30,
		}
	}

	return *c.ScMinPostsIntervalMs
}

func (c *SplitHttpConfig) GetNormalizedScMaxBufferedPosts() int {
	if c.ScMaxBufferedPosts == 0 {
		return 30
	}

	return int(c.ScMaxBufferedPosts)
}

func (c *SplitHttpConfig) GetNormalizedScStreamUpServerSecs() RangeConfig {
	if c.ScStreamUpServerSecs == nil || c.ScStreamUpServerSecs.To == 0 {
		return RangeConfig{
			From: 20,
			To:   80,
		}
	}

	return *c.ScMinPostsIntervalMs
}

func (m *XmuxConfig) GetNormalizedMaxConcurrency() RangeConfig {
	if m.MaxConcurrency == nil {
		return RangeConfig{
			From: 0,
			To:   0,
		}
	}

	return *m.MaxConcurrency
}

func (m *XmuxConfig) GetNormalizedMaxConnections() RangeConfig {
	if m.MaxConnections == nil {
		return RangeConfig{
			From: 0,
			To:   0,
		}
	}

	return *m.MaxConnections
}

func (m *XmuxConfig) GetNormalizedCMaxReuseTimes() RangeConfig {
	if m.CMaxReuseTimes == nil {
		return RangeConfig{
			From: 0,
			To:   0,
		}
	}

	return *m.CMaxReuseTimes
}

func (m *XmuxConfig) GetNormalizedHMaxRequestTimes() RangeConfig {
	if m.HMaxRequestTimes == nil {
		return RangeConfig{
			From: 0,
			To:   0,
		}
	}

	return *m.HMaxRequestTimes
}

func (m *XmuxConfig) GetNormalizedHMaxReusableSecs() RangeConfig {
	if m.HMaxReusableSecs == nil {
		return RangeConfig{
			From: 0,
			To:   0,
		}
	}

	return *m.HMaxReusableSecs
}

func (c RangeConfig) rand() int32 {
	return int32(crypto.RandBetween(int64(c.From), int64(c.To)))
}
