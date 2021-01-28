package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fetcher/types"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	err  error
	tmpl *template.Template
)

const (
	SEP_LIST      = "|"
	SEP_KEY_VALUE = ":"
	SEP_ITEM      = ","
)

func Parse(subs []types.Subscription, templateFile string, startPort int, customs []types.CustomOutbound) (string, error) {
	initTemplate(templateFile)
	ret := make(map[string]string)
	// 处理机场订阅链接
	for _, sub := range subs {
		if !sub.Enable {
			continue
		}
		err = singleParse(sub.Name, sub.URL, sub.Rule, ret, false)
		if err != nil {
			logrus.Errorf("single parse error: %v", err)
		}
	}
	// 添加自己额外的配置(比如自建节点)
	for _, custom := range customs {
		if !custom.Enable {
			continue
		}
		logrus.Infof("* Read custom outbound: %v", custom.Ps)
		customOutboundString, err := ioutil.ReadFile(custom.Filename)
		if err != nil {
			logrus.Errorf("** Read custom outbound: %v error: %v, ignore it", custom.Ps, err)
			continue
		}
		newName := fmt.Sprintf("custom-%v", custom.Ps)
		logrus.Infof("** Read custom outbound: %v OK", newName)
		ret[newName] = string(customOutboundString)
	}

	return makeV2rayConfigFile(ret, startPort)
}

// single_parse 解析一个机场的订阅链接, 生成map，key为ps，value为单独的outbound(single=false)或者这个由outbound构成的完成的v2ray配置文件(inbounds+outbound)(single=true)
// @param name为了防止多个机场的ps重复，因此在每个ps前面加个这个机场的name
// @param nodesContent 由多行组成，每行是一个base64编码的单一配置信息(也就是后来的outbound的编码格式)
func singleParse(name string, subURL string, filterConfig string, ret map[string]string, single bool) error {
	logrus.Infof("> single parse %v...", name)

	rawString, err := download(subURL, nil)
	if err != nil {
		return err
	}
	nodesContent, err := Decode(rawString)
	if err != nil {
		return err
	}
	cfgs := parseFilterConfigs(filterConfig)

	reader := strings.NewReader(nodesContent)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		node := scanner.Text()
		if node == "" {
			continue
		}
		logrus.Debugf("node raw data: %v", node)
		ps, parsed, err := do_parse(node, single)
		if err != nil {
			// logrus.Errorf("parse node: %v error: %v,skip...", node, err)
			logrus.WithFields(logrus.Fields{"node": node, "error": err}).Errorf("parse node error")
		} else {
			newName := fmt.Sprintf("%v-%v", name, ps)
			logrus.Infof(">> 3 add parsed node: %v to result", newName)
			ret[newName] = parsed
		}
	}

	filter(ret, cfgs)
	logrus.Infof("single parse done.")
	return nil
}

const (
	MaxPortNo = 65535
)

var (
	ErrPortRange = errors.New("port range not enough")
)

type Multi struct {
	Ps             string
	InPort         string
	InboundString  string
	OutboundString string
}

func parseFilterConfigs(filterList string) []*types.FilterConfig {
	var cfgs []*types.FilterConfig
	if len(filterList) > 0 {
		// split by '|'
		lists := strings.Split(filterList, SEP_LIST)
		for _, list := range lists {
			if len(list) == 0 {
				continue
			}
			logrus.Debugf("list item: %v", list)
			// split by ':'
			items := strings.Split(list, SEP_KEY_VALUE)
			if len(items) != 2 {
				logrus.Fatalf("filter list format error")
			}
			cfg := &types.FilterConfig{}
			switch items[0] {
			case "b":
				cfg.Mode = types.ModeBlackList
			case "w":
				cfg.Mode = types.ModeWhiteList
			default:
				logrus.Fatalf("filter list format error: unknow filter type")
			}
			cfg.Lists = strings.Split(items[1], SEP_ITEM)
			cfgs = append(cfgs, cfg)
		}
	}
	return cfgs
}

func initTemplate(tmplFile string) {
	logrus.Infof("init template")
	var err error
	if tmpl != nil {
		return
	}
	tmpl, err = template.ParseFiles(tmplFile)
	if err != nil {
		logrus.Fatalf("parse template file error: %v", err)
	}
}

// makeV2rayConfigFile 生成所有所有节点都放在同一个配置文件中
// outbounds 中每个对应一个outbound
func makeV2rayConfigFile(outbounds map[string]string, startPort int) (string, error) {
	if len(outbounds) == 0 {
		return "", fmt.Errorf("No outbound supplied")
	}
	if MaxPortNo-startPort < len(outbounds) {
		logrus.WithFields(logrus.Fields{"startPort": startPort, "outbounds len": len(outbounds)}).Errorf(ErrPortRange.Error())
		return "", ErrPortRange
	}

	multiObjs := []Multi{}

	var b bytes.Buffer
	inPort := startPort
	// 根据outbound来生成同等数量的inbound，这些inboud从startPort开始，每次累加1
	// 并且没有被使用(listen),如果被使用了，则用下一个
	for ps, outbound := range outbounds {
		for portInUse(inPort) {
			inPort += 1
			if inPort > MaxPortNo {
				return "", ErrPortRange
			}
		}
		err = tmpl.ExecuteTemplate(&b, "inbound", map[string]string{"ps": ps, "port": fmt.Sprintf("%d", inPort)})
		if err != nil {
			return "", err
		}
		inboundString := b.String()
		ioutil.ReadAll(&b)
		// logrus.Debugf("inbound: %v", inboundString)

		multi := Multi{Ps: ps, InPort: fmt.Sprintf("%d", inPort), InboundString: inboundString, OutboundString: outbound}

		multiObjs = append(multiObjs, multi)
		inPort += 1
	}

	err = tmpl.ExecuteTemplate(&b, "multi-outbounds", multiObjs)
	if err != nil {
		return "", err
	}
	logrus.Infof("makeV2rayConfigFile done.")
	return b.String(), nil
}

func portInUse(port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// logrus.Debugf("port: %d not in use", port)
		return false
	}
	conn.Close()
	logrus.Debugf("port: %d in use", port)
	return true
}

const (
	PrefixVmess = "vmess://"
	PrefixSs    = "ss://"
)

func do_parse(node string, single bool) (string, string, error) {
	switch {
	case strings.HasPrefix(node, PrefixVmess):
		return parse_vmess(node, single)
	case strings.HasPrefix(node, PrefixSs):
		return parse_ss(node, single)
	default:
		return "", "", fmt.Errorf("Only support vmess ss")
	}
}

func filter(nodes map[string]string, cfgs []*types.FilterConfig) {
	for _, cfg := range cfgs {
		switch cfg.Mode {
		case types.ModeBlackList:
			for ps, _ := range nodes {
				for _, black := range cfg.Lists {
					if strings.Contains(ps, black) {
						logrus.Infof(">> filter: delete black list item: %s", ps)
						delete(nodes, ps)
					}
				}
			}
		case types.ModeWhiteList:
			for ps, _ := range nodes {
				exist := false
			INNER:
				for _, white := range cfg.Lists {
					if strings.Contains(ps, white) {
						exist = true
						break INNER
					}
				}
				if !exist {
					logrus.Infof(">> filter: delete non white list item: %s", ps)
					delete(nodes, ps)
				}
			}
		default:
		}
	}
}
