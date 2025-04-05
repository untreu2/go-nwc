package nwc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
)

type Client struct {
	RelayURL     string
	WalletPubKey string
	ClientSecret string
	ClientPubKey string
}

type InvoiceDetails struct {
	Type            string                 `json:"type"`
	Invoice         string                 `json:"invoice"`
	Description     string                 `json:"description"`
	DescriptionHash string                 `json:"description_hash"`
	Preimage        string                 `json:"preimage"`
	PaymentHash     string                 `json:"payment_hash"`
	Amount          int64                  `json:"amount"`
	FeesPaid        int64                  `json:"fees_paid"`
	CreatedAt       int64                  `json:"created_at"`
	ExpiresAt       int64                  `json:"expires_at"`
	SettledAt       int64                  `json:"settled_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type TransactionDetails = InvoiceDetails

func NewClient(uri string) (*Client, error) {
	relayURL, walletPubKey, secret, err := ParseNWCURI(uri)
	if err != nil {
		return nil, err
	}
	clientPubKey, err := nostr.GetPublicKey(secret)
	if err != nil {
		return nil, err
	}
	return &Client{
		RelayURL:     relayURL,
		WalletPubKey: walletPubKey,
		ClientSecret: secret,
		ClientPubKey: clientPubKey,
	}, nil
}

func (c *Client) sendRequest(method string, params map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"method": method,
		"params": params,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	sharedSecret, err := nip04.ComputeSharedSecret(c.WalletPubKey, c.ClientSecret)
	if err != nil {
		return "", err
	}
	encryptedContent, err := nip04.Encrypt(string(payloadBytes), sharedSecret)
	if err != nil {
		return "", err
	}

	ev := nostr.Event{
		PubKey:    c.ClientPubKey,
		CreatedAt: nostr.Now(),
		Kind:      23194,
		Tags:      nostr.Tags{{"p", c.WalletPubKey}},
		Content:   encryptedContent,
	}
	if err := ev.Sign(c.ClientSecret); err != nil {
		return "", err
	}

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, c.RelayURL)
	if err != nil {
		return "", err
	}
	defer relay.Close()

	if err := relay.Publish(ctx, ev); err != nil {
		return "", err
	}

	filters := nostr.Filters{{
		Kinds:   []int{23195},
		Authors: []string{c.WalletPubKey},
		Tags:    nostr.TagMap{"e": []string{ev.ID}},
		Limit:   1,
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		return "", err
	}

	for e := range sub.Events {
		decrypted, err := nip04.Decrypt(e.Content, sharedSecret)
		if err != nil {
			return "", err
		}
		return decrypted, nil
	}
	return "", errors.New("yanıt alınamadı")
}

func (c *Client) GetBalance() (*struct {
	Balance int64 `json:"balance"`
}, error) {
	raw, err := c.sendRequest("get_balance", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			Balance int64 `json:"balance"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp.Result, nil
}

func (c *Client) MakeInvoice(amount int, memo string) (string, error) {
	raw, err := c.sendRequest("make_invoice", map[string]interface{}{
		"amount":      amount,
		"description": memo,
	})
	if err != nil {
		return "", err
	}
	var resp struct {
		Result InvoiceDetails `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", err
	}
	return resp.Result.Invoice, nil
}

func (c *Client) PayInvoice(invoice string) (*struct {
	Preimage string `json:"preimage"`
	FeesPaid int64  `json:"fees_paid"`
}, error) {
	raw, err := c.sendRequest("pay_invoice", map[string]interface{}{
		"invoice": invoice,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			Preimage string `json:"preimage"`
			FeesPaid int64  `json:"fees_paid"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp.Result, nil
}

func (c *Client) PayKeysend(pubkey string, amount int) (*struct {
	Preimage string `json:"preimage"`
	FeesPaid int64  `json:"fees_paid"`
}, error) {
	raw, err := c.sendRequest("pay_keysend", map[string]interface{}{
		"amount": amount,
		"pubkey": pubkey,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			Preimage string `json:"preimage"`
			FeesPaid int64  `json:"fees_paid"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp.Result, nil
}

func (c *Client) LookupInvoice(param string) (*InvoiceDetails, error) {
	raw, err := c.sendRequest("lookup_invoice", map[string]interface{}{
		"invoice": param,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result InvoiceDetails `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp.Result, nil
}

func (c *Client) ListTransactions() ([]TransactionDetails, error) {
	raw, err := c.sendRequest("list_transactions", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			Transactions []TransactionDetails `json:"transactions"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return resp.Result.Transactions, nil
}

func (c *Client) GetInfo() (*struct {
	Alias         string   `json:"alias"`
	Pubkey        string   `json:"pubkey"`
	Network       string   `json:"network"`
	Methods       []string `json:"methods"`
	Color         string   `json:"color"`
	BlockHeight   int64    `json:"block_height"`
	BlockHash     string   `json:"block_hash"`
	Notifications []string `json:"notifications"`
}, error) {
	raw, err := c.sendRequest("get_info", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			Alias         string   `json:"alias"`
			Pubkey        string   `json:"pubkey"`
			Network       string   `json:"network"`
			Methods       []string `json:"methods"`
			Color         string   `json:"color"`
			BlockHeight   int64    `json:"block_height"`
			BlockHash     string   `json:"block_hash"`
			Notifications []string `json:"notifications"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return &resp.Result, nil
}
