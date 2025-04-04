# go-nwc

A simple Go library for [NIP-47: Nostr Wallet Connect (NWC)](https://github.com/nostr-protocol/nips/blob/master/47.md).

This package lets you send commands to a Lightning wallet over Nostr using NIP-47. It supports invoice creation, payments, and more.

## Features

- Create Lightning invoices (`make_invoice`)
- Pay BOLT11 invoices (`pay_invoice`)
- Send keysend payments (`pay_keysend`)
- Check wallet balance (`get_balance`)
- Look up invoice/payment info (`lookup_invoice`)
- List wallet transactions (`list_transactions`)
- Get wallet/node info (`get_info`)

## Example

```go
package main

import (
    "fmt"
    "github.com/untreu2/go-nwc"
)

func main() {
    uri := "nostr+walletconnect://<wallet-pubkey>?relay=wss%3A%2F%2Frelay.damus.io&secret=<client-secret>"
    client, err := nwc.NewClient(uri)
    if err != nil {
        panic(err)
    }

    // Get wallet info
    info, _ := client.GetInfo()
    fmt.Println("Wallet Alias:", info.Alias)

    // Check balance
    balance, _ := client.GetBalance()
    fmt.Println("Balance:", balance.Balance, "msat")

    // Create invoice
    invoice, _ := client.MakeInvoice(1000, "test payment")
    fmt.Println("Invoice:", invoice)

    // Pay invoice
    result, _ := client.PayInvoice(invoice)
    fmt.Println("Paid! Preimage:", result.Preimage, "Fees:", result.FeesPaid)

    // Keysend payment
    ksResult, _ := client.PayKeysend("<receiver-pubkey>", 500)
    fmt.Println("Keysend preimage:", ksResult.Preimage)

    // Lookup invoice
    lookup, _ := client.LookupInvoice(invoice)
    fmt.Println("Invoice Type:", lookup.Type, "Hash:", lookup.PaymentHash)

    // List transactions
    txs, _ := client.ListTransactions()
    for _, tx := range txs {
        fmt.Printf("[%s] %d msat - hash: %s\n", tx.Type, tx.Amount, tx.PaymentHash)
    }
}
```
