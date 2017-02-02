# @cjsx React.DOM
React = require 'react'

Hud = React.createClass
    render: ->
      <div className="hud">
        <h4>&mdash;Jacob Straszynski&mdash;</h4>
        <i className="fa fa-3x fa-group"></i>
        <div className="balance">3.54</div>
      </div>

module.exports = Hud
