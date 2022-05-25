#!/bin/bash

# Stop immediately on error
set -e

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/validate-commons.sh

section "Setup"
if which go > /dev/null || ! which accumulate > /dev/null ; then
    echo "Installing CLI"
    go install ./cmd/accumulate
    export PATH="${PATH}:$(go env GOPATH)/bin"
fi
[ -z "${MNEMONIC}" ] || accumulate --use-unencrypted-wallet key import mnemonic ${MNEMONIC}
echo



section "Generate a Lite Token Account"
accumulate --use-unencrypted-wallet account list 2>&1 | grep -q ACME || accumulate --use-unencrypted-wallet account generate
LITE_ACME=$(accumulate --use-unencrypted-wallet account list -j | jq -re .liteAccounts[0].liteAccount)
LITE_ID=$(cut -d/ -f-3 <<< "$LITE_ACME")
TXS=()
for i in {1..1}
do
	TXS=(${TXS[@]} $(cli-tx faucet ${LITE_ACME}))
done
for tx in "${TXS[@]}"
do
	echo $tx
	wait-for-tx $tx
done

accumulate --use-unencrypted-wallet account get ${LITE_ACME} 1> /dev/null && success || die "Cannot find ${LITE_ACME}"

section "Add credits to lite account"
TXID=$(cli-tx credits ${LITE_ACME} ${LITE_ID} 2700)
wait-for-tx $TXID
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ID} | jq -r .data.creditBalance)
[ "$BALANCE" -ge 2700 ] || die "${LITE_ID} should have at least 2700 credits but only has ${BALANCE}"
success

section "Generate keys"
ensure-key keytest-1-0
ensure-key keytest-2-0
ensure-key keytest-2-1
ensure-key keytest-2-2
ensure-key keytest-2-3-orig
ensure-key keytest-2-3-new
ensure-key keytest-3-0
ensure-key keytest-3-1
ensure-key manager
echo

section "Create an ADI"
wait-for cli-tx adi create ${LITE_ID} keytest keytest-1-0 keytest/book
accumulate --use-unencrypted-wallet adi get keytest 1> /dev/null && success || die "Cannot find keytest"

section "Verify fee charge"
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ID} | jq -r .data.creditBalance)
[ "$BALANCE" -ge 100 ] && success || die "${LITE_ID} should have at least 100 credits but only has ${BALANCE}"

section "Recreating an ADI fails and the synthetic transaction is recorded"
TXID=`cli-tx adi create ${LITE_ID} keytest keytest-1-0 keytest/book` || return 1
wait-for-tx --no-check $TXID
SYNTH=`accumulate --use-unencrypted-wallet tx get -j ${TXID} | jq -re '.syntheticTxids[0]'`
STATUS=`accumulate --use-unencrypted-wallet tx get -j ${SYNTH} | jq --indent 0 .status`
[ $(echo $STATUS | jq -re .delivered) = "true" ] || die "Synthetic transaction was not delivered"
[ $(echo $STATUS | jq -re '.code // 0') -ne 0 ] || die "Synthetic transaction did not fail"
echo $STATUS | jq -re .message 1> /dev/null || die "Synthetic transaction does not have a message"
success

section "Add credits to the ADI's key page 1"
wait-for cli-tx credits ${LITE_ACME} keytest/book/1 60000
BALANCE=$(accumulate --use-unencrypted-wallet -j page get keytest/book/1 | jq -r .data.creditBalance)
[ "$BALANCE" -ge 6000000 ] && success || die "keytest/book/1 should have 6000000 credits but has ${BALANCE}"

section "Create additional Key Pages"
wait-for cli-tx page create keytest/book keytest-1-0 keytest-2-0
wait-for cli-tx page create keytest/book keytest-1-0 keytest-3-0
accumulate --use-unencrypted-wallet page get keytest/book/2 1> /dev/null || die "Cannot find page keytest/book/2"
accumulate --use-unencrypted-wallet page get keytest/book/3 1> /dev/null || die "Cannot find page keytest/book/3"
success

section "Add credits to the ADI's key page 2"
wait-for cli-tx credits ${LITE_ACME} keytest/book/2 1000
BALANCE=$(accumulate --use-unencrypted-wallet -j page get keytest/book/2 | jq -r .data.creditBalance)
[ "$BALANCE" -ge 1000 ] && success || die "keytest/book/2 should have 1000 credits but has ${BALANCE}"

