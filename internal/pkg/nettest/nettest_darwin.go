//go:build darwin

package nettest

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-ping/ping"
	"github.com/ptonlix/spokenai/configs"
	"go.uber.org/zap"
)

const (
	sendcount = 3
	timeout   = time.Second * 5
)

type TestResult struct {
	host   string
	result bool
}
type TestClient struct {
	hostAddr []string
	logger   *zap.Logger
}

func NewTestClient(logger *zap.Logger, addr []string) *TestClient {
	return &TestClient{logger: logger, hostAddr: addr}
}

func (t *TestClient) LoopTest() (*TestResult, error) {
	for _, host := range t.hostAddr {
		pinger, err := ping.NewPinger(host)
		if err != nil {
			t.logger.Error("Network Test Error:", zap.String("error", fmt.Sprintf("%+v", err)))
			return nil, err
		}
		pinger.Count = sendcount
		pinger.Timeout = timeout
		pinger.SetPrivileged(false)
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			t.logger.Error("Network Test Error:", zap.String("error", fmt.Sprintf("%+v", err)))
			return nil, err
		}
		if pinger.Statistics().PacketsRecv < sendcount-1 {
			t.logger.Error("Network Test Error:", zap.String("error", fmt.Sprintf("sendcount:%d recvcount%d", sendcount, pinger.Statistics().PacketsRecv)))
			return &TestResult{host: host, result: false}, nil
		}
	}
	return &TestResult{host: "", result: true}, nil
}

func GetResult(logger *zap.Logger) bool {
	u, err := url.Parse(configs.Get().OpenAi.Base.ApiHost)
	if err != nil {
		return false
	}
	testlist := []string{
		u.Host,
		"www.oyster-iot.cloud",
	}
	client := NewTestClient(logger, testlist)
	result, err := client.LoopTest()

	if err != nil {
		return false
	}
	return result.result
}
