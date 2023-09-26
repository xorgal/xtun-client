package content

import (
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/xorgal/xtun-client/app/lib"
	"github.com/xorgal/xtun-client/internal"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/netutil"
)

func BuildSetupScreen(w fyne.Window) fyne.CanvasObject {
	separator := widget.NewSeparator()
	box := container.NewVBox(widget.NewForm(), separator, widget.NewForm())

	syncDeviceSettingsCheck := widget.NewCheck("", func(checked bool) {
		internal.AppState.SyncDeviceSettings = checked
		newForm := createSetupForm(w)
		(*box).Objects[2] = newForm
		box.Refresh()
	})
	syncDeviceSettingsCheck.SetChecked(internal.AppState.SyncDeviceSettings)

	skipTLSVerifyCheck := widget.NewCheck("", func(checked bool) {
		config.AppConfig.InsecureSkipVerify = checked
	})
	skipTLSVerifyCheck.SetChecked(config.AppConfig.InsecureSkipVerify)

	prefForm := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Auto-configure",
				Widget: syncDeviceSettingsCheck,
			},
			{
				Text:   "Skip TLS verify",
				Widget: skipTLSVerifyCheck,
			},
		},
	}

	(*box).Objects[0] = prefForm

	setupForm := createSetupForm(w)
	(*box).Objects[2] = setupForm

	return box
}

func createSetupForm(w fyne.Window) *widget.Form {
	serverAddrEntry := widget.NewEntry()
	serverAddrEntry.SetText(config.AppConfig.ServerAddr)

	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetText(config.AppConfig.Key)

	deviceNameEntry := widget.NewEntry()
	deviceNameEntry.SetText(config.AppConfig.DeviceName)

	bufferSizeEntry := lib.NewNumericalEntry()
	bufferSizeEntry.SetText(strconv.Itoa(config.AppConfig.BufferSize))

	mtuEntry := lib.NewNumericalEntry()
	mtuEntry.SetText(strconv.Itoa(config.AppConfig.MTU))

	compressEntry := widget.NewCheck("", nil)
	compressEntry.SetChecked(config.AppConfig.Compress)

	formItems := []*widget.FormItem{
		{
			Text:   "Server address",
			Widget: serverAddrEntry,
		},
		{
			Text:   "Key",
			Widget: keyEntry,
		},
	}

	if !internal.AppState.SyncDeviceSettings {
		// Initialize entries with config data

		formItems = append(formItems,
			&widget.FormItem{
				Text:   "Device name",
				Widget: deviceNameEntry,
			},
			&widget.FormItem{
				Text:   "Buffer Size",
				Widget: bufferSizeEntry,
			},
			&widget.FormItem{
				Text:   "MTU",
				Widget: mtuEntry,
			},
			&widget.FormItem{
				Text:   "Compress",
				Widget: compressEntry,
			},
		)
	}

	return &widget.Form{
		Items: formItems,
		OnSubmit: func() {
			config.AppConfig.ServerAddr = serverAddrEntry.Text
			config.AppConfig.Key = keyEntry.Text
			config.AppConfig.DeviceName = deviceNameEntry.Text

			if internal.AppState.SyncDeviceSettings {
				serverConfig, err := internal.GetServerConfiguration(config.AppConfig)
				if err != nil {
					lib.ShowErrorDialog(w, err)
					return
				} else {
					config.AppConfig.BufferSize = serverConfig.BufferSize
					config.AppConfig.MTU = serverConfig.MTU
					config.AppConfig.Compress = serverConfig.Compress
				}
			} else {
				config.AppConfig.BufferSize, _ = strconv.Atoi(bufferSizeEntry.Text)
				config.AppConfig.MTU, _ = strconv.Atoi(mtuEntry.Text)
				config.AppConfig.Compress = compressEntry.Checked
			}

			gateway, err := netutil.DiscoverGateway(true)
			if err != nil {
				lib.ShowErrorDialog(w, err)
			}
			config.AppConfig.LocalGateway = gateway.String()

			// DeviceId will be set in GetIP function
			req, res, err := internal.GetIP(config.AppConfig)
			if err != nil {
				lib.ShowErrorDialog(w, err)
				return
			} else {
				config.AppConfig.DeviceId = req.DeviceId
				config.AppConfig.CIDR = res.Client
				config.AppConfig.ServerIP = res.Server
			}

			internal.AppState.IsInitialized = true

			internal.SaveStateFile(internal.AppState)
			internal.SaveConfigFile(config.AppConfig)
			log.Println("New configuration saved")
			w.SetContent(BuildHomeScreen(w))
		},
		OnCancel: func() {
			log.Println("Configuration cancelled")
			// Todo: go home
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}
}
