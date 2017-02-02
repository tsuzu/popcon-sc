// This package name is quite generic. I'm not using viper her ATM
// because I'm not sure if I can use it to store / retrieve the btcnet
// struct.
package config

import (
	"github.com/conformal/btcnet"
)

// Network parameters are necessary during the derivation of public keys
// from extended keys. They vary between mainnet and regtest.
// TOLEARN Which parameters exactly are used during the derivation? How do they
// affect things?
//
// TODO This flag being wrong in production could represent a signficant risk.
// Would money be lost if we tried to sign transfers using bunk keys? Or would
// they simply "bounce"?
var btc_network *btcnet.Params

func init() {
	// TODO Production settings comment out below. I would like to manage this
	// via yaml if possible?
	// btc_network = &btcnet.MainNetParams
	btc_network = &btcnet.RegressionNetParams
}

// Return the current configured btcnet.
func BTCNet() *btcnet.Params {
	return btc_network
}
