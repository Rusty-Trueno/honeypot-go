package main

import (
	"flag"
	"honeypot/controller"
)

var (
	node = ""
	env  = ""
)

func init() {
	flag.StringVar(&node, "node", "hqq", "the node of this mapper")
	flag.StringVar(&env, "env", "windows", "the current env")
}

func main() {
	controller.Run(node, env)
}
