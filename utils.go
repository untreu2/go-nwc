package nwc

import (
	"fmt"
	"net/url"
)

func ParseNWCURI(uri string) (relayURL, walletPubKey, secret string, err error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return
	}
	walletPubKey = parsed.Host
	if walletPubKey == "" {
		walletPubKey = parsed.Path
		if len(walletPubKey) > 0 && walletPubKey[0] == '/' {
			walletPubKey = walletPubKey[1:]
		}
	}
	query := parsed.Query()
	relayURL, err = url.QueryUnescape(query.Get("relay"))
	if err != nil || relayURL == "" {
		err = fmt.Errorf("relay parametresi eksik")
		return
	}
	secret = query.Get("secret")
	if secret == "" {
		err = fmt.Errorf("secret parametresi eksik")
		return
	}
	return
}