section "Attempting to lock key page 2 using itself fails"
wait-for cli-tx page lock keytest/book/2 keytest-2-0 && die "Key page 2 locked itself" || success

section "Lock key page 2 using page 1"
wait-for cli-tx page lock keytest/book/2 keytest-1-0
success

section "Attempting to update key page 3 using page 2 fails"
cli-tx page key add keytest/book/3 keytest-2-0 1 keytest-3-1 && die "Executed disallowed operation" || success

section "Unlock key page 2 using page 1"
wait-for cli-tx page unlock keytest/book/2 keytest-1-0
success

section "Update key page 3 using page 2"
cli-tx page key add keytest/book/3 keytest-2-0 keytest-3-1
success

section "Add credits to the ADI's key page 2"
wait-for cli-tx credits ${LITE_ACME} keytest/book/2 100
BALANCE=$(accumulate --use-unencrypted-wallet -j page get keytest/book/2 | jq -r .data.creditBalance)
[ "$BALANCE" -ge 100 ] && success || die "keytest/book/2 should have 100 credits but has ${BALANCE}"

section "Add a key to page 2 using a key from page 3"
wait-for cli-tx page key add keytest/book/2 keytest-2-0 1 keytest-2-1
wait-for cli-tx page key add keytest/book/2 keytest-2-0 1 keytest-2-2
wait-for cli-tx page key add keytest/book/2 keytest-2-0 1 keytest-2-3-orig
success

section "Set threshold to 2 of 2"
wait-for cli-tx tx execute keytest/book/2 keytest-2-0 '{"type": "updateKeyPage", "operation": [{ "type": "setThreshold", "threshold": 2 }]}'
THRESHOLD=$(accumulate --use-unencrypted-wallet -j get keytest/book/2 | jq -re .data.threshold)
[ "$THRESHOLD" -eq 2 ] && success || die "Bad keytest/book/2 threshold: want 2, got ${THRESHOLD}"

section "Update a key with only that key's signature"
wait-for cli-tx key update keytest/book/2 keytest-2-3-orig keytest-2-3-new || die "Failed to update key"
accumulate --use-unencrypted-wallet -j get key keytest keytest-2-3-orig > /dev/null && die "Still found old key" || true
accumulate --use-unencrypted-wallet -j get key keytest keytest-2-3-new | jq -C --indent 0 || die "Could not find new key"
success

section "Create an ADI Token Account"
wait-for cli-tx account create token --scratch keytest keytest-1-0 0 keytest/tokens ACME keytest/book
accumulate --use-unencrypted-wallet account get keytest/tokens 1> /dev/null || die "Cannot find keytest/tokens"
accumulate --use-unencrypted-wallet -j account get keytest/tokens | jq -re .data.scratch 1> /dev/null || die "keytest/tokens is not a scratch account"
success

section "Send tokens from the lite token account to the ADI token account"
wait-for cli-tx tx create ${LITE_ACME} keytest/tokens 5
BALANCE=$(accumulate --use-unencrypted-wallet -j account get keytest/tokens | jq -r .data.balance)
[ "$BALANCE" -eq 500000000 ] && success || die "${LITE_ACME} should have 5 tokens but has $(expr ${BALANCE} / 100000000)"

section "Send tokens from the ADI token account to the lite token account using the multisig page"
TXID=$(cli-tx tx create keytest/tokens keytest-2-0 ${LITE_ACME} 1)
wait-for-tx $TXID
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.pending 1> /dev/null || die "Transaction is not pending"
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.delivered 1> /dev/null && die "Transaction was delivered"
success

section "Signing the transaction with the same key does not deliver it"
wait-for cli-tx-sig tx sign keytest/tokens keytest-2-0 $TXID
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.pending 1> /dev/null || die "Transaction is not pending"
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.delivered 1> /dev/null && die "Transaction was delivered"
wait-for-tx $TXID
success

section "Query pending by URL"
accumulate --use-unencrypted-wallet -j get keytest/tokens#pending | jq -re .items[0] &> /dev/null|| die "Failed to retrieve pending transactions"
accumulate --use-unencrypted-wallet -j get keytest/tokens#signature | jq -re .items[0] &> /dev/null|| die "Failed to retrieve signature transactions"
success

