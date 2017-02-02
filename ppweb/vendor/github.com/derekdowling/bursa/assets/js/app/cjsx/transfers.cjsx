# @cjsx React.DOM
React = require 'react'
require 'es6-shim' # {... extraProps}

{ Route, Routes, Link } = require 'react-router'

transfers = require '../support/transfers.coffee'

module.exports = Transfers = React.createClass
    render: ->
      <div className="transfers">
      </div>
