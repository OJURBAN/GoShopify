package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/corpix/uarand"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Session struct {
	Client    *http.Client
	Useragent string
	logger    *zap.Logger
}

func (s Session) UserAgent() string {
	return s.Useragent
}

func NewLogger() *zap.Logger {
	// info level enabler
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})

	// error and fatal level enabler
	errorFatalLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel || level == zapcore.FatalLevel
	})

	// write syncers
	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	// tee core
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			stdoutSyncer,
			infoLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			stderrSyncer,
			errorFatalLevel,
		),
	)

	// finally construct the logger with the tee core
	logger := zap.New(core)

	return logger
}

// dc.smartproxy.com:10001:user-OskarU-country-gb:xjChiur25bwLFEm0q3
// dc.smartproxy.com:10002:user-OskarU-country-gb:xjChiur25bwLFEm0q3
// dc.smartproxy.com:10003:user-OskarU-country-gb:xjChiur25bwLFEm0q3
// Constructor
func NewSession() *Session {
	sess := new(Session)
	sess.Useragent = uarand.GetRandom()
	sess.logger = NewLogger()

	// Setup transport
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	//// Setup proxy
	//proxy := http.ProxyURL(&url.URL{
	//	Scheme: "https",
	//	User:   url.UserPassword("user-OskarU-country-gb", "xjChiur25bwLFEm0q3"),
	//	Host:   "dc.smartproxy.com:10001",
	//})
	//t.Proxy = proxy

	// Setup cookiejar
	jar, err := cookiejar.New(nil)

	if err != nil {
		sess.logger.Error("Error creating cookiejar", zap.Error(err))
	}

	// Setup client
	sess.Client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
		Jar:       jar,
	}

	return sess
}

func (s Session) Get(url string, headers map[string][]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		//Handle Error
		s.logger.Error("Error creating request", zap.Error(err))
	}

	if len(headers) != 0 {
		fmt.Print("Headers set")
		req.Header = http.Header(headers)
	}

	req.Header.Set("User-Agent", s.Useragent)

	res, err := s.Client.Do(req)
	if err != nil {
		//Handle Error
		s.logger.Error("Error sending request", zap.Error(err))
	}

	return res, err
}

func (s Session) Post(url string, headers map[string][]string, body string) (resp *http.Response, err error) {
	parsedBody := strings.NewReader(body)
	req, err := http.NewRequest("POST", url, parsedBody)
	if err != nil {
		//Handle Error
		s.logger.Error("Error creating request", zap.Error(err))
	}
	if len(headers) != 0 {
		fmt.Print("Headers set")
		req.Header = http.Header(headers)
	}
	req.Header.Set("User-Agent", s.Useragent)
	res, err := s.Client.Do(req)
	if err != nil {
		//Handle Error
		s.logger.Error("Error sending request", zap.Error(err))
	}
	return res, err
}

func (s Session) PostJson(url string, headers map[string][]string, body map[string]interface{}) (resp *http.Response, err error) {
	req := new(http.Request)
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("could not marshal json: %s\n", err)
			return nil, err
		}
		req, err = http.NewRequest("POST", url, bytes.NewReader(jsonData))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}

	if err != nil {
		//Handle Error
		s.logger.Error("Error creating request", zap.Error(err))
	}

	if len(headers) != 0 {
		fmt.Print("Headers set")
		req.Header = http.Header(headers)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", s.Useragent)

	res, err := s.Client.Do(req)

	if err != nil {
		//Handle Error
		s.logger.Error("Error sending request", zap.Error(err))
	}

	return res, err
}

func (s Session) PostForm(url string, headers map[string][]string, form url.Values) (resp *http.Response, err error) {
	parsedBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", url, parsedBody)
	if err != nil {
		//Handle Error
		s.logger.Error("Error creating request", zap.Error(err))
	}
	if len(headers) != 0 {
		fmt.Print("Headers set")
		req.Header = http.Header(headers)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", s.Useragent)
	res, err := s.Client.Do(req)

	if err != nil {
		//Handle Error
		s.logger.Error("Error sending request", zap.Error(err))
	}
	return res, err
}
