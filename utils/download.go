package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
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

	downloadTimeoutEnv := os.Getenv("DOWNLOAD_TIMEOUT")
	downloadTimeout, err := strconv.Atoi(downloadTimeoutEnv)
	if err != nil {
		logrus.Debugf("Use default_download_timeout: %v", default_download_timeout)
		downloadTimeout = default_download_timeout
	} else {
		logrus.Debugf("Use env DOWNLOAD_TIMEOUT: %v", downloadTimeout)
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
