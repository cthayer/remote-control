package client

import (
	"github.com/cthayer/remote_control/internal/logger"
)

func init() {
	// initialize the logger
	_, err := logger.InitLogger("info")

	if err != nil {
		println(err.Error())
	}
}
