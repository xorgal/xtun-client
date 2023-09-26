package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xorgal/xtun-core/pkg/config"
)

type IDirPath struct {
	BinaryDir  string
	AppDataDir string
	TempDir    string
}

type IFilePath struct {
	BinaryPath string
	ConfigPath string
	StatePath  string
	PidPath    string
}

var BinaryFile = "xtun.exe"
var ConfigFile = "config.json"
var StateFile = "state.json"
var PidFile = ".xtun.pid"

var DirPath = IDirPath{
	BinaryDir:  ".",
	AppDataDir: ".",
	TempDir:    ".",
}

var FilePath = IFilePath{
	BinaryPath: fmt.Sprintf("%s/%s", DirPath.BinaryDir, BinaryFile),
	ConfigPath: fmt.Sprintf("%s/%s", DirPath.AppDataDir, ConfigFile),
	StatePath:  fmt.Sprintf("%s/%s", DirPath.AppDataDir, StateFile),
	PidPath:    fmt.Sprintf("%s/%s", DirPath.TempDir, PidFile),
}

func SaveConfigFile(config config.Config) error {
	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(FilePath.ConfigPath, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfigFile() error {
	file, err := os.ReadFile(FilePath.ConfigPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &config.AppConfig)
	if err != nil {
		return err
	}
	return nil
}

func IsConfigFileExists() bool {
	if _, err := os.Stat(FilePath.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Printf("error reading %s: %v", FilePath.ConfigPath, err)
			return false
		}
	} else {
		return true
	}
}

func SaveStateFile(state IAppState) error {
	file, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(FilePath.StatePath, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadStateFile() error {
	file, err := os.ReadFile(FilePath.StatePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &AppState)
	if err != nil {
		return err
	}
	return nil
}

func IsStateFileExists() bool {
	if _, err := os.Stat(FilePath.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Printf("error reading %s: %v", FilePath.StatePath, err)
			return false
		}
	} else {
		return true
	}
}

func SavePidFile() error {
	_, err := os.OpenFile(FilePath.PidPath, os.O_CREATE|os.O_EXCL, 0666)
	return err
}

func RmPidFile() error {
	err := os.Remove(FilePath.PidPath)
	return err
}
