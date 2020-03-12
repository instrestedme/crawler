package main

import (
	"crawler/global"
)

func init() {
	global.Init()
}

func main() {
	//爬虫服务
	go ServeCrawle()
	// 网页服务
	ServeWeb()
}
