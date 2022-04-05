package cmd

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"github.com/spf13/cobra"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	api2 "gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	url2 "gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/pkg/client/signing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/types"
)

func runCmdFunc(fn func([]string) (string, error)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		out, err := fn(args)
		printOutput(cmd, out, err)
	}
}

func getRecord(urlStr string, rec interface{}) (*api2.MerkleState, error) {
	u, err := url2.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	params := api2.UrlQuery{
		Url: u,
	}
	res := new(api2.ChainQueryResponse)
	res.Data = rec
	if err := Client.RequestAPIv2(context.Background(), "query", &params, res); err != nil {
		return nil, err
	}
	return res.MainChain, nil
}

func prepareSigner(origin *url2.URL, args []string) ([]string, *signing.Signer, error) {
	ct := 0
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("insufficent arguments on comand line")
	}

	signer := new(signing.Signer)
	signer.Type = protocol.SignatureTypeLegacyED25519
	signer.Timestamp = nonceFromTimeNow()

	if IsLiteAccount(origin.String()) {
		privKey, err := LookupByLabel(origin.String())
		if err != nil {
			return nil, nil, fmt.Errorf("unable to find private key for lite token account %s %v", origin.String(), err)
		}

		signer.Url = origin
		signer.Version = 1
		signer.PrivateKey = privKey
		return args, signer, nil
	}

	privKey, err := resolvePrivateKey(args[0])
	if err != nil {
		return nil, nil, err
	}
	signer.PrivateKey = privKey
	ct++

	keyInfo, err := getKey(origin.String(), privKey[32:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get key for %q : %v", origin, err)
	}

	if len(args) < 2 {
		signer.Url = keyInfo.KeyPage
	} else if v, err := strconv.ParseUint(args[1], 10, 64); err == nil {
		signer.Url = protocol.FormatKeyPageUrl(keyInfo.KeyBook, v)
		ct++
	} else {
		signer.Url = keyInfo.KeyPage
	}

	var page *protocol.KeyPage
	_, err = getRecord(signer.Url.String(), &page)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get %q : %v", keyInfo.KeyPage, err)
	}
	signer.Version = page.Version

	return args[ct:], signer, nil
}

func parseArgsAndPrepareSigner(args []string) ([]string, *url2.URL, *signing.Signer, error) {
	principal, err := url2.Parse(args[0])
	if err != nil {
		return nil, nil, nil, err
	}

	args, signer, err := prepareSigner(principal, args[1:])
	if err != nil {
		return nil, nil, nil, err
	}

	return args, principal, signer, nil
}

func IsLiteAccount(url string) bool {
	u, err := url2.Parse(url)
	if err != nil {
		log.Fatal(err)
	}
	key, _, _ := protocol.ParseLiteTokenAddress(u)
	return key != nil
}

