package pgp

import (
	"log"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func SignData(data string, privateKey string, passphrase []byte) (string, error) {
	log.Println("Signing data", data)
	privateKeyObj, err := crypto.NewKeyFromArmored(privateKey)
	if err != nil {
		// TODO: Notify admin
		return "", err
	}

	unlockedKeyObj, err := privateKeyObj.Unlock(passphrase)
	if err != nil {
		return "", err
	}

	var message = crypto.NewPlainMessageFromString(data)
	signingKeyRing, err := crypto.NewKeyRing(unlockedKeyObj)
	if err != nil {
		return "", err
	}

	pgpSignature, err := signingKeyRing.SignDetached(message)
	if err != nil {
		return "", err
	}

	armored, err := pgpSignature.GetArmored()
	if err != nil {
		return "", err
	}

	return armored, nil
}
