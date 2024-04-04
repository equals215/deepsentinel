package agent

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
)

var socketAddress = "/tmp/deepsentinel.sock"
var stop = struct {
	val bool
	sync.Mutex
}{}

func startSocketServer() (*net.UnixListener, error) {
	os.Remove(socketAddress)

	l, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: socketAddress,
		Net:  "unix",
	})
	if err != nil {
		return nil, err
	}

	log.Debugf("Listening on %s ...", socketAddress)
	return l, nil
}

// socketIPCHandler is a function that handles incoming IPC messages
// messages are formed by the cli and sent to the agent
// format is like so "command:arg1,arg2,arg3"
func socketIPCHandler(sock *net.UnixListener) {
	for {
		stop.Lock()
		if stop.val {
			stop.Unlock()
			return
		}
		stop.Unlock()
		sock.SetDeadline(time.Now().Add(1 * time.Millisecond))
		conn, err := sock.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				// Accept timeout, continue to check for stop signal
				continue
			}
			fmt.Println("Error accepting:", err.Error())
			continue
		}

		log.Debug("New client connected.")
		handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	recvMessage := string(buf[:n])
	log.Debug("Received message:", recvMessage)
	if recvMessage == "ping" {
		log.Trace("Answering with pong")
		conn.Write([]byte("pong"))
	} else if recvMessage == "stop" {
		log.Info("Gracefully stopping agent...")
		stop.Lock()
		stop.val = true
		stop.Unlock()
		conn.Write([]byte("ok"))
	} else {
		resp, err := processRequest(recvMessage)
		if err != nil {
			log.Errorf("Error processing request: %s", err.Error())
			return
		}
		config.RefreshClientConfig()
		refresh := true
		config.PrintClientConfig(refresh)
		conn.Write([]byte(resp))
	}
}

func processRequest(message string) (string, error) {
	// split the message into command and arguments
	// command is before the first colon
	// arguments are comma separated
	//
	// example: "server-address:https://example.com"
	// command: "server-address"
	// arguments: "https://example.com"
	parts := strings.Split(message, "=")
	if handler, ok := instructionMap[parts[0]]; ok {
		log.Trace("Processing instruction:", parts[0])
		args := strings.Split(parts[1], ",")
		argInterfaces := make([]interface{}, len(args))
		for i, arg := range args {
			argInterfaces[i] = arg
		}
		err := handler(argInterfaces...)
		if err != nil {
			return "", err
		}
		return "ok", nil
	}
	return "", fmt.Errorf("unknown instruction: %s", parts[0])
}

func testIPCSocket() error {
	log.Trace("Sending ping to daemon")
	message := "ping"
	resp, err := sendMessageToDaemon(message)
	if err != nil {
		return err
	}
	if resp != "pong" {
		return fmt.Errorf("unexpected response: %s", resp)
	}

	log.Trace("Daemon is alive!")
	return nil
}

func sendMessageToDaemon(message string) (string, error) {
	conn, err := net.Dial("unix", socketAddress)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	log.Trace("Connected to daemon")

	log.Tracef("Sending message: %s", message)
	_, err = conn.Write([]byte(message))
	if err != nil {
		return "", err
	}
	log.Trace("Message sent")

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	log.Tracef("Received response: %s", string(buf[0:n]))

	return string(buf[0:n]), nil
}
