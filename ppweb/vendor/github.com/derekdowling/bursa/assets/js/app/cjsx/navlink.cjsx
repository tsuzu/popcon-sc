# @cjsx React.DOM
React = require 'react'

{ Route, Routes, Link, State } = require 'react-router'

module.exports = NavLink = React.createClass
  mixins: [ State ]

  render: ->
    isActive = @isActive(@props.to, @getParams(), @getQuery())
    className = if isActive  then 'active' else ''
    <li className={className}><Link {... @props}/></li>
