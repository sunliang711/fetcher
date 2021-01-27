package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type VmessNode struct {
	Host string
	Path string
	Tls  string
	Add  string
	Port string
	Aid  string
	Net  string
	Type string
	V    string
	Ps   string
	Id   string
	// Class int
}

var (
	noQuoteRe *regexp.Regexp
)

func init() {
	var err error
	// 有时候返回的字段的值不是字符串，这里匹配它，然后把它加上上引号构成字符串
	noQuoteRe, err = regexp.Compile(`(:\s*)(\d+)`)
	if err != nil {
		logrus.Fatalf("compile quote regexp error: %v", err)
	}
}

func parse_vmess(node string, full bool) (string, string, error) {
	logrus.Infof(">> parse vmess..")
	node = node[len(PrefixVmess):]

	decoded, err := Decode(node)
	if err != nil {
		logrus.Errorf("decode vmess node error: %v", err)
		return "", "", err
	}
	logrus.Debugf("decoded vmess node: %s", decoded)
	//TODO not tested
	// ${1} ${2}是正则的两个捕获
	decoded = noQuoteRe.ReplaceAllString(decoded, `${1}"${2}"`)

	// decoded data format:
	// {
	// "host":"",
	// "path":"",
	// "tls":"",
	// "add":"14.17.97.145",
	// "port":5010,
	// "aid":2,
	// "net":"tcp",
	// "type":"none",
	// "v":"2",
	// "ps":"台湾 01 [D3/VR/IPLC]",
	// "id":"32470e14-85fb-3bf0-aa0c-1f7ba46b58b7",
	// "class":3
	// }

	return convert_vmess(decoded, full)
}

// https://github.com/2dust/v2rayN/wiki/%E5%88%86%E4%BA%AB%E9%93%BE%E6%8E%A5%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E(ver-2)
// json数据如下
// {
// "v": "2",
// "ps": "备注别名",
// "add": "111.111.111.111",
// "port": "32000",
// "id": "1386f85e-657b-4d6e-9d56-78badb75e1fd",
// "aid": "100",
// "net": "tcp",
// "type": "none",
// "host": "www.bbb.com",
// "path": "/",
// "tls": "tls"
// }

// v:配置文件版本号,主要用来识别当前配置
// net ：传输协议（tcp\kcp\ws\h2\quic)
// type:伪装类型（none\http\srtp\utp\wechat-video） *tcp or kcp or QUIC
// host：伪装的域名
// 1)http host中间逗号(,)隔开
// 2)ws host
// 3)h2 host
// 4)QUIC securty
// path:path
// 1)ws path
// 2)h2 path
// 3)QUIC key/Kcp seed
// tls：底层传输安全（tls)

// @param full 表示把outbound生成到完整的单outbound的配置文件中
// 如果要生成多outbounds则用false
func convert_vmess(node string, full bool) (string, string, error) {
	logrus.Infof(">> convert vmess")

	var (
		v2Path string
		v2Host string
	)

	reader := strings.NewReader(node)

	var vnode VmessNode
	err = json.NewDecoder(reader).Decode(&vnode)
	if err != nil {
		logrus.Errorf("decode data to json object error: %v", err)
		return "", "", err
	}
	//TODO vnode.Ps vnode.Add vnode.Id 去掉空白字符
	if vnode.Tls != "tls" {
		vnode.Tls = "none"
	}

	// get v2Host v2Path
	if vnode.V == "2" {
		v2Host = vnode.Host
		v2Path = vnode.Path
	} else {
		switch vnode.Net {
		case "tcp":
			v2Host = vnode.Host
			v2Path = ""
		case "kcp":
			v2Host = ""
			v2Path = ""
		case "ws":
			v2HostTmp := vnode.Host
			if v2HostTmp != "" {
				slice := strings.Split(v2HostTmp, ";")
				if len(slice) > 0 {
					v2Host = slice[0]
					v2Path = slice[0]
				} else {
					v2Host = ""
					v2Path = v2Host
				}
			}
		case "h2":
			v2Host = ""
			v2Path = vnode.Path
		default:
			return "", "", fmt.Errorf("unknow net filed: %v", vnode.Net)
		}
	}

	vnode.Host = v2Host
	vnode.Path = v2Path

	var (
		tcp string = "null"
		kcp string = "null"
		ws  string = "null"
		h2  string = "null"
		tls string = "null"
	)

	if vnode.Tls == "tls" {
		tls = fmt.Sprintf(`{"allowInsecure":true,"serverName":"%v"}`, vnode.Host)
	} else {
		tls = "null"
	}

	if strings.Contains(vnode.Host, ",") {
		vnode.Host = strings.ReplaceAll(vnode.Host, ",", `","`)
	}

	w := bytes.Buffer{}
	switch vnode.Net {
	case "tcp":
		if vnode.Type == "http" {
			tmpl.ExecuteTemplate(&w, "tcp", vnode)
			tcp = w.String()
		} else {
			tcp = "null"
		}
	case "kcp":
		tmpl.ExecuteTemplate(&w, "kcp", vnode)
		kcp = w.String()
	case "ws":
		tmpl.ExecuteTemplate(&w, "ws", vnode)
		ws = w.String()
	case "h2":
		tmpl.ExecuteTemplate(&w, "h2", vnode)
		h2 = w.String()
	default:
		return "", "", fmt.Errorf("unknow net field: %v", vnode.Net)
	}
	ioutil.ReadAll(&w)

	m := map[string]string{
		"ps":           vnode.Ps,
		"address":      vnode.Add,
		"port":         vnode.Port,
		"id":           vnode.Id,
		"alterId":      vnode.Aid,
		"network":      vnode.Net,
		"security":     vnode.Tls,
		"tlsSettings":  tls,
		"tcpSettings":  tcp,
		"kcpSettings":  kcp,
		"wsSettings":   ws,
		"httpSettings": h2,
	}
	err = tmpl.ExecuteTemplate(&w, "outbound", m)
	if err != nil {
		return "", "", fmt.Errorf("ExecuteTemplate error: %v", err)
	}
	outbound := w.String()
	ioutil.ReadAll(&w)
	logrus.Debugf("vmess node result: %v", outbound)

	if full {
		tmpl.ExecuteTemplate(&w, "single-outbound", map[string]string{"outbound": outbound})
		outbound = w.String()
		ioutil.ReadAll(&w)
	}

	return vnode.Ps, outbound, nil
}
