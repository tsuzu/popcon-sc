Promise = require 'bluebird'

class Dispatcher
  register: (callback) ->
    @callbacks ||= []
    @callbacks.push callback
    @callbacks.length - 1

  dispatch: (payload) ->
    resolves = []
    rejects = []

    # Expose resolvers / rejecters. E.g. as if we had Q.defer()
    @callbacks.forEach (_, i) ->
      new Promise (resolve, reject) ->
        resolves[i] = resolve
        rejects[i] = reject

    @callbacks.forEach (callback, i) ->
      Promise.resolve(callback(payload)).then ->
        resolves[i](payload)
      # Failure
      ,->
        rejects[i](new Error('Dispatcher failed'))

    null

class AppDispatcher extends Dispatcher
  onAction: (action) ->
    @dispatch source: 'ACTION_DISPATCHER', action: action

# Expose a singleton instance.
appDispatcher = new AppDispatcher()

module.exports = appDispatcher
