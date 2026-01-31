package main

import "flag"

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.toml", "path to config TOML")
	flag.Parse()

}
