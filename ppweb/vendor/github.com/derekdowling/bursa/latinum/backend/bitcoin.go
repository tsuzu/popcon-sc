package backend

import (
	"errors"
	"fmt"
	"log"

	"github.com/derekdowling/bursa/latinum/backend/client"
	shared_config "github.com/derekdowling/bursa/latinum/shared/config"
	"github.com/derekdowling/bursa/latinum/vault"

	"github.com/conformal/btcjson"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	_ "github.com/derekdowling/bursa/config"
)

// Not quite sure what we want here yet
type TransferResponse struct {
	msg  string
	code int
}

type Transfer struct {
}

type CurrencyBackend interface {
	ExecuteTransfer(transfer *Transfer) *TransferResponse
}

type Latinum struct {
	client *btcrpcclient.Client
}

// Generates bitcoins and gives them to an address. Used only during testing
// in regtest mode where setgenerate is actually functional.
//
// When we generate our initial bitcoins for test purposes, they're implicitly
// associated with bitcoind's default wallet. Generally, all of Bursa's other
// operations take the burden of wallet management out of bitcoind.
//
// Amt shoult be set to > 101 in order to ensure we've confirmed some blocks.
// to make the newly minted bitcoins spendable. Yes - we're "mining" here. The
// 100 confirmations thing is to iron out blockchain forks.
func GenerateInto(amt float64, src_private_key *btcutil.WIF, src_address btcutil.Address, encoded_address string) error {
	var addresses []btcutil.Address

	addresses = append(addresses, src_address)

	// Not sure what sensible values here are. The min confirmations shouldn't be
	unspent, err := client.Get().ListUnspentMinMaxAddresses(10, 1000, addresses)

	// Aggregate unspent transactions until we have more than then requested amount.
	// Who needs ruby? Good old for loops.
	current_amt := 0.0
	var inputs []btcjson.TransactionInput
	var amounts = make(map[btcutil.Address]btcutil.Amount)

	for _, utxo := range unspent {
		if current_amt > amt {
			break
		}

		inputs = append(inputs, btcjson.TransactionInput{Txid: utxo.TxId, Vout: utxo.Vout})
		current_amt += utxo.Amount
	}

	if current_amt < amt {
		return errors.New("Insufficient funds in server wallet")
	}

	fmt.Println("current amt:", current_amt)

	// Transaction fee is the difference between in/out.
	tx_fee := 0.001

	// Calculate change to send back to ourselves.
	change := current_amt - tx_fee - amt

	fmt.Println("change", change)
	fmt.Println("amt", amt)
	fmt.Println("tx_fee", tx_fee)
	fmt.Println("new address", src_address.String())
	fmt.Println("encoded address", encoded_address)

	// All of these different address types are painful to juggle. We should settle
	// on a common denominator - if possible. That may not be the case as some of
	// the transformation are irreversible - e.g. if we pass along a public key
	// hash, we can't derive the public key that generated it.
	dest_address, err := btcutil.DecodeAddress(encoded_address, shared_config.BTCNet())
	if err != nil {
		log.Print("Couldn't decode destination address", err)
		return err
	}

	amounts[src_address], _ = btcutil.NewAmount(change)
	amounts[dest_address], _ = btcutil.NewAmount(amt)

	unsigned_raw_tx, err := client.Get().CreateRawTransaction(inputs, amounts)

	fmt.Println(inputs)

	// TODO we may want to return a new error rather than the descended one because
	// it ends up leaking the underlying abstraction details to our caller.
	if err != nil {
		log.Print("Couldn't generate unsigned raw transaction", err)
		return err
	}

	// WIF is the format returned by bitcoin-cli dumpprivkey
	signed, err := vault.SignWithEncodedWIFKey(
		unsigned_raw_tx,
		src_private_key.String(),
	)

	if err != nil {
		log.Print("Couldn't sign the damn thing.", err)
		return err
	}

	fmt.Println("signed")
	sha_hash, err := client.Get().SendRawTransaction(signed, false)

	if err != nil {
		log.Fatalf("Couldn't send the signed transaction.", err)
	}

	log.Println(sha_hash)

	return err
}

// Sends `amount` bitcoin between the source bursa user and destination bursa
// user.
func SendBetween(amt float64, src_user_id int64, dest_user_id int64) error {
	var addresses []btcutil.Address

	src_address := vault.GetAddressForUser(src_user_id)
	dest_address := vault.GetAddressForUser(dest_user_id)

	addresses = append(addresses, src_address)

	// Not sure what sensible values here are. The min confirmations shouldn't be
	unspent, err := client.Get().ListUnspentMinMaxAddresses(10, 1000, addresses)

	// Aggregate unspent transactions until we have more than then requested amount.
	// Who needs ruby? Good old for loops.
	// Functionify this.
	current_amt := 0.0
	var inputs []btcjson.TransactionInput
	var amounts = make(map[btcutil.Address]btcutil.Amount)

	for _, utxo := range unspent {
		if current_amt > amt {
			break
		}

		inputs = append(inputs, btcjson.TransactionInput{Txid: utxo.TxId, Vout: utxo.Vout})
		current_amt += utxo.Amount
	}

	if current_amt < amt {
		return errors.New("Insufficient funds in src wallet")
	}

	// Transaction fee is the difference between in/out.
	tx_fee := 0.001

	// Calculate change to send back to ourselves.
	change := current_amt - tx_fee - amt

	amounts[src_address], _ = btcutil.NewAmount(change)
	amounts[dest_address], _ = btcutil.NewAmount(amt)

	unsigned_raw_tx, err := client.Get().CreateRawTransaction(inputs, amounts)

	fmt.Println(inputs)

	// TODO we may want to return a new error rather than the descended one because
	// it ends up leaking the underlying abstraction details to our caller.
	if err != nil {
		log.Print("Couldn't generate unsigned raw transaction", err)
		return err
	}

	// WIF is the format returned by bitcoin-cli dumpprivkey
	signed, err := vault.SignByUserId(
		unsigned_raw_tx,
		src_user_id,
	)

	if err != nil {
		log.Print("Couldn't sign the damn thing.", err)
		return err
	}

	sha_hash, err := client.Get().SendRawTransaction(signed, false)

	if err != nil {
		log.Fatalf("Couldn't send the signed transaction.", err)
	}

	log.Println(sha_hash)

	return err

}
