package main

type RespBasic struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type RespInfo struct {
	RespBasic
	Data struct {
		Author      string `json:"author"`
		Host        string `json:"host"`
		Description string `goyaml:"description"`
		Proxies     []struct {
			LocalPort  int `json:"lport"`
			RemotePort int `json:"rport"`
		} `json:"proxies"`
		Ftp struct {
			Path string `json:"path"`
			Port int    `json:"port"`
		} `json:"ftp"`
	} `json:"data"`
}
