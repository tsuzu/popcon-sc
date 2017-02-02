React = require 'react'

module.exports = Balance = React.createClass
    render: ->
      <span className="balance badge">{@props.balance}</span>
