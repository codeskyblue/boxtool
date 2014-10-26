package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/bobappleyard/readline"
	"github.com/franela/goreq"
	"github.com/gobuild/goyaml"
)

var cfg struct {
	Uid    string `goyaml:"uid"`
	Driver string `goyaml:"driver"`
	Server string `goyaml:"server"`
}

var respInfo *RespInfo

func HttpCall(serv string) (*RespBasic, error) {
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://%s/api/%s", cfg.Server, strings.TrimPrefix(serv, "/")),
	}.Do()
	if err != nil {
		return nil, err
	}
	info := new(RespBasic)
	err = res.Body.FromJsonTo(info)
	return info, err
}

func HttpCall2(serv string, data interface{}) error {
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://%s/api/%s", cfg.Server, strings.TrimPrefix(serv, "/")),
	}.Do()
	if err != nil {
		return err
	}
	return res.Body.FromJsonTo(data)
}

func init() {
	//if runtime.GOOS != "windows" {
	//	log.Fatal("this program is designed for windows")
	//}

	// Load config file
	cfgdata, err := ioutil.ReadFile("_config.yml")
	if err != nil {
		log.Fatal(err)
	}
	if err = goyaml.Unmarshal(cfgdata, &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Check serverbox connection")
	// Check serverbox configuration
	info, err := HttpCall("/echo/abcdefg")
	if err != nil {
		log.Fatal(err)
	}
	if info.Message != "abcdefg" {
		log.Fatal("server api(/echo) check failed")
	}

	respInfo = new(RespInfo)
	err = HttpCall2("/info/"+cfg.Uid, respInfo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected\n")
}

func cmdInfo() {
	d := respInfo.Data
	fmt.Printf("Host: %s\n", d.Host)
	fmt.Printf("Author: %s\n", d.Author)
	fmt.Println("Proxy engine started...")
	for _, p := range d.Proxies {
		fmt.Printf("\tlocalhost:%d --> %s:%d\n", p.LocalPort, d.Host, p.RemotePort)
	}
}

func main() {
	cmdInfo()
	fmt.Printf("Welcome to serverbox console\n\ndriver:%s\n", cfg.Driver)
	prefix := fmt.Sprintf(">> [box] %s@%s(%s) $ ",
		respInfo.Data.Author, respInfo.Data.Host, respInfo.Data.Description)
	p, err := NewProxy(":2022", "localhost:32200")
	if err != nil {
		log.Fatal(err)
	}
	go p.ListenAndServe()
	for {
		l, err := readline.String(prefix)
		if err != nil {
			if err != io.EOF {
				log.Fatal("error: ", err)
			}
			break
		}
		switch l {
		case "h", "help":
			fmt.Println(genHelp("Usage:", map[string]string{
				"h,help":     "Show help information",
				"ns,netstat": "Show netstat info",
				"info":       "show basic infomation",
			}))
		case "ns", "netstat":
			fmt.Printf("%d bytes send, %d bytes received\n", p.sentBytes, p.receivedBytes)
		case "info":
			cmdInfo()
		default:
			fmt.Printf("- %s: command not found, type help for more information\n", l)
			continue
		}
		readline.AddHistory(l)
	}
	println("hello")
}