section "Query pending chain at height 0 by URL"
TXID=$(accumulate --use-unencrypted-wallet -j get keytest/tokens#pending/0 | jq -re .transactionHash) && success || die "Failed to query pending chain by height"

section "Query pending chain with hash by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/tokens#pending/${TXID} | jq -re .transactionHash) || die "Failed to query pending chain by hash"
[ "$RESULT" == "$TXID" ] && success || die "Querying by height and by hash gives different results"

section "Query pending chain range by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/tokens#pending/0:10 | jq -re .total)
[ "$RESULT" -ge 1 ] && success || die "No entries found"

section "Query signature chain range by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get "keytest/tokens#signature" | jq -re .total) || die "Failed to get entries"
[ "$RESULT" -eq 2 ] || die "Wrong total: want 2, got $RESULT"
success

section "Sign the pending transaction using the other key"
TXID=$(accumulate --use-unencrypted-wallet -j get keytest/tokens#pending | jq -re .items[0])
wait-for cli-tx-sig tx sign keytest/tokens keytest-2-1 $TXID
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.pending 1> /dev/null && die "Transaction is pending"
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.delivered 1> /dev/null || die "Transaction was not delivered"
wait-for-tx $TXID
success

section "Signing the transaction after it has been delivered fails"
cli-tx-sig tx sign keytest/tokens keytest-2-2 $TXID && die "Signed the transaction after it was delivered" || success

# section "Bug AC-551"
# api-v2 '{"jsonrpc": "2.0", "id": 4, "method": "metrics", "params": {"metric": "tps", "duration": "1h"}}' | jq -e .result.data.value 1> /dev/null
# success

section "API v2 faucet (AC-570)"
BEFORE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ACME} | jq -r .data.balance)
wait-for api-tx '{"jsonrpc": "2.0", "id": 4, "method": "faucet", "params": {"url": "'${LITE_ACME}'"}}'
AFTER=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ACME} | jq -r .data.balance)
DIFF=$(expr $AFTER - $BEFORE)
[ $DIFF -eq 200000000000000 ] && success || die "Faucet did not work, want +200000000000000, got ${DIFF}"

section "Parse acme faucet TXNs (API v2, AC-603)"
api-v2 '{ "jsonrpc": "2.0", "id": 0, "method": "query-tx-history", "params": { "url": "7117c50f04f1254d56b704dc05298912deeb25dbc1d26ef6/ACME", "count": 10 } }' | jq -r '.result.items | map(.type)[]' | grep -q acmeFaucet
success

section "Include Merkle state (API, AC-604)"
accumulate --use-unencrypted-wallet -j adi get keytest | jq -e .mainChain.roots 1> /dev/null || die "Failed: response does not include main chain roots"
accumulate --use-unencrypted-wallet -j adi get keytest | jq -e .mainChain.height 1> /dev/null || die "Failed: response does not include main chain height"
api-v2 '{"jsonrpc": "2.0", "id": 0, "method": "query", "params": {"url": "keytest"}}' | jq -e .result.mainChain.roots 1> /dev/null
api-v2 '{"jsonrpc": "2.0", "id": 0, "method": "query", "params": {"url": "keytest"}}' | jq -e .result.mainChain.height 1> /dev/null
success

section "Query with txid and chainId (API v2, AC-602)"
# TODO Verify query-chain
TXID=$(accumulate --use-unencrypted-wallet -j tx history keytest 0 1 | jq -re '.items[0].txid')
GOT=$(api-v2 '{"jsonrpc": "2.0", "id": 0, "method": "query-tx", "params": {"txid": "'${TXID}'"}}' | jq -re .result.txid)
[ "${TXID}" = "${GOT}" ] || die "Failed to find TX ${TXID}"
success

section "Query transaction receipt"
TXID=$(accumulate --use-unencrypted-wallet -j tx history keytest 0 1 | jq -re '.items[0].txid')
(accumulate --use-unencrypted-wallet -j tx get --prove $TXID | jq -e .receipts[0] -C --indent 0) && success || die "Failed to get receipt for ${TXID}"

section "Create a token issuer"
wait-for cli-tx token create keytest keytest-1-0 keytest/token-issuer TOK 10 1000000
accumulate --use-unencrypted-wallet get keytest/token-issuer 1> /dev/null || die "Cannot find keytest/token-issuer"
success

