package pgp

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

func GenerateKeyPair(name, email, passphrase string) (KeyPair, error) {
	privKey, err := helper.GenerateKey(name, email, []byte(passphrase), "x25519", 0)
	if err != nil {
		// TODO: Notify admin
		return KeyPair{}, err
	}
	keyrng, err := crypto.NewKeyFromArmored(privKey)
	if err != nil {
		// TODO: Notify admin
		return KeyPair{}, err
	}

	pubKey, err := keyrng.GetArmoredPublicKey()
	if err != nil {
		// TODO: Notify admin
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}, nil
}
