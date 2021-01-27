package utils

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func Decode(content string) (string, error) {
	// get length of content
	lenOfContent := len(content)
	logrus.Debugf("Original decode content: %v", content)
	logrus.Debugf("Len of content: %v", lenOfContent)

	// padding with "=" if lenOfContent %4 !=0
	if lenOfContent%4 != 0 {
		padstring := string("===="[lenOfContent%4:])

		content = fmt.Sprintf("%v%v", content, padstring)
		logrus.Debugf("Padding with: %v", padstring)
	}
	// why?
	// replace '-' with '+'
	// replace '_' with '/'
	content = strings.ReplaceAll(content, "-", "+")
	content = strings.ReplaceAll(content, "_", "/")
	logrus.Debugf("Decoded decode content: %v", content)

	ret, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", fmt.Errorf("Decode error: %v", err)
	}

	return string(ret), nil
}
