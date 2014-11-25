package main

import (
	"log"

	"github.com/ian-kent/gofigure"
)

/* go build -o example .
 * ./example -h
 * ./example -local-addr="test"
 * ./example -remote-addr="test"
 * ./example -local-addr="test" -remote-addr="test"
 * BAR_REMOTE_ADDR="test" ./example
 * BAR_LOCAL_ADDR="test" ./example -remote-addr="test"
 * BAR_LOCAL_ADDR="test" ./example -local-addr="override"
 */

// Define a struct
type config struct {
	// Add a gofigure field to set envPrefix and order
	gofigure interface{} `envPrefix:"BAR" order:"cmd,env"`
	// Define some configuration items
	RemoteAddr string `env:"REMOTE_ADDR" cmd:"remote-addr"`
	LocalAddr  string `env:"LOCAL_ADDR" cmd:"local-addr"`
}

func main() {
	var cfg config
	// Pass a reference to Gofigure
	err := gofigure.Gofigure(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Fields on cfg should be set!
	log.Printf("%+v", cfg)
}
