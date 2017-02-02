Dispatcher = require '../dispatchers/Dispatcher'
expect = require('chai').expect

class Action
  # Fluently delegates to a constructor.
  @build: ->
    new @(arguments...)

  dispatch: ->
    Dispatcher.dispatch @

  name: ->
    "#{@constructor.name}"

class WalletAction extends Action

class WalletCreateAction extends WalletAction
  # The address of the parent wallet.
  constructor: (@parentAddress) ->
    expect(@parentAddress).to.be.a('string')

class WalletDestroyAction extends WalletAction
  constructor: (@wallet) ->

module.exports = { WalletAction, WalletCreateAction, WalletDestroyAction }
