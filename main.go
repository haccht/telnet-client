package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
	"github.com/LeeEirc/tclientlib"
)

var (
	Addr   string
	Port     string
)

func init() {
	flag.StringVar(&Addr, "addr", "127.0.0.1", "telnet address")
	flag.StringVar(&Port, "port", "23", "telnet port")
}

func main() {
	flag.Parse()
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer term.Restore(fd, state)

	w, h, _ := term.GetSize(fd)
	conf := tclientlib.Config{
		Timeout:  10 * time.Second,
		TTYOptions: &tclientlib.TerminalOptions{
			Wide: w,
			High: h,
		},
	}
	client, err := tclientlib.Dial("tcp", net.JoinHostPort(Addr, Port), &conf)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	sigChan := make(chan struct{}, 1)
	go func() {
		_, _ = io.Copy(os.Stdin, client)
		sigChan <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(client, os.Stdout)
		sigChan <- struct{}{}
	}()

	sigwinchCh := make(chan os.Signal, 1)
	signal.Notify(sigwinchCh, syscall.SIGWINCH)
	for {
		select {
		case <-sigChan:
			return

		case sigwinch := <-sigwinchCh:
			if sigwinch == nil {
				return
			}

			w, h, _ := term.GetSize(fd)
			err := client.WindowChange(w, h)
			if err != nil {
				log.Println("Unable to send window-change request.")
				continue
			}
		}
	}
}
