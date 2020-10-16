#!/usr/bin/env bash

function usage() {
  echo ""
  echo "Usage:"
  echo "    $0 [--keyName fileNamePrefix] [--keepOpenSSLFiles]"
  echo ""
  echo "Where:"
  echo "    --keyName specifies the prefix to the generated public(.pub)/private(.key) key files"
  echo "    --keepOpenSSLFiles doesn't delete the intermediary openssl public/private keys"
  echo ""
  exit -1
}

keepOpenSSLFiles=false

while (("$#")); do
  case "$1" in
  --keyName)
    keyName=$2
    shift 2
    ;;
  --keepOpenSSLFiles)
    keepOpenSSLFiles=true
    shift 1
    ;;
  --help)
    shift
    usage
    ;;
  *)
    echo "Error: Unsupported command line parameter $1"
    usage
    ;;
  esac
done

if [ "$keyName" == "" ]; then
  echo "Error: Please specify a key name prefix."
  exit -1
fi

#TODO - check whether any files may be overwritten by the script
#TODO - check that all the utilities are available and have the necessary versions

echo "Key name prefix:" $keyName
echo "keepOpenSSLFiles:" $keepOpenSSLFiles

openSSLPrivateKeyFile="$keyName-ossl.key"

openssl genpkey -algorithm x25519 -out $openSSLPrivateKeyFile

openSSLHexOut="$keyName-ossl.text"

openssl pkey -in $openSSLPrivateKeyFile -text >$openSSLHexOut

cat $openSSLHexOut
#sed picks up the relevant lines from the output (6-8 for the public key and 10-12 for the private key)
#tr removes any spaces/:/EOL characters and leaves just the alphanumeric ones
#xxd converts the hex input to binary output
#base64 converts the binary input to based64 encoded output
privateKey=$(sed -n '6,8 p' $openSSLHexOut | tr -cd '[:alnum:]' | xxd -r -p | base64)
publicKey=$(sed -n '10,12 p' $openSSLHexOut | tr -cd '[:alnum:]' | xxd -r -p | base64)

publicKeyFile="$keyName.pub"
privateKeyFile="$keyName.key"

#using printf in order to avoid the newline that echo adds
printf "$publicKey" >$publicKeyFile
printf "{\"data\":{\"bytes\":\"$privateKey\"},\"type\":\"unlocked\"}" >$privateKeyFile

if [ "${keepOpenSSLFiles}" == "false" ]; then
  rm $openSSLPrivateKeyFile $openSSLHexOut
fi
