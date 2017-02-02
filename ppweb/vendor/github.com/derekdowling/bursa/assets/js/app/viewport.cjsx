# @cjsx React.DOM
React = require 'react'
{ RouteHandler } = require('react-router')

Hud = require  './cjsx/hud.cjsx'
Header = require  './cjsx/header.cjsx'
Nav = require  './cjsx/nav.cjsx'

module.exports = Viewport = React.createClass
  render: ->
    <div id="app" className="app">
      <Header />
      <div className="main-container">
          <aside id="nav-container" className="nav-container nav-vertical">
            <Hud />
            <Nav />
          </aside>
          <section id="content" className="content-container animate-fade-up">
             <div className="page page-dashboard">
              <RouteHandler />
             </div>
          </section>
      </div>
    </div>

