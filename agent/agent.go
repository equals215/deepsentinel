package agent

import (
	ipc "github.com/james-barrow/golang-ipc"
	log "github.com/sirupsen/logrus"
)

func work() {
}

// ipcHandler is a function that handles incoming IPC messages
// messages are formed by the cli and sent to the agent
// format is like so "command:arg1,arg2,arg3"
func ipcHandler(server *ipc.Server) {
	for {
		message, err := server.Read()

		if err != nil {
			log.Errorf("failed to read message: %s", err.Error())
		}

		log.Infof("received message: %s", message.Data)

	}

}
