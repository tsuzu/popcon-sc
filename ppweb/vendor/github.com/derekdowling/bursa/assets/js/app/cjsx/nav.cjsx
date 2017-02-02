# @cjsx React.DOM
React = require 'react'

{ Route, Routes, Link, ActiveState } = require 'react-router'

NavLink = require './navlink.cjsx'

# <li><Link to="/wallets"><i className="fa fa-bitcoin"></i><span>Wallets</span></Link></li>
# <li><Link to="/transfers"><i className="fa fa-send"></i><span>Transfers</span></Link></li>
# <li><Link to="/delegate"><i className="fa fa-share-alt"></i><span>Delegate</span></Link></li>
Nav = React.createClass
    render: ->
      <div className="nav-wrapper">
        <ul id="nav" className="nav" data-slim-scroll data-collapse-nav data-highlight-active>
          <NavLink to='/wallets'><i className="fa fa-bitcoin"></i><span>Wallets</span></NavLink>
          <NavLink to='/transfers'><i className="fa fa-send"></i><span>Tranfers</span></NavLink>
          <NavLink to='/delegate'><i className="fa fa-share-alt"></i><span>Delegate</span></NavLink>
        </ul>
      </div>

module.exports = Nav
