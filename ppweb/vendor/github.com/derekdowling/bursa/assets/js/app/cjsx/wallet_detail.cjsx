# @cjsx React.DOM
React = require 'react'
require 'es6-shim' # {... extraProps}

{ Navigation } = require 'react-router'

Balance = require './balance.cjsx'
Hash = require './hash.cjsx'

module.exports = WalletDetail = React.createClass
    mixins: [ Navigation ]
    getInitialState: ->
      label: "Bursa.io"
      address: "1KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
      balance: 50.0

    render: ->
      <div className="wallets">
        <div className="row">
          <div className="col-sm-12">
              <h2>Wallets</h2>
          </div>
        </div>
      </div>
