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
	flag.StringVar(&node, "node", "edge", "the node of this mapper")
	flag.StringVar(&env, "env", "linux", "the current env")
}

func main() {
	controller.Run(node, env)
}
