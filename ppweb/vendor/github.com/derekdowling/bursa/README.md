# Bursa

![Build Status](https://magnum.travis-ci.com/derekdowling/bursa.svg?token=iq92sEsRxzbuqGK3drsX&branch=master)

### Installing the Backend

	* Clone the repository into your go/src/github.com directory

	* CD into /bursa and run "godep restore".
	
	* sudo ln -s $GOPATH/src/github/derekdowling/bursa /bursa
	  
	  Vagrant mounts the /bursa location into /bursa/bursa (Yes, confusing). 
	  We just want it in a consistent location across dev machines so that we
	  don't need to configure our .kitchen.yml file shared folder config.
	  

### Running The Website

Assuming you have already installed chef/kitchen below:

  * To launch the web server

				go run src/bursa.io/server/server.go or ./start-server

  * To run all the tests:

				go test

### Running the Convey Test Server

  * Simply run the script

				./test-server
				
## Dependency Management

Dependencies in go are weird. There's a great discussion here: https://news.ycombinator.com/item?id=7109090

If you're running Bursa locally:

1. Move your project directory into your GOPATH. This is annoying.
2. Place a symlink where your project directory used to be if you like your normal developer layout the way it is.

Now, godep will actually work:

	godep save ./...
	
### FAQ

#### I'm changing my code, but it seems to be using old / cached version.

	rm -rf $GOPATH/pkg/github.com/derekdowling/bursa
	
E.g. go will sometimes use the built-in versions of your code inside of the package directory. There's a method to the madness. I don't know it. When someone figures it out, update this faq.


#### Testing in Vagrant is slow.

Well, godep IMHO should be building all the packages but I'm not sure if it does. In particular, the conformal btec package is a dog. You can build it with `go build $GOPATH/conformal/btcec.`

If you add the -x flag when running tests you'll see what go is doing. Note any expensive build steps for dependencies and build them. Don't forget about the earlier FAQ question if you do.