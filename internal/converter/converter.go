package converter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func NodeToShareLink(nodeYaml []byte) (string, error) {
	var node map[string]any
	if err := yaml.Unmarshal(nodeYaml, &node); err != nil {
		return "", err
	}

	nodeType, ok := node["type"].(string)
	if !ok {
		return "", fmt.Errorf("unknown node type")
	}

	switch nodeType {
	case "vmess":
		return vmessToLink(node)
	case "ss":
		return ssToLink(node)
	case "trojan":
		return trojanToLink(node)
	case "vless":
		return vlessToLink(node)
	case "ssr":
		return ssrToLink(node)
	case "socks5":
		return socks5ToLink(node)
	case "http":
		return httpToLink(node)
	default:
		return "", fmt.Errorf("unsupported node type: %s", nodeType)
	}
}

func vmessToLink(node map[string]any) (string, error) {
	vmessConfig := map[string]any{
		"v":    "2",
		"ps":   getString(node, "name"),
		"add":  getString(node, "server"),
		"port": getPort(node),
		"id":   getString(node, "uuid"),
		"aid":  getInt(node, "alterId"),
		"net":  getString(node, "network"),
		"type": getString(node, "cipher"),
		"host": getString(node, "servername"),
		"path": getString(node, "ws-path"),
		"tls":  "",
	}

	if getBool(node, "tls") {
		vmessConfig["tls"] = "tls"
	}

	jsonData, err := json.Marshal(vmessConfig)
	if err != nil {
		return "", err
	}

	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData), nil
}

func ssToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	password := getString(node, "password")
	cipher := getString(node, "cipher")
	name := getString(node, "name")

	userInfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cipher, password)))
	link := fmt.Sprintf("ss://%s@%s:%s", userInfo, server, port)
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func trojanToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	password := getString(node, "password")
	name := getString(node, "name")

	link := fmt.Sprintf("trojan://%s@%s:%s", url.QueryEscape(password), server, port)
	
	params := url.Values{}
	if sni := getString(node, "sni"); sni != "" {
		params.Add("sni", sni)
	}
	if alpn := getString(node, "alpn"); alpn != "" {
		params.Add("alpn", alpn)
	}
	
	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func vlessToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	uuid := getString(node, "uuid")
	name := getString(node, "name")

	link := fmt.Sprintf("vless://%s@%s:%s", uuid, server, port)
	
	params := url.Values{}
	if getBool(node, "tls") {
		params.Add("security", "tls")
	}
	if sni := getString(node, "servername"); sni != "" {
		params.Add("sni", sni)
	}
	if typeStr := getString(node, "type"); typeStr != "" {
		params.Add("type", typeStr)
	}
	
	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func ssrToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	password := getString(node, "password")
	cipher := getString(node, "cipher")
	protocol := getString(node, "protocol")
	obfs := getString(node, "obfs")
	name := getString(node, "name")

	part1 := fmt.Sprintf("%s:%s:%s:%s:%s:%s", 
		server, port, protocol, cipher, obfs, 
		base64.StdEncoding.EncodeToString([]byte(password)))
	
	params := url.Values{}
	if name != "" {
		params.Add("remarks", base64.StdEncoding.EncodeToString([]byte(name)))
	}
	
	link := "ssr://" + base64.URLEncoding.EncodeToString([]byte(part1))
	if len(params) > 0 {
		link += "/?" + params.Encode()
	}
	return link, nil
}

func socks5ToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	username := getString(node, "username")
	password := getString(node, "password")
	name := getString(node, "name")

	var link string
	if username != "" && password != "" {
		link = fmt.Sprintf("socks5://%s:%s@%s:%s", 
			url.QueryEscape(username), url.QueryEscape(password), server, port)
	} else {
		link = fmt.Sprintf("socks5://%s:%s", server, port)
	}
	
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func httpToLink(node map[string]any) (string, error) {
	server := getString(node, "server")
	port := getPort(node)
	username := getString(node, "username")
	password := getString(node, "password")
	name := getString(node, "name")

	var link string
	if username != "" && password != "" {
		link = fmt.Sprintf("http://%s:%s@%s:%s", 
			url.QueryEscape(username), url.QueryEscape(password), server, port)
	} else {
		link = fmt.Sprintf("http://%s:%s", server, port)
	}
	
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getPort(m map[string]any) string {
	if v, ok := m["port"]; ok {
		switch val := v.(type) {
		case int:
			return strconv.Itoa(val)
		case int64:
			return strconv.FormatInt(val, 10)
		case float64:
			return strconv.Itoa(int(val))
		case string:
			return val
		}
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func NodesToShareLinks(nodes [][]byte) ([]string, error) {
	var links []string
	for _, node := range nodes {
		link, err := NodeToShareLink(node)
		if err == nil {
			links = append(links, link)
		}
	}
	return links, nil
}

func GenerateShareLinksFile(links []string) []byte {
	return []byte(strings.Join(links, "\n"))
}
