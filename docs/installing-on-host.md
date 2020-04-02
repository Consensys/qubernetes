## Install
```shell
$> brew install ruby

# check ruby version > 2.6
$> ruby --version
   ruby 2.6.3
$> gem install colorize
```

### Install Prerequisites For Generating Resources on Host (without Docker) 
* [`bootnode`](https://github.com/ethereum/go-ethereum/tree/master/cmd/bootnode) (geth) for generating keys. 
   ```
     # what you should see if installed.
     $> bootnode
     Fatal: Use -nodekey or -nodekeyhex to specify a private key
   ```
   
   If you have geth source on your machine: 
   ```
    $> cd go-ethereum 
    go-ethereum $> make all
    # or place this in your .bash_profile or equivalent file
    $> export PATH="~/go/src/github.com/ethereum/go-ethereum/build/bin:$PATH"
   ```
* [nodejs](https://nodejs.org/en/download/) Istanbul only.
  ```
   # tested with version 10.15
   $> node --version
   v10.15.
   ```
* web3 `$> npm web3` Istanbul only.

* [constellation-node](https://github.com/jpmorganchase/constellation)
  ```
  $> brew install berkeley-db leveldb libsodium
  $> brew install haskell-stack
  $> git clone https://github.com/jpmorganchase/constellation.git WHATEVER/DIRECTORY
  $> cd constellation
  constellation $> stack setup
  constellation $> stack install
  ```

* [istanbul-tools](https://github.com/jpmorganchase/istanbul-tools) Istanbul only.
  ```
   # install
   $> go get github.com/jpmorganchase/istanbul-tools/cmd/istanbul  
   ```