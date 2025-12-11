package speedtest

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/showwin/speedtest-go/speedtest"
)

type Result struct {
	ServerName    string
	ServerCountry string
	Latency       time.Duration
	Download      float64
	Upload        float64
}

// "socks://127.0.0.1:7890"
// speedtest.WithUserConfig(&speedtest.UserConfig{Proxy: proxy})
func Run(ctx context.Context, options ...speedtest.Option) (*Result, error) {
	var speedtestClient = speedtest.New()

	for _, option := range options {
		option(speedtestClient)
	}

	serverList, err := speedtestClient.FetchServers()
	if err != nil {
		return nil, err
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		return nil, err
	}
	s := targets[0]
	// Please make sure your host can access this test server,
	// otherwise you will get an error.
	// It is recommended to replace a server at this time
	s.PingTestContext(ctx, nil)
	s.DownloadTestContext(ctx)
	s.UploadTestContext(ctx)
	// Note: The unit of s.DLSpeed, s.ULSpeed is bytes per second, this is a float64.
	fmt.Printf("Latency: %s, Download: %s, Upload: %s\n", s.Latency, s.DLSpeed, s.ULSpeed)
	log.Info().Dur("latenry", s.Latency).Float64("download", float64(s.DLSpeed)).Float64("upload", float64(s.ULSpeed)).
		Str("name", s.Name).Str("country", s.Country).Msg("speedtest result")
	s.Context.Reset() // reset counter
	return &Result{Latency: s.Latency,
		Download:      float64(s.DLSpeed),
		Upload:        float64(s.ULSpeed),
		ServerName:    s.Name,
		ServerCountry: s.Country}, nil
}
