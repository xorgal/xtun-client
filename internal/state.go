package internal

type IAppState struct {
	SyncDeviceSettings bool
	IsInitialized      bool
}

var DefaultState = IAppState{
	SyncDeviceSettings: true,
	IsInitialized:      false,
}

var AppState IAppState
