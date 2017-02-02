package client

import (
	"github.com/conformal/btcrpcclient"
	"github.com/derekdowling/bursa/config"
)

// TODO g_ prefix? What is the proper way of initalizing global variables in go?
var g_client *btcrpcclient.Client

// Return a connected client
func Get() *btcrpcclient.Client {
	// NOTE I was really hoping to have bursa.io/config loaded for it's side
	// effects via the underscore modifier, but it seems to not read the
	// configuration data when done that way.
	if g_client == nil {
		g_client, _ = btcrpcclient.New(&btcrpcclient.ConnConfig{
			// There appears to be a bug with nested booleans and viper.GetBool.
			// HttpPostMode: viper.GetBool("bitcoin.HttpPostMode"),
			HttpPostMode: config.App.GetBool("bitcoin.HttpPostMode"),
			DisableTLS:   config.App.GetBool("bitcoin.DisableTLS"),
			Host:         config.App.GetString("bitcoin.Host"),
			User:         config.App.GetString("bitcion.User"),
			Pass:         config.App.GetString("bitcion.Pass"),
		}, nil)
	}
	return g_client
}
