React = require 'react'

module.exports = Hash = React.createClass
    render: ->
      <span className="hash">{@props.hash}</span>