// Remarshal uses mapstructure to convert a generic JSON-decoded map into a struct.
func Remarshal(src interface{}, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// This is a hack to reduce how much we have to change
type QueryResponse struct {
	Type           string                      `json:"type,omitempty"`
	MainChain      *api2.MerkleState           `json:"mainChain,omitempty"`
	Data           interface{}                 `json:"data,omitempty"`
	ChainId        []byte                      `json:"chainId,omitempty"`
	Origin         string                      `json:"origin,omitempty"`
	KeyPage        *api2.KeyPage               `json:"keyPage,omitempty"`
	Txid           []byte                      `json:"txid,omitempty"`
	Signatures     []protocol.Signature        `json:"signatures,omitempty"`
	Status         *protocol.TransactionStatus `json:"status,omitempty"`
	SyntheticTxids [][32]byte                  `json:"syntheticTxids,omitempty"`
}

func GetUrl(url string) (*QueryResponse, error) {
	var res QueryResponse

	u, err := url2.Parse(url)
	if err != nil {
		return nil, err
	}
	params := api2.UrlQuery{}
	params.Url = u

	err = queryAs("query", &params, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func getAccount(url string) (protocol.Account, error) {
	qr, err := GetUrl(url)
	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(qr.Data)
	if err != nil {
		return nil, err
	}

	return protocol.UnmarshalAccountJSON(json)
}

func queryAs(method string, input, output interface{}) error {
	err := Client.RequestAPIv2(context.Background(), method, input, output)
	if err == nil {
		return nil
	}

	ret, err := PrintJsonRpcError(err)
	if err != nil {
		return err
	}

	return fmt.Errorf("%v", ret)
}

func dispatchTxAndPrintResponse(action string, payload protocol.TransactionBody, txHash []byte, origin *url2.URL, signer *signing.Signer) (string, error) {
	res, err := dispatchTxRequest(action, payload, txHash, origin, signer)
	if err != nil {
		return "", err
	}

	return ActionResponseFrom(res).Print()
}

func dispatchTxRequest(action string, payload protocol.TransactionBody, txHash []byte, origin *url2.URL, signer *signing.Signer) (*api2.TxResponse, error) {
	var env *protocol.Envelope
	var sig protocol.Signature
	var err error
	switch {
	case payload != nil && txHash == nil:
		env, err = buildEnvelope(payload, origin)
		if err != nil {
			return nil, err
		}
		sig, err = signer.Initiate(env.Transaction)
	case payload == nil && txHash != nil:
		payload = new(protocol.SignPending)
		env = new(protocol.Envelope)
		env.TxHash = txHash
		env.Transaction = new(protocol.Transaction)
		env.Transaction.Body = payload
		env.Transaction.Header.Principal = origin
		sig, err = signer.Sign(txHash)
	default:
		panic("cannot specify a transaction hash and a payload")
	}
	if err != nil {
		return nil, err
	}
	env.Signatures = append(env.Signatures, sig)

	req := new(api2.TxRequest)
	req.TxHash = txHash
	req.Origin = env.Transaction.Header.Principal
	req.Signer.Timestamp = sig.GetTimestamp()
	req.Signer.Url = sig.GetSigner()
	req.Signer.PublicKey = sig.GetPublicKey()
	req.KeyPage.Version = sig.GetSignerVersion()
	req.Signature = sig.GetSignature()
	req.Memo = env.Transaction.Header.Memo
	req.Metadata = env.Transaction.Header.Metadata

	if TxPretend {
		req.CheckOnly = true
	}

	if action == "execute" {
		dataBinary, err := payload.MarshalBinary()
		if err != nil {
			return nil, err
		}
		req.Payload = hex.EncodeToString(dataBinary)
	} else {
		req.Payload = payload
	}
	if err != nil {
		return nil, err
	}

	var res api2.TxResponse
	if err := Client.RequestAPIv2(context.Background(), action, req, &res); err != nil {
		_, err := PrintJsonRpcError(err)
		return nil, err
	}

	return &res, nil
}

func buildEnvelope(payload protocol.TransactionBody, origin *url2.URL) (*protocol.Envelope, error) {
	env := new(protocol.Envelope)
	env.Transaction = new(protocol.Transaction)
	env.Transaction.Body = payload
	env.Transaction.Header.Principal = origin
	env.Transaction.Header.Memo = Memo

	if Metadata == "" {
		return env, nil
	}

	if !strings.Contains(Metadata, ":") {
		env.Transaction.Header.Metadata = []byte(Metadata)
		return env, nil
	}

	dataSet := strings.Split(Metadata, ":")
	switch dataSet[0] {
	case "hex":
		bytes, err := hex.DecodeString(dataSet[1])
		if err != nil {
			return nil, err
		}
		env.Transaction.Header.Metadata = bytes
	case "base64":
		bytes, err := base64.RawStdEncoding.DecodeString(dataSet[1])
		if err != nil {
			return nil, err
		}
		env.Transaction.Header.Metadata = bytes
	default:
		env.Transaction.Header.Metadata = []byte(dataSet[1])
	}
	return env, nil
}

type ActionResponse struct {
	TransactionHash types.Bytes                 `json:"transactionHash"`
	SignatureHashes []types.Bytes               `json:"signatureHashes"`
	SimpleHash      types.Bytes                 `json:"simpleHash"`
	Log             types.String                `json:"log"`
	Code            types.String                `json:"code"`
	Codespace       types.String                `json:"codespace"`
	Error           types.String                `json:"error"`
	Mempool         types.String                `json:"mempool"`
	Result          *protocol.TransactionStatus `json:"result"`
}

type ActionDataResponse struct {
	EntryHash types.Bytes32 `json:"entryHash"`
	ActionResponse
}

type ActionLiteDataResponse struct {
	AccountUrl types.String  `json:"accountUrl"`
	AccountId  types.Bytes32 `json:"accountId"`
	ActionDataResponse
}

func ActionResponseFromLiteData(r *api2.TxResponse, accountUrl string, accountId []byte, entryHash []byte) *ActionLiteDataResponse {
	ar := &ActionLiteDataResponse{}
	ar.AccountUrl = types.String(accountUrl)
	_ = ar.AccountId.FromBytes(accountId)
	ar.ActionDataResponse = *ActionResponseFromData(r, entryHash)
	return ar
}

func ActionResponseFromData(r *api2.TxResponse, entryHash []byte) *ActionDataResponse {
	ar := &ActionDataResponse{}
	_ = ar.EntryHash.FromBytes(entryHash)
	ar.ActionResponse = *ActionResponseFrom(r)
	return ar
}

func ActionResponseFrom(r *api2.TxResponse) *ActionResponse {
	ar := &ActionResponse{
		TransactionHash: r.TransactionHash,
		SignatureHashes: make([]types.Bytes, len(r.SignatureHashes)),
		SimpleHash:      r.SimpleHash,
		Error:           types.String(r.Message),
		Code:            types.String(fmt.Sprint(r.Code)),
	}
	for i, hash := range r.SignatureHashes {
		ar.SignatureHashes[i] = hash
	}

	result := new(protocol.TransactionStatus)
	if Remarshal(r.Result, result) != nil {
		return ar
	}

	ar.Code = types.String(fmt.Sprint(result.Code))
	ar.Error = types.String(result.Message)
	return ar
}

type JsonRpcError struct {
	Msg string
	Err jsonrpc2.Error
}

func (e *JsonRpcError) Error() string { return e.Msg }

var (
	ApiToString = map[protocol.AccountType]string{
		protocol.AccountTypeLiteTokenAccount: "Lite Account",
		protocol.AccountTypeTokenAccount:     "ADI Token Account",
		protocol.AccountTypeIdentity:         "ADI",
		protocol.AccountTypeKeyBook:          "Key Book",
		protocol.AccountTypeKeyPage:          "Key Page",
		protocol.AccountTypeDataAccount:      "Data Chain",
		protocol.AccountTypeLiteDataAccount:  "Lite Data Chain",
	}
)

func amountToBigInt(tokenUrl string, amount string) (*big.Int, error) {
	//query the token
	qr, err := GetUrl(tokenUrl)
	if err != nil {
		return nil, fmt.Errorf("error retrieving token url, %v", err)
	}
	t := protocol.TokenIssuer{}
	err = Remarshal(qr.Data, &t)
	if err != nil {
		return nil, err
	}

	amt, _ := big.NewFloat(0).SetPrec(128).SetString(amount)
	if amt == nil {
		return nil, fmt.Errorf("invalid amount %s", amount)
	}
	oneToken := big.NewFloat(math.Pow(10.0, float64(t.Precision)))
	amt.Mul(amt, oneToken)
	iAmt, _ := amt.Int(big.NewInt(0))
	return iAmt, nil
}

func GetTokenUrlFromAccount(u *url2.URL) (*url2.URL, error) {
	var err error
	var tokenUrl *url2.URL
	if IsLiteAccount(u.String()) {
		_, tokenUrl, err = protocol.ParseLiteTokenAddress(u)
		if err != nil {
			return nil, fmt.Errorf("cannot extract token url from lite token account, %v", err)
		}
	} else {
		res, err := GetUrl(u.String())
		if err != nil {
			return nil, err
		}
		if res.Type != protocol.AccountTypeTokenAccount.String() {
			return nil, fmt.Errorf("expecting token account but received %s", res.Type)
		}
		ta := protocol.TokenAccount{}
		err = Remarshal(res.Data, &ta)
		if err != nil {
			return nil, fmt.Errorf("error remarshaling token account, %v", err)
		}
		tokenUrl = ta.TokenUrl
	}
	if tokenUrl == nil {
		return nil, fmt.Errorf("invalid token url was obtained from %s", u.String())
	}
	return tokenUrl, nil
}
func amountToString(precision uint64, amount *big.Int) string {
	bf := big.Float{}
	bd := big.Float{}
	bd.SetFloat64(math.Pow(10.0, float64(precision)))
	bf.SetInt(amount)
	bal := big.Float{}
	bal.Quo(&bf, &bd)
	return bal.Text('f', int(precision))
}

func formatAmount(tokenUrl string, amount *big.Int) (string, error) {
	//query the token
	tokenData, err := GetUrl(tokenUrl)
	if err != nil {
		return "", fmt.Errorf("error retrieving token url, %v", err)
	}
	t := protocol.TokenIssuer{}
	err = Remarshal(tokenData.Data, &t)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", amountToString(t.Precision, amount), t.Symbol), nil
}

func natural(name string) string {
	var splits []int

	var wasLower bool
	for i, r := range name {
		if wasLower && unicode.IsUpper(r) {
			splits = append(splits, i)
		}
		wasLower = unicode.IsLower(r)
	}

	w := new(strings.Builder)
	w.Grow(len(name) + len(splits))

	var word string
	var split int
	var offset int
	for len(splits) > 0 {
		split, splits = splits[0], splits[1:]
		split -= offset
		offset += split
		word, name = name[:split], name[split:]
		w.WriteString(word)
		w.WriteRune(' ')
	}

	w.WriteString(name)
	return w.String()
}

func nonceFromTimeNow() uint64 {
	t := time.Now()
	return uint64(t.Unix()*1e6) + uint64(t.Nanosecond())/1e3
}

func QueryAcmeOracle() (*protocol.AcmeOracle, error) {
	params := api.DataEntryQuery{}
	params.Url = protocol.PriceOracle()

	res := new(api.ChainQueryResponse)
	entry := new(api.DataEntryQueryResponse)
	res.Data = entry

	err := Client.RequestAPIv2(context.Background(), "query-data", &params, &res)
	if err != nil {
		return nil, err
	}

	if entry.Entry.Data == nil {
		return nil, fmt.Errorf("no data in oracle account")
	}
	acmeOracle := new(protocol.AcmeOracle)
	if err = json.Unmarshal(entry.Entry.Data[0], acmeOracle); err != nil {
		return nil, err
	}
	return acmeOracle, err
}
