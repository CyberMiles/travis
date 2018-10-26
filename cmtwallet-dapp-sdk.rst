CMT Wallet - dApp SDK Developer doc

 version 0.8    updated at 2018/10/24

Introduction

This document user helps DApp developers access the CMT Wallet DApp SDK.
In general, DApp requires a hosting environment to interact with the user's wallet, just like metamask  CMT Walelet provides this environment in the app.
In DApp browser, DApp can do the same and more things in metamask.
 To keep things simple, this document will use DApp browser for CMT Walelet* DApp browser* , DApp for DApp webpage.

Web3JS

CMT Wallet DApp browser is fully compatible with metamask, you can migrate DApp directly to CMT Wallet without even writing any code.
When the DApp is loaded by the DApp browser, we will inject a web3-cmt.js, so the DApp does not have to have its own built-in web3-cmt.js (but you can do the same), the web3 version we are currently injecting is 0.19, You can access this global object window.web3.
Dapp browser will set web3.cmt.defaultAccount The value of the user is the current wallet address of the user, and the web3 HttpProvider is set to the same as the node configuration of the CMT Wallet.


API

web3.cmt.sendTransaction
For DApp, the most common operation is to send a transaction, usually calling the web3.cmt.sendTransaction method of web3.js, DApp browser will listen to this method call, display a modal layer to let the user see the transaction information. The user can enter the password signature and then send the transaction. After the transaction is successful, it will return a txHash. If it fails, it will return the error value.

Common web3 api:


* Check the current active account on (web3.cmt.coinbase)
* Get the balance of any account (web3.cmt.getBalance)
* Send a transaction (web3.cmt.sendTransaction)
* Sign the message with the private key of the current account (web3.personal.sign)

Error handling
The DApp browser only handles some errors (such as the user entering the wrong password), most of the transaction errors will be returned to the DApp, DApps should handle these errors and prompt the user. We have done i18n processing of the error content, most of the time You can pop up error.message directly.
The user cancels the operation and the Dapp browser returns the error code "1001"

window.cmtwallet.closeDapp()
Close the current DApp page and return to the discovery page

window.cmtwallet.getCurrentLanguage()
Get the user's current language settings. This information may be useful if the DApp needs to support multiple languages, but we have added the locale parameter to the DApp url. In most cases you don't need to call this API.
Available locale:
zh-CN
en-US

window.cmtwallet.getSdkVersion()
Get the current CMT Wallet Dapp SDK version number: 1

window.cmtwallet.getPlatform()
Get the current CMT Wallet phone system
android
ios

Developer mode
In the CMTWallet APP, by default you can't access the DApp by typing (or scanning) a url. You need to open the developer mode first (* I → About us → Click CMT Wallet logo five times*).

dApp SDK Example Link:
https://cube-api.cybermiles.io/static/html/cw/cmtwallet-dappsdk-example.html
