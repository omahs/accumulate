package cmd

import (
	"fmt"
	"strconv"

	url2 "github.com/AccumulateNetwork/accumulate/internal/url"
	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Issue and get tokens",
}

var tokenCmdGet = &cobra.Command{
	Use:   "get [url]",
	Short: "get token by URL",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := GetToken(args[0])
		printOutput(cmd, out, err)
	},
}

var tokenCmdCreate = &cobra.Command{
	Use:   "create [origin adi or lite url] [adi signer key name (if applicable)] [token url] [symbol] [precision (0 - 18)] [properties URL (optional)]",
	Short: "Create new token",
	Run: func(cmd *cobra.Command, args []string) {
		var out string
		var err error
		if len(args) > 1 {
			out, err = CreateToken(args[0], args[1:])
		} else {
			fmt.Println("Usage:")
			PrintTokenCreate()
		}
		printOutput(cmd, out, err)
	},
}

var tokenCmdIssue = &cobra.Command{
	Use:   "issue [adi token url] [signer key name] [recipient url] [amount]",
	Short: "send tokens from a token url to a recipient",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var out string
		var err error
		if len(args) > 1 {
			out, err = IssueTokenToRecipient(args[0], args[1:])
		} else {
			fmt.Println("Usage:")
			PrintTokenCreate()
		}
		printOutput(cmd, out, err)
	},
}

func init() {
	tokenCmd.AddCommand(
		tokenCmdGet,
		tokenCmdCreate,
		tokenCmdIssue)
}

func PrintTokenGet() {
	fmt.Println("  accumulate token get [url] Get token by URL")
}

func PrintTokenCreate() {
	fmt.Println("  accumulate token create [origin adi url] [signer key name] [token url] [symbol] [precision (0 - 18)] [properties URL (optional)] 	Create new token")
	fmt.Println("  accumulate token create [origin lite url] [token url] [symbol] [precision (0 - 18)] [properties URL (optional)] 	Create new token")
}

func GetToken(url string) (string, error) {
	res, err := GetUrl(url)
	if err != nil {
		return "", err
	}

	return PrintChainQueryResponseV2(res)

}

func CreateToken(origin string, args []string) (string, error) {
	originUrl, err := url2.Parse(origin)
	if err != nil {
		return "", err
	}

	args, si, privKey, err := prepareSigner(originUrl, args)
	if err != nil {
		return "", err
	}

	if len(args) < 4 {
		return "", fmt.Errorf("insufficient number of arguments")
	}

	url := args[0]
	symbol := args[1]
	precision := args[2]
	properties := args[3]

	prcsn, err := strconv.Atoi(precision)
	if err != nil {
		return "", err
	}

	u, err := url2.Parse(url)
	if err != nil {
		return "", err
	}

	params := protocol.CreateToken{}
	params.Url = u.String()
	params.Symbol = symbol
	params.Precision = uint64(prcsn)

	if len(properties) > 0 {
		//todo: add GetPropertiesUrl-> as a good measure if propertiesUrl is specified, then we should make sure the url is already populated and valid
		u, err := url2.Parse(properties)
		if err != nil {
			return "", fmt.Errorf("invalid properties url, %v", err)
		}
		params.Properties = u.String()
	}

	params.Properties = properties

	res, err := dispatchTxRequest("create-token", &params, nil, originUrl, si, privKey)
	if err != nil {
		return "", err
	}

	return ActionResponseFrom(res).Print()
}

func IssueTokenToRecipient(origin string, args []string) (string, error) {
	originUrl, err := url2.Parse(origin)
	if err != nil {
		return "", err
	}

	args, si, privKey, err := prepareSigner(originUrl, args)
	if err != nil {
		return "", err
	}

	if len(args) < 2 {
		return "", fmt.Errorf("insufficient number of parameters provided")
	}
	recipient, err := url2.Parse(args[0])
	if err != nil {
		return "", err
	}

	//query the token prescision and reformat amount argument into a bigInt.
	amt, err := amountToBigInt(originUrl.String(), args[1])
	if err != nil {
		return "", err
	}

	params := protocol.IssueTokens{}
	params.Recipient = recipient.String()
	params.Amount.Set(amt)

	res, err := dispatchTxRequest("create-token", &params, nil, originUrl, si, privKey)
	if err != nil {
		return "", err
	}

	return ActionResponseFrom(res).Print()
}