section "Issue tokens"
LITE_TOK=$(echo $LITE_ACME | cut -d/ -f-3)/keytest/token-issuer
wait-for cli-tx token issue keytest/token-issuer keytest-1-0 ${LITE_TOK} 123.0123456789
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_TOK} | jq -r .data.balance)
[ "$BALANCE" -eq 1230123456789 ] && success || die "${LITE_TOK} should have 1230123456789 keytest tokens but has ${BALANCE}"

section "Send tokens to the lite account (TOK)"
wait-for cli-tx tx create ${LITE_ACME} ${LITE_TOK} 10000 # 10000 ACME is 100 TOK
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_TOK} | jq -r .data.balance)
[ "$BALANCE" -ge 2230123456789 ] && success || die "${LITE_TOK} should have at least 2230123456789 tokens but only has ${BALANCE}"

section "Burn tokens"
wait-for cli-tx token burn ${LITE_TOK} 100
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_TOK} | jq -r .data.balance)
[ "$BALANCE" -eq 1230123456789 ] && success || die "${LITE_TOK} should have 1230123456789 keytest tokens but has ${BALANCE}"

section "Create lite data account and write the data"
ACCOUNT_ID=$(accumulate --use-unencrypted-wallet -j account create data --lite keytest keytest-1-0 "Factom PRO" "Tutorial" | jq -r .accountUrl)
[ "$ACCOUNT_ID" == "acc://b36c1c4073305a41edc6353a094329c24ffa54c029a521aa" ] || die "${ACCOUNT_ID} does not match expected value"
accumulate --use-unencrypted-wallet data get $ACCOUNT_ID 0 1 1> /dev/null || die "lite data entry not found"
wait-for cli-tx data write-to keytest keytest-1-0 $ACCOUNT_ID "data test"
accumulate --use-unencrypted-wallet data get $ACCOUNT_ID 0 2 1> /dev/null || die "lite data error"
accumulate --use-unencrypted-wallet -j get "${ACCOUNT_ID}#txn/0" | jq -re .status.result.entryHash &> /dev/null || die "Entry hash is missing from transaction results"
accumulate --use-unencrypted-wallet -j get "${ACCOUNT_ID}#txn/0" | jq -re .status.result.accountID &> /dev/null || die "Account ID is missing from transaction results"
success

section "Create lite data account with first entry"
ACCOUNT_ID=$(accumulate --use-unencrypted-wallet -j account create data --lite keytest keytest-1-0 "First Data Entry" "Check" --lite-data "first entry" | jq -r .accountUrl)
[ "$ACCOUNT_ID" == "acc://4df014cc532c140066add495313e0ffaecba1eba5454cefa" ] || die "${ACCOUNT_ID} does not match expected value"
accumulate --use-unencrypted-wallet -j data get $ACCOUNT_ID 0 1 1> /dev/null || die "lite data entry not found"
wait-for cli-tx data write-to keytest keytest-1-0 $ACCOUNT_ID "data test"
accumulate --use-unencrypted-wallet data get $ACCOUNT_ID 0 2 1> /dev/null || die "lite data error"
accumulate --use-unencrypted-wallet -j get "${ACCOUNT_ID}#txn/0" | jq -re .status.result.entryHash &> /dev/null || die "Entry hash is missing from transaction results"
accumulate --use-unencrypted-wallet -j get "${ACCOUNT_ID}#txn/0" | jq -re .status.result.accountID &> /dev/null || die "Account ID is missing from transaction results"
success

section "Create ADI Data Account"
wait-for cli-tx account create data --scratch keytest keytest-1-0 keytest/data
accumulate --use-unencrypted-wallet account get keytest/data 1> /dev/null || die "Cannot find keytest/data"
accumulate --use-unencrypted-wallet -j account get keytest/data | jq -re .data.scratch 1> /dev/null || die "keytest/data is not a scratch account"
success

section "Write data to ADI Data Account"
JSON=$(accumulate --use-unencrypted-wallet -j data write keytest/data keytest-1-0 foo bar)
TXID=$(echo $JSON | jq -re .transactionHash)
echo $JSON | jq -C --indent 0
wait-for-tx $TXID
echo $JSON | jq -re .result.result.entryHash 1> /dev/null || die "Deliver response does not include the entry hash"
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.result.entryHash 1> /dev/null || die "Transaction query response does not include the entry hash"
success

