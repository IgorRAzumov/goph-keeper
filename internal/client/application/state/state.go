package state

import (
	"goph-keeper/internal/client/api"
	clientcfg "goph-keeper/internal/client/config"
	"goph-keeper/internal/client/store"
)

// State — сессия клиента: конфиг, локальные данные и API-клиент.
type State struct {
	ConfigPath string
	DataPath   string
	MasterPass string
	API        *api.Client
	Config     clientcfg.File
	Data       *store.Data
}

// Load открывает конфиг и локальное хранилище.
func Load(masterPassword string) (*State, error) {
	config, err := clientcfg.Load()
	if err != nil {
		return nil, err
	}
	dataPath, err := clientcfg.DataPath()
	if err != nil {
		return nil, err
	}
	data, err := store.Open(dataPath)
	if err != nil {
		return nil, err
	}
	cfgPath, _ := clientcfg.Path()
	return &State{
		ConfigPath: cfgPath,
		DataPath:   dataPath,
		MasterPass: masterPassword,
		Config:     config,
		Data:       data,
		API:        api.NewClient(config.ServerURL),
	}, nil
}

// Save сохраняет конфиг и data.json.
func (state *State) Save() error {
	if err := clientcfg.Save(state.Config); err != nil {
		return err
	}
	return state.Data.Save(state.DataPath)
}

// SetServerURL обновляет URL сервера и пересоздаёт API-клиент.
func (state *State) SetServerURL(serverURL string) {
	if serverURL == "" {
		return
	}
	state.Config.ServerURL = serverURL
	state.API = api.NewClient(serverURL)
}
