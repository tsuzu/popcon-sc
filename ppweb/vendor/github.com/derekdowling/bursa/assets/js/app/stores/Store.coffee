{EventEmitter } = require 'events'

# Uppercases the first letter of a string.
upperCaseFirst = (string) ->
  string.charAt(0) + string.slice(1)

# Returns the standard name of a handler method for a given action.
# Kept private here because generally all you should need a combination of
# Store#canHandle and Store#handlerForAction
actionHandlerName = (action) ->
    "on"+upperCaseFirst(action.constructor.name)

class Store extends EventEmitter
  constructor: ->
    @initialize()

  # The name of this stores specific change event.
  changeEventName: ->
    "change.#{@constructor.name}"

  # Adds a change listener to this store.
  addChangeListener: (callback) ->
    @addListener @changeEventName(), callback

  # Removes a change listener for this store.
  removeChangeListener: (callback) ->
    @removeListener @changeEventName(), callback

  # Emits a change event for this store using it's standardized name.
  emitChange: ->
    @emit @changeEventName()

  # Handle a given action.
  handle: (action) ->
    @[actionHandlerName(action)](action)

  # Whether this store is able to handle the given action.
  canHandle: (action) ->
    @[actionHandlerName(action)]

  # Invoked when an action initiate by the view is dispatched to this store.
  onViewAction: (action) ->
    if @canHandle action
      @handle(action)
    else
      console.debug "#{@constructor.name} does not handle #{action.name()}"

module.exports = Store