section "Create a sub ADI"
wait-for cli-tx adi create keytest keytest-1-0 keytest/sub1 keytest-2-0 keytest/sub1/book
accumulate --use-unencrypted-wallet adi get keytest/sub1 1> /dev/null && success || die "Cannot find keytest/sub1"

section "Add credits to the sub ADI's key page 0"
wait-for cli-tx credits ${LITE_ACME} keytest/sub1/book/1 60000
BALANCE=$(accumulate --use-unencrypted-wallet -j page get keytest/sub1/book/1 | jq -r .data.creditBalance)
[ "$BALANCE" -ge 60000 ] && success || die "keytest/sub1/book/1 should have 60000 credits but has ${BALANCE}"

section "Create Data Account for sub ADI"
wait-for cli-tx account create data --scratch keytest/sub1 keytest-2-0 keytest/sub1/data
accumulate --use-unencrypted-wallet account get keytest/sub1/data 1> /dev/null || die "Cannot find keytest/sub1/data"
accumulate --use-unencrypted-wallet -j account get keytest/sub1/data | jq -re .data.scratch 1> /dev/null || die "keytest/sub1/data is not a scratch account"
success

section "Write data to sub ADI Data Account"
JSON=$(accumulate --use-unencrypted-wallet -j data write keytest/sub1/data keytest-2-0 "foo" "bar")
TXID=$(echo $JSON | jq -re .transactionHash)
echo $JSON | jq -C --indent 0
wait-for-tx $TXID
echo $JSON | jq -re .result.result.entryHash 1> /dev/null || die "Deliver response does not include the entry hash"
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.result.entryHash 1> /dev/null || die "Transaction query response does not include the entry hash"
success

section "Issue a new token"
JSON=$(accumulate --use-unencrypted-wallet -j token create keytest keytest-1-0 keytest/foocoin bar 8 )
TXID=$(echo $JSON | jq -re .transactionHash)
echo $JSON | jq -C --indent 0
wait-for-tx $TXID
RESULT=$(accumulate --use-unencrypted-wallet -j token get keytest/foocoin)
[ "$(echo $RESULT | jq -re .data.symbol)" == "bar" ] || die "Token issuance failed with invalid symbol"
[ "$(echo $RESULT | jq -re .data.precision)" -eq 8 ] || die "Token issuance failed with invalid precision"
success

section "Query latest data entry by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/data#data | jq -re .data.entry.data[0])
[ "$RESULT" == $(echo -n bar | xxd -p) ] && success || die "Latest entry is not 'foo', got '$RESULT'"

section "Query data entry at height 0 by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/data#data/0 | jq -re .data.entry.data[0])
[ "$RESULT" == $(echo -n foo | xxd -p) ] && success || die "Entry at height 0 is not 'foo'"

section "Query data entry with hash by URL"
ENTRY=$(accumulate --use-unencrypted-wallet -j get keytest/data#data/0 | jq -re .data.entryHash)
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/data#data/${ENTRY} | jq -re .data.entry.data[0])
ENTRY2=$(accumulate --use-unencrypted-wallet -j get keytest/data#data/${ENTRY} | jq -re .data.entryHash)
[ "$RESULT" == $(echo -n foo | xxd -p) ] || die "Entry with hash ${ENTRY} is not 'foo'"
[ "$ENTRY" == "$ENTRY2" ] || die "Entry hash mismatch ${ENTRY} ${ENTRY2}"
success

section "Query data entry range by URL"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/data#data/0:10 | jq -re .data.total)
[ "$RESULT" -ge 1 ] && success || die "No entries found"

section "Create another ADI (manager)"
wait-for cli-tx adi create ${LITE_ID} manager manager manager/book
accumulate --use-unencrypted-wallet adi get manager 1> /dev/null && success || die "Cannot find manager"

section "Add credits to manager's key page 1"
wait-for cli-tx credits ${LITE_ACME} manager/book/1 1000
BALANCE=$(accumulate --use-unencrypted-wallet -j page get manager/book/1 | jq -r .data.creditBalance)
[ "$BALANCE" -ge 100000 ] && success || die "manager/book/1 should have 100000 credits but has ${BALANCE}"

section "Create token account with manager"
TXID=$(cli-tx account create token keytest keytest-1-0 --authority keytest/book,manager/book keytest/managed-tokens ACME) || "Failed to create managed token account"
accumulate tx sign keytest manager@manager/book $TXID
wait-for-tx --ignore-pending $TXID
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/managed-tokens -j | jq -re '.data.authorities | length')
[ "$RESULT" -eq 2 ] || die "Expected 2 authorities, got $RESULT"
success

section "Remove manager from token account"
TXID=$(cli-tx auth remove keytest/managed-tokens keytest-1-0 manager/book) || die "Failed to initiate txn to remove manager"
wait-for-tx $TXID
accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .status.pending 1> /dev/null || die "Transaction is not pending"
wait-for cli-tx-sig tx sign keytest/managed-tokens manager@manager $TXID || die "Failed to sign transaction"
wait-for-tx --ignore-pending $TXID || die "Transaction was not delivered"
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/managed-tokens -j | jq -re '.data.authorities | length')
[ "$RESULT" -eq 1 ] || die "Expected 1 authority, got $RESULT"
success

