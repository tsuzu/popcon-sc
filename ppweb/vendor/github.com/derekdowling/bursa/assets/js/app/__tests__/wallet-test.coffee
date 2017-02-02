jest.dontMock('../stores/WalletStore')
WalletStore = require '../stores/WalletStore'

describe 'Wallet Tests', ->
  it 'Finds the root wallet', ->
    expect(WalletStore.findWallet("1KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c").label)
      .toEqual("Bursa.io")
  it 'Finds a level 2 wallet', ->
    expect(WalletStore.findWallet("2KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c").label)
      .toEqual("Capital")
  it 'Finds a level 2 sibling wallet', ->
    expect(WalletStore.findWallet("3KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c").label)
      .toEqual("Marketing")
  it 'Finds a level 3 wallet', ->
    expect(WalletStore.findWallet("6KNp2RrFvtRLh7FX6qAwYzqN6d1bmM849c").label)
      .toEqual("Equipment")
