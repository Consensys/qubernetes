#/bin/ash

if [ $# -lt 1 ]; then
  echo " An address to vote in, or out, must be provided: "
  echo " ./propose_ibft.sh HEX_ADDRESS (true|false)"
fi

ADDRESS=$1
VOTE_BOOL=true

if [ $# -eq 2 ]; then
 VOTE_BOOL=$2
fi
RES=$(geth --exec "istanbul.propose(\"$1\", $VOTE_BOOL)" attach ipc:$QUORUM_HOME/dd/geth.ipc)
echo $RES