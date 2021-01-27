package utils

import "github.com/sirupsen/logrus"

func parse_ss(node string, full bool) (string, string, error) {
	node = node[len(PrefixSs):]
	logrus.WithField("node", node).Infof("parse_ss")
	return "", "", nil
}