section "Add manager to token account"
TXID=$(cli-tx auth add keytest/managed-tokens keytest-1-0 manager/book) || die "Failed to add the manager"
accumulate tx sign keytest manager@manager/book $TXID
wait-for-tx --ignore-pending $TXID
RESULT=$(accumulate --use-unencrypted-wallet -j get keytest/managed-tokens -j | jq -re '.data.authorities | length')
[ "$RESULT" -eq 2 ] || die "Expected 2 authorities, got $RESULT"
success

## For an unknown reason this fails in validate docker
# section "Query the lite identity"
# accumulate --use-unencrypted-wallet -s local get $(dirname $LITE_ACME) -j | jq -e -C --indent 0 .data && success || die "Failed to get $(dirname $LITE_ACME)"

# section "Query the lite identity directory"
# accumulate --use-unencrypted-wallet adi directory $(dirname $LITE_ACME) 0 10 1> /dev/null || die "Failed to get directory for $(dirname $LITE_ACME)"
# TOTAL=$(accumulate --use-unencrypted-wallet -j adi directory $(dirname $LITE_ACME) 0 10 | jq -re .total)
# [ "$TOTAL" -eq 2 ] && success || die "Expected directory 2 entries for $(dirname $LITE_ACME), got $TOTAL"

section "Create ADI Data Account with wait"
accumulate --use-unencrypted-wallet account create data --wait 1m keytest keytest-1-0 keytest/data1 1> /dev/null || die "Failed to create account"
accumulate --use-unencrypted-wallet account get keytest/data1 1> /dev/null || die "Cannot find keytest/data1"

section "Query credits"
RESULT=$(accumulate --use-unencrypted-wallet -j oracle  | jq -re .price)
[ "$RESULT" -ge 0 ] && success || die "Expected 500, got $RESULT"

section "Transaction with Memo"
TXID=$(cli-tx tx create keytest/tokens keytest-1-0 ${LITE_ACME} 1 --memo memo)
wait-for-tx $TXID
MEMO=$(accumulate --use-unencrypted-wallet -j tx get $TXID | jq -re .transaction.header.memo) || die "Failed to query memo"
[ "$MEMO" == "memo" ] && success || die "Expected memo, got $MEMO"

section "Refund on expensive synthetic txn failure"
BALANCE=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ID} | jq -r .data.creditBalance)
wait-for --no-check cli-tx adi create ${LITE_ID} keytest keytest-1-0 keytest/book
echo "sleeping for 10 sec"
sleep 10 &
wait
BALANCE1=$(accumulate --use-unencrypted-wallet -j account get ${LITE_ID} | jq -r .data.creditBalance)
BALANCE=$((BALANCE-100))
[ "$BALANCE" -eq "$BALANCE1" ] && success || die "Expected $BALANCE, got $BALANCE1"

section "Token refund on txn failure"
BALANCE=$(accumulate --use-unencrypted-wallet -j account get keytest/tokens | jq -r .data.balance)
TXID=$(cli-tx tx create keytest/tokens keytest-2-0 acc://invalid-account 1)
wait-for-tx $TXID
BALANCE1=$(accumulate --use-unencrypted-wallet -j account get keytest/tokens | jq -r .data.balance)
[ $BALANCE -eq $BALANCE1 ] && success || die "Expected $BALANCE, got $BALANCE1"
