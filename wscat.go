/**
 * wscat: WebSocket cat
 * Copytright (c) 2017 Atsushi Ezura <zura@zura.org>
 * This software is released under the MIT License, see LICENSE.md
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"net/url"
	"os"
)

type wscatConfig struct {
	Url          string
	Origin       string
	Conn         *websocket.Conn
	Reader       *os.File
	Writer       *os.File
	SendFilename string
	RecvFilename string
}

func (wscat *wscatConfig) run() {
	wsConn, err := websocket.Dial(wscat.Url, "tcp", wscat.Origin)
	if err != nil {
		pmessage := fmt.Sprintf("\x1b[31m[Fatal] cannot connect to server: %s\x1b[0m", wscat.Url)
		panic(pmessage)
	}
	wscat.Conn = wsConn

	// WebSocket packet sender.
	go func() {
		if wscat.Reader != os.Stdin {
			defer wscat.Reader.Close()
		}
		scanner := bufio.NewScanner(wscat.Reader)
		for scanner.Scan() {
			text := scanner.Text()
			websocket.Message.Send(wscat.Conn, text)
		}
		if err := scanner.Err(); err != nil {
			panic("\x1b[31m[Fatal] send error\x1b[0m")
		}
	}()

	// WebSocket packet receiver.
	if wscat.Writer != os.Stdout {
		defer wscat.Writer.Close()
	}
	for {
		var wsMessage string
		if err := websocket.Message.Receive(wscat.Conn, &wsMessage); err != nil {
			panic("\x1b[31m[Fatal] receive error\x1b[0m")
		}
		fmt.Fprintln(wscat.Writer, "\x1b[32m"+wsMessage+"\x1b[0m")
	}
}

func (wscat *wscatConfig) getOptions() {
	flag.StringVar(&wscat.Url, "c", "", "specified an WebSocket URL to connect to, ex) ws://echo.websocket.org/")
	flag.StringVar(&wscat.SendFilename, "i", "", "specified a filename which inputs sending data from.")
	flag.StringVar(&wscat.RecvFilename, "o", "", "specified a filename which outputs receiving data to.")

	flag.Parse()
}

func (wscat *wscatConfig) init() {
	// If it is no specified a WebSocket URL using -c option,
	// this assumes a string at the end of the command as a URL.
	if wscat.Url == "" {
		narg := flag.NArg()
		wscat.Url = flag.Arg(narg - 1)
	}
	if wscat.Url == "" {
		panic("\x1b[31minvalid URL\x1b[0m")
	}

	url, err := url.Parse(wscat.Url)
	if err != nil {
		panic("\x1b[31mURL parsing\x1b[0m")
	}
	wscat.Origin = "http://" + url.Host

	if wscat.SendFilename == "" {
		wscat.Reader = os.Stdin
	} else {
		reader, err := os.Open(wscat.SendFilename)
		if err != nil {
			panic(fmt.Sprintf("\x1b[31mCannot open %s\x1b[0m", wscat.SendFilename))
		}
		wscat.Reader = reader
	}

	if wscat.RecvFilename == "" {
		wscat.Writer = os.Stdout
	} else {
		writer, err := os.Create(wscat.RecvFilename)
		if err != nil {
			panic(fmt.Sprintf("\x1b[31mCannot open %s\x1b[0m", wscat.RecvFilename))
		}
		wscat.Writer = writer
	}
}

func main() {
	wscat := wscatConfig{}
	wscat.getOptions()
	wscat.init()
	wscat.run()
}
