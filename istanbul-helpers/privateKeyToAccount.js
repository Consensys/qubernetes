#!/usr/bin/env node

const Web3 = require('web3');
const web3 = new Web3();
process.argv.forEach(function (val, index, array) {
  if (index > 1) {
   acctObj = web3.eth.accounts.privateKeyToAccount(val);
   console.log(acctObj.address);
  }
});

