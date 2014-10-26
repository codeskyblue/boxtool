package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/bobappleyard/readline"
	"github.com/gobuild/goyaml"
)

var cfg struct {
	Driver string `goyaml:"driver"`
}

type info struct {
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
}

func main() {
	fmt.Printf("Welcome to serverbox console\n\ndriver:%s\n", cfg.Driver)
	p, err := NewProxy(":2022", "localhost:32200")
	if err != nil {
		log.Fatal(err)
	}
	go p.ListenAndServe()
	for {
		l, err := readline.String("[box]hzsunshx@neifu$>> ")
		if err != nil {
			if err != io.EOF {
				log.Fatal("error: ", err)
			}
			break
		}
		switch l {
		case "h", "help":
			fmt.Println("Help me!!")
		case "ns", "netstat":
			fmt.Println(p.sentBytes, p.receivedBytes)
		default:
			fmt.Printf("- %s: command not found, type help for more information\n", l)
		}
		readline.AddHistory(l)
	}
	println("hello")
}
