// Copyright © 2016 Tuenti Technologies S.L.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var version = "dev"

func main() {
	var haproxyPath, haproxyPIDFile, haproxyConfigFile, controlSocketFile string
	var syslogPort int
	var showVersion bool
	flag.IntVar(&syslogPort, "syslog-port", 514, "Port for embedded syslog server")
	flag.StringVar(&haproxyPath, "haproxy", "/usr/local/sbin/haproxy", "Path to haproxy binary")
	flag.StringVar(&haproxyPIDFile, "haproxy-pidfile", "/var/run/haproxy.pid", "Pidfile for haproxy")
	flag.StringVar(&controlSocketFile, "control-socket", "unix:/var/run/haproxyctl.sock", "Socket file for control commands")
	flag.StringVar(&haproxyConfigFile, "haproxy-config", "/usr/local/etc/haproxy/haproxy.cfg", "Path to configuration file for haproxy")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	syslog := NewSyslogServer(syslogPort)
	if err := syslog.Start(); err != nil {
		log.Fatalf("Couldn't start embedded syslog: %v\n", err)
	}
	defer syslog.Stop()

	haproxy := NewHaproxyServer(haproxyPath, haproxyPIDFile, haproxyConfigFile)
	if err := haproxy.Start(); err != nil {
		log.Fatalf("Couldn't start haproxy: %v\n", err)
	}
	defer haproxy.Stop()

	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	controller := NewController(controlSocketFile, haproxy)

	go func() {
		for {
			log.Printf("Signal received: %v\n", <-done)
			if err := controller.Stop(); err != nil {
				log.Fatalf("Couldn't cleanly stop controller: %v", err)
			}
		}
	}()

	if err := controller.Run(); err != nil {
		log.Fatalf("Controller failed: %v\n", err)
	}
}
