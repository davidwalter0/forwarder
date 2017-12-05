// https://play.golang.org/p/H9rbjo7zg9
package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"encoding/json"
	"github.com/davidwalter0/go-cfg"
	"github.com/davidwalter0/transform"
	yaml "gopkg.in/yaml.v2"
)

type Cfg struct {
	Address string `json:"address" help:"host:port on which to serve" default:":8888"`
}

var c Cfg

type Handler struct{}

func init() {
	if err := cfg.Parse(&c); err != nil {
		fmt.Fprintf(os.Stderr, "Config failed %v", err)
		os.Exit(1)
	}

}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var buffer string
	if requestDump, err := httputil.DumpRequest(r, true); err != nil {
		fmt.Println(time.Now().Format(time.RFC3339), err)
	} else {
		buffer += "--------------------------------\n"
		buffer += time.Now().Format(time.RFC3339)
		buffer += "\n--------------------------------\n"
		buffer += "Server Host IP " + IP + "\n"
		buffer += string(requestDump) + "\n"
		buffer += string(r.RemoteAddr) + "\n"
		buffer += string(r.RequestURI) + "\n"
		if _, err := w.Write([]byte(buffer)); err != nil {
			fmt.Println(time.Now().Format(time.RFC3339), err)
		} else {
			fmt.Println(buffer)
		}
	}
}

var IP string

func main() {
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ip)
	IP = ip
	http.ListenAndServe(c.Address, Handler{})
}

// Yamlify object to yaml string
func Yamlify(data interface{}) string {
	data, err := transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

// Jsonify an object
func Jsonify(data interface{}) string {
	var err error
	data, err = transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	// s, err := json.MarshalIndent(data, "", "  ") // spaces)
	s, err := json.MarshalIndent(data, "", "  ") // spaces)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
