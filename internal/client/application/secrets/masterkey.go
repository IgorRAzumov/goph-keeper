package secrets

import (
	"errors"
	"os"
	"strings"

	"goph-keeper/internal/client/application/state"
	clientcrypto "goph-keeper/internal/client/crypto"
)

// MasterKey выводит ключ шифрования из master password и salt в конфиге.
func MasterKey(state *state.State) ([]byte, error) {
	pass := state.MasterPass
	if pass == "" {
		pass = strings.TrimSpace(os.Getenv("GOPHKEEPER_MASTER_PASSWORD"))
	}
	if pass == "" {
		return nil, errors.New("master password required (GOPHKEEPER_MASTER_PASSWORD)")
	}
	if state.Config.MasterSalt == "" {
		return nil, errors.New("master salt missing: login first")
	}
	return clientcrypto.DeriveKey(pass, state.Config.MasterSalt)
}
