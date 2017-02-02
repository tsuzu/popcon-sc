package backend

import (
	"bytes"
	"github.com/conformal/btcjson"
	"github.com/conformal/btcutil"
	"github.com/derekdowling/bursa/klutz"
	"github.com/derekdowling/bursa/latinum/backend/client"
	shared_config "github.com/derekdowling/bursa/latinum/shared/config"
	"github.com/derekdowling/bursa/latinum/vault"
	"github.com/derekdowling/bursa/models"
	"github.com/derekdowling/bursa/testutils"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"os/exec"
	"testing"
	"time"
	"io/ioutil"
)

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

// Runs a command and bails if it fails.
func requireExec(fail_message string, name string, params ...string) {
	var out bytes.Buffer
	cmd := exec.Command(name, params...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatalf(fail_message, err, out.String())
	} else {
		out.Reset()
	}
}

// Establish our test state.
// And then I realized we could just back up and restore the regtest directory.
// YOLO
func reset(root_user_id int64) (btcutil.Address, *btcutil.WIF) {
	requireExec("Couldn't reset regtest db", "rm", "-rf", "/home/vagrant/.bitcoin/regtest")
	requireExec("Couldn't restart bitcoind", "service", "bitcoin-server", "restart")

	if err := klutz.Flail(
		10,
		time.Duration(1500)*time.Millisecond,
		func() error { return client.Get().Ping() },
	); err != nil {

		log.Fatalf("bitcoin-server didnt seem to restart.", err)
	}

	if err := client.Get().SetGenerate(true, 101); err != nil {
		log.Fatalf("Couldn't list utxos", err)
	}

	var unspent []btcjson.ListUnspentResult
	var err error

	unspent, err = client.Get().ListUnspent()
	if err != nil {
		log.Fatalf("Couldn't list utxos", err)
	}

	if len(unspent) != 1 {
		log.Fatalf("Expected 1 unspent utxo got:", unspent)
	}

	local_addresses := []string{}

	current_amt := 0.0
	var inputs []btcjson.TransactionInput

	// There should only be 1 iteration here as we anticipate a single utxo.
	for _, utxo := range unspent {
		local_addresses = append(local_addresses, utxo.Address)

		inputs = append(inputs, btcjson.TransactionInput{Txid: utxo.TxId, Vout: utxo.Vout})
		current_amt += utxo.Amount
	}

	// There should be one unspent transaction with a random address that we want
	// to return to our caller.
	encoded_src_address := local_addresses[0]
	src_address, _ = btcutil.DecodeAddress(
		encoded_src_address,
		shared_config.BTCNet(),
	)

	// Bitcoind starts up with some random private key that we need to steal.
	private_src_key, _ = client.Get().DumpPrivKey(src_address)

	ioutil.WriteFile("/bursa/bursa/latinum/fixtures/regtest_pk.key", private_src_key, 0777)

	return src_address, private_src_key
}

func TestSpec(t *testing.T) {
	return;
	db, err := models.Connect()
	if err != nil {
		log.Fatalf("Couldn't connect to database during testing", err)
	}

	// TODO Doing this all over is ugly.
	user_a := models.User{
		Name: testutils.SuffixedId("bitcoin_test_user_a"),
	}
	db.Save(&user_a)

	user_b := models.User{
		Name: testutils.SuffixedId("bitcoin_test_user_a"),
	}
	db.Save(&user_b)

	root_user := models.User{
		Name: testutils.SuffixedId("root_user"),
	}
	db.Save(&root_user)

	// Sets up our initial root user with 50 bitcoins.
	root_address, root_pk := reset(root_user.Id)

	Convey("Latinum Tests", t, func() {
		Convey("Generate()", func() {
			Convey("Should generate bitcoins for testing", func() {
				// Generate()
			})
		})

		Convey("Send()", func() {
			Convey("Should send bitcoins from a to b", func() {
				// address_a = vault.NewMaster()

				// key_a, _ = vault.NewMaster()
				// key_b, _ = vault.NewMaster()

				// address_a, _ = vault.GetEncodedAddress(key_a)
				// address_b, _ = vault.GetEncodedAddress(key_b)
			})
		})

		Convey("GenerateIntoAddress()", func() {
			Convey("Should send bitcoins from a to b", func() {
				key_a, _ := vault.NewMaster()
				// TODO these should probably return a potential error.
				address_a := vault.GetEncodedAddress(key_a)
				err := GenerateInto(0.5, root_pk, root_address, address_a)
				So(err, ShouldBeNil)
			})

			Convey("Should send bitcoins between users", func() {
				var err error

				_, err = vault.NewMasterForUser(user_a.Id)
				So(err, ShouldBeNil)

				_, err = vault.NewMasterForUser(user_b.Id)
				So(err, ShouldBeNil)

				SendBetween(0.5, user_a.Id, user_b.Id)
			})
		})
	})
}
