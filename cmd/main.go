package main

import (
	"log"
	"os"

	"github.com/xorgal/xtun-client/app"
	"github.com/xorgal/xtun-client/app/lib"
	"github.com/xorgal/xtun-client/internal"
	"github.com/xorgal/xtun-core/pkg/config"
)

// ===============================
// To be obtained at later stage:
// ===============================
// DeviceId           string
// ServerAddr         string
// ServerIP           string
// LocalGateway       string
// CIDR               string
// ===============================

func init() {
	os.Setenv("FYNE_THEME", "light")

	fileExists := internal.IsStateFileExists()
	if !fileExists {
		internal.SaveStateFile(internal.DefaultState)
	} else {
		internal.LoadStateFile()
	}

	fileExists = internal.IsConfigFileExists()
	if !fileExists {
		config.AppConfig.DeviceName = "xtun"
		config.AppConfig.Key = "xtun@2023"
		config.AppConfig.BufferSize = 65536
		config.AppConfig.MTU = 1500
		config.AppConfig.InsecureSkipVerify = false
		config.AppConfig.Compress = false
		config.AppConfig.GlobalMode = true
		config.AppConfig.ServerMode = false
		config.AppConfig.GUIMode = true
		config.AppConfig.Protocol = "wss"

		internal.SaveConfigFile(config.AppConfig)
	} else {
		internal.LoadConfigFile()
	}

	lib.InitAppLogger()
}

func main() {
	err := internal.SavePidFile()
	if err != nil {
		if os.IsExist(err) {
			log.Fatalf("Another instance of the app is already running.")
		}
		log.Fatalf("Unable to start the app: %v", err)
	}
	defer internal.RmPidFile()

	app.RunLoop()
}
