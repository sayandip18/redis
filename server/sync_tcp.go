package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/sayandip/redis/config"
	"github.com/sayandip/redis/core"
)

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes, then repeated read until
	// we get EOF or designated delimiter
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}
	
	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.RedisCmd, c io.ReadWriter) {
	err := core.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
}

func RunSyncTCPServer() {
	log.Println("starting a synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		log.Println("err", err)
		return
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println("err", err)
		}
		con_clients++

		for {
			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				con_clients--
				if err == io.EOF {
					break
				}
			}
			respond(cmd, c)
		}

	}

}