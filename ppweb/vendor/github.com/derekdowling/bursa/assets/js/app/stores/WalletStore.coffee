AppDispatcher = require '../dispatchers/Dispatcher'
Store = require './Store'

class WalletStore extends Store
  initialize: ->
    @wallets =
      "1KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
        label: "Bursa.io"
        address: "1KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
        balance: 50.0
        wallets:
          "2KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
            label: "Capital"
            address: "2KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
            balance: 20.0
            wallets:
              "5KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
                label: "Infrastructure"
                address: "5KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
                balance: 14.65
                wallets: {}
              "6KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
                label: "Equipment"
                address: "6KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
                balance: 5.35
                wallets: {}
          "3KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
            label: "Marketing"
            address: "3KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
            balance: 21.29
            wallets: {}
          "4KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c":
            label: "Payroll"
            address: "4KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c"
            balance: 8.71
            wallets: {}

  # @see WalletCreateAction
  onWalletCreateAction: (action) ->
    wallet = @findWallet action.parentAddress
    if wallet
      @createWalletUnder wallet
      @emitChange()
    else
      throw new Error "Parent wallet not found"

  # Creates a children wallet directly descended from the provided wallet object.
  createWalletUnder: (parentWallet) ->
    now = ""+Date.now()
    parentWallet.wallets[now] = {
      label: "New Wallet"
      address: now
      balance: 0.0
      wallets: {}
    }

  # Finds a wallet by it's address.
  findWallet: (targetAddress) ->
    findWallet = (wallets) ->
      for address, wallet of wallets
        if address == targetAddress
          return wallet
        else
          wallet = findWallet wallet.wallets
          if wallet?
            return wallet
      return null

    return findWallet @wallets

# We expose a singleton instead of the class itself. Potentially, if we want to
# migrate to full-di at some point, this will make our task easier.
# Also makes subclassing harder.
walletStore = new WalletStore()

AppDispatcher.register walletStore.onViewAction.bind(walletStore)

module.exports = walletStore
