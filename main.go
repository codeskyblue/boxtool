package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/codeskyblue/readline" // a fork version in order to support windows
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
	// for _, p := range d.Proxies {
	// 	fmt.Printf("\tlocalhost:%d --> %s:%d\n", p.LocalPort, d.Host, p.RemotePort)
	// }
}

func cmdQuit() {
	os.Exit(0)
}

func main() {
	cmdInfo()
	fmt.Printf("Welcome to serverbox console\n\ndriver:%s\n", cfg.Driver)
	prefix := fmt.Sprintf(">> [box] %s@%s(%s) $ ",
		respInfo.Data.Author, respInfo.Data.Host, respInfo.Data.Description)
	d := respInfo.Data

	fmt.Println("Proxy engine started...")
	proxies := make([]*Proxy, 0, len(d.Proxies))
	for _, p := range d.Proxies {
		px, err := NewProxy(fmt.Sprintf(":%d", p.LocalPort), fmt.Sprintf("%s:%d", d.Host, p.RemotePort))
		if err != nil {
			log.Fatal(err)
		}
		proxies = append(proxies, px)
		fmt.Printf("\tstart localhost:%d --> %s:%d\n", p.LocalPort, d.Host, p.RemotePort)
		go px.ListenAndServe()
	}

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
				"exit":       "exit program",
			}))
		case "ns", "netstat":
			println("netstat")
			for _, px := range proxies {
				fmt.Printf("[%v]: %d bytes send, %d bytes received\n", px.laddr, px.sentBytes, px.receivedBytes)
			}
		case "info":
			cmdInfo()
		case "exit", "quit":
			cmdQuit()
		default:
			fmt.Printf("- %s: command not found, type help for more information\n", l)
			continue
		}
		readline.AddHistory(l)
	}
}
