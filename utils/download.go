package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// unit: ms
	default_download_timeout  = 5000
	max_download_content_size = 50
)

func download(subscriptionURL string, headers map[string][]string) (string, error) {
	logrus.Infof("Downloading url: %v ...", subscriptionURL)
	//TODO add proxy support

	// Note: just support GET
	req, err := http.NewRequest("GET", subscriptionURL, nil)
	if err != nil {
		return "", fmt.Errorf("New request error: %v", err)
	}
	for k, vs := range headers {
		for _, v := range vs {
			logrus.Debugf("Add http request header: %v->%v", k, v)
			req.Header.Add(k, v)
		}
	}

	downloadTimeout := viper.GetInt("download_timeout")
	if downloadTimeout == 0 {
		logrus.Infof("Use default_download_timeout: %v", default_download_timeout)
		downloadTimeout = default_download_timeout
	} else {
		logrus.Infof("Use config download_timeout: %v", downloadTimeout)
	}

	client := http.Client{Timeout: time.Millisecond * time.Duration(downloadTimeout)}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read response error: %v", err)
	}

	atMost := len(ret)
	if atMost > max_download_content_size {
		atMost = max_download_content_size
	}
	logrus.Infof("Downloaded content: %v...", string(ret)[:atMost])

	return string(ret), nil
}
