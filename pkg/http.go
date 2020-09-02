package pkg

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

func GetOBSFile(url *string, timeout *int) error {
	glog.Info("start to download...")
	client := TestHTTPClient(timeout)
	request, err := http.NewRequest(http.MethodGet, *url, nil)
	if err != nil {
		glog.Error(err)
		return err
	}

	var reader io.ReadCloser
	getIt := func() (bool, error) {
		resp, err2 := client.Do(request)
		if err2 != nil {
			glog.Errorf("%v, will retry", err2)
			return false, err2
		}
		reader = resp.Body
		return true, nil
	}

	retryErr := retry(6, getIt, 5*time.Second)
	if retryErr != nil {
		glog.Errorf("retry final error: %v", retryErr)
	}

	out, err2 := os.Create("download.file")
	if err2 != nil {
		glog.Error(err2)
		return err2
	}
	defer out.Close()
	_, err3 := io.Copy(out, reader)
	if err3 != nil {
		glog.Error(err3)
		return err3
	}

	glog.Info("finish download!")
	return nil
}

func retry(retryCount int, fn func() (bool, error), interval time.Duration) error {
	var (
		err error
		ok  bool
	)
	for count := 0; count < retryCount; count++ {
		ok, err = fn()
		if ok {
			return err
		}
		time.Sleep(interval)
		glog.Warningf("retry count: %d", count+1)
	}

	return err
}

func TestHTTPClient(timeout *int) http.Client {
	t := int32(*timeout)
	return http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   time.Duration(t) * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			MaxIdleConnsPerHost:   200,
			ResponseHeaderTimeout: 120 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true},
		},
	}
}
