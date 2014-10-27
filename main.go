package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/log"

	"github.com/codeskyblue/go-sh"
	"github.com/codeskyblue/readline" // a fork version in order to support windows
	"github.com/franela/goreq"
	"github.com/gobuild/goyaml"
	"github.com/kballard/go-shellquote"
)

var cfg struct {
	Uid    string `goyaml:"uid"`
	Driver string `goyaml:"driver"`
	Server string `goyaml:"server"`
}

var (
	respInfo *RespInfo
	// FTPUSE      = "ftpuse"
	FTPUSE      = filepath.Join(SelfDir(), "ftpuse", "ftpuse.exe")
	httpTimeout = 3000 * time.Millisecond
)

func HttpCall(serv string) (*RespBasic, error) {
	res, err := goreq.Request{
		Uri:     fmt.Sprintf("http://%s/api/%s", cfg.Server, strings.TrimPrefix(serv, "/")),
		Timeout: httpTimeout,
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
		Uri:     fmt.Sprintf("http://%s/api/%s", cfg.Server, strings.TrimPrefix(serv, "/")),
		Timeout: httpTimeout,
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
	if respInfo.Status != 0 {
		log.Fatal(respInfo.Message)
	}
	fmt.Println("connected\n")

	// ftpusePath, err := exec.LookPath(FTPUSE)
	// if err != nil {
	// 	fmt.Println("ftpuse not found in %PATH%, start check C:\\..")
	// 	ftpusePath = filepath.Join(`C:\Program Files\FERRO Software\FtpUse`, "ftpuse")
	// 	if Exists(ftpusePath) {
	// 		fmt.Println("Use C:\\...ftpuse")
	// 		FTPUSE = ftpusePath
	// 	} else {
	// 		fmt.Println("ftpuse not found.")
	// 	}
	// } else {
	// 	FTPUSE = ftpusePath
	// }
}

func cmdInfo() {
	d := respInfo.Data
	fmt.Printf("Author: %s\n", d.Author)
	fmt.Printf("Description: %s\n", d.Description)
	fmt.Printf("Host: %s\n", d.Host)
	fmt.Printf("Ftp: ftp://%s:%d/%s\n", d.Host, d.Ftp.Port, d.Ftp.Path)
}

func cmdMount(args ...string) {
	d := respInfo.Data

	fmt.Printf("\nMount ftp to %s\n", cfg.Driver)
	err := sh.Command(FTPUSE, cfg.Driver,
		d.Host+"/"+d.Ftp.Path,
		fmt.Sprintf("/PORT:%d", d.Ftp.Port)).Run()
	if err != nil {
		log.Printf("Mount failed to %s: %s", cfg.Driver, err)
	}
}

func cmdUnmount(args ...string) {
	sh.Command(FTPUSE, cfg.Driver, "/DELETE").Run()
}

func cmdServCtrl(action string) {
	ri, err := HttpCall("/codectl/" + action + "/" + cfg.Uid)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("Status: %d\n", ri.Status)
	fmt.Println(ri.Message)
}

func cmdCodeCtrl(action string) {
	ri, err := HttpCall("/codectl/" + action + "/" + cfg.Uid)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("Status: %d\n", ri.Status)
	fmt.Println(ri.Message)
}

func main() {
	cmdInfo()
	prefix := fmt.Sprintf(">> [box] %s@%s(%s) $ ",
		respInfo.Data.Author, respInfo.Data.Host, respInfo.Data.Description)
	prefix = ">>> "
	d := respInfo.Data

	fmt.Println("Proxy engine started...")
	proxies := make([]*Proxy, 0, len(d.Proxies))
	for _, p := range d.Proxies {
		px, err := NewProxy(fmt.Sprintf("localhost:%d", p.LocalPort), fmt.Sprintf("%s:%d", d.Host, p.RemotePort))
		if err != nil {
			log.Fatal(err)
		}
		proxies = append(proxies, px)
		fmt.Printf("\t%v --> %v\n", px.laddr, px.raddr)
		go px.ListenAndServe()
	}

	fmt.Printf("\nWelcome to serverbox console\n\n")
	for {
		l, err := readline.String(prefix)
		if err != nil {
			if err != io.EOF {
				log.Fatal("error: ", err)
			}
			break
		}

		args, err := shellquote.Split(l)
		if err != nil {
			log.Println("shellquote", err)
			continue
		}
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "h", "help":
			fmt.Println(genHelp("Usage:", map[string]string{
				"help,h":         "Show help information",
				"netstat,ns":     "Show netstat info",
				"info,i":         "Show basic infomation",
				"quit,exit":      "Exit program",
				"mount,m":        "Mount ftp to local driver",
				"unmount":        "Unmount ftp driver",
				"reload,restart": "Service control",
				"update,up":      "Update code(svn)",
				"revert,rv":      "Revert code(svn)",
				"status,st":      "Status code(svn)",
			}))
		case "ns", "netstat":
			for _, px := range proxies {
				fmt.Printf("[%v]: %d bytes send, %d bytes received\n", px.laddr, px.sentBytes, px.receivedBytes)
			}
		case "m", "mount":
			cmdMount(args[1:]...)
		case "unmount":
			cmdUnmount(args[1:]...)
		case "i", "info":
			cmdInfo()
		case "reload", "restart":
			cmdServCtrl(args[0])
		case "status", "st", "update", "up", "revert", "rv":
			cmdCodeCtrl(args[0])
		case "exit", "quit":
			os.Exit(0)
		default:
			fmt.Printf("- %s: command not found, type help for more information\n", l)
			continue
		}
		readline.AddHistory(l)
	}
}
