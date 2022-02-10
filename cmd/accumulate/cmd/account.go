package cmd

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mdp/qrterminal"
	"github.com/spf13/cobra"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	url2 "gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/types"
)

func init() {
	accountCmd.AddCommand(
		accountGetCmd,
		accountCreateCmd,
		accountQrCmd,
		accountGenerateCmd,
		accountListCmd,
		accountRestoreCmd)

	accountCreateCmd.AddCommand(
		accountCreateTokenCmd,
		accountCreateDataCmd)

	accountCreateDataCmd.AddCommand(
		accountCreateDataLiteCmd)

	accountCreateTokenCmd.Flags().BoolVar(&flagAccount.Scratch, "scratch", false, "Create a scratch token account")
	accountCreateDataCmd.Flags().BoolVar(&flagAccount.Scratch, "scratch", false, "Create a scratch data account")
	accountCreateDataCmd.Flags().BoolVar(&flagAccount.Lite, "lite", false, "Create a lite data account")
}

var flagAccount = struct {
	Lite    bool
	Scratch bool
}{}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Create and get token accounts",
}

var accountGetCmd = &cobra.Command{
	Use:   "get [url]",
	Short: "Get an account by URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := GetAccount(args[0])
		printOutput(cmd, out, err)
	},
}

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deprecation Warning!\nTo create a token account, in future please specify either \"token\" or \"data\"\n\n")
		//this will be removed in future release and replaced with usage: PrintAccountCreate()
		out, err := CreateAccount(cmd, args[0], args[1:])
		printOutput(cmd, out, err)
	},
}

var accountCreateTokenCmd = &cobra.Command{
	Use:   "token [actor adi] [signing key name] [key index (optional)] [key height (optional)] [new token account url] [tokenUrl] [keyBook (optional)]",
	Short: "Create an ADI token account",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := CreateAccount(cmd, args[0], args[1:])
		printOutput(cmd, out, err)
	},
}

var accountCreateDataCmd = &cobra.Command{
	Use:   "data",
	Short: "Create a data account",
	Run: func(cmd *cobra.Command, args []string) {
		var out string
		var err error
		if flagAccount.Lite {
			if len(args) < 2 {
				PrintDataLiteAccountCreate()
				return
			}
			out, err = CreateLiteDataAccount(args[0], args[1:])
		} else {
			if len(args) < 3 {
				PrintDataAccountCreate()
				PrintDataLiteAccountCreate()
				return
			}
			out, err = CreateDataAccount(args[0], args[1:])
		}
		printOutput(cmd, out, err)
	},
}

var accountCreateDataLiteCmd = &cobra.Command{
	Use:   "lite",
	Short: "Create a lite data account",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deprecation Warning!\nTo create a lite data account, use `accumulate account create data --lite ...`\n\n")
		if len(args) < 2 {
			PrintDataLiteAccountCreate()
			return
		}
		out, err := CreateLiteDataAccount(args[0], args[1:])
		printOutput(cmd, out, err)
	},
}

var accountQrCmd = &cobra.Command{
	Use:   "qr [url]",
	Short: "Display QR code for lite account URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := QrAccount(args[0])
		printOutput(cmd, out, err)
	},
}

var accountGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate random lite token account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		out, err := GenerateAccount()
		printOutput(cmd, out, err)
	},
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display all lite token accounts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		out, err := ListAccounts()
		printOutput(cmd, out, err)
	},
}

var accountRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore old lite token accounts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		out, err := RestoreAccounts()
		printOutput(cmd, out, err)
	},
}

// Get whatever is stored at the specified Accumulate URL.
func GetAccount(url string) (string, error) {

	// Get WHATEVER is at the provided URL. We'll check later to see what's
	// actually there.
	res, err := GetUrl(url)
	if err != nil {
		// Either the URL was invalid or there was nothing there.
		return "", err
	}

	// Check if what was retrieved is a token account or data account.
	if res.Type != types.AccountTypeTokenAccount.String() && res.Type != types.AccountTypeLiteTokenAccount.String() &&
		res.Type != types.AccountTypeDataAccount.String() && res.Type != types.AccountTypeLiteDataAccount.String() {
		return "", fmt.Errorf("expecting token account or data account but received %v", res.Type)
	}

	return PrintChainQueryResponseV2(res)
}

// Render the given Accumulate URL into a QR code.
func QrAccount(subject string) (string, error) {
	url, err := url2.Parse(subject)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid Accumulate URL: %v\n", subject, err)
	}

	buffer := bytes.NewBufferString("")
	qrterminal.GenerateWithConfig(url.String(), qrterminal.Config{
		Level:          qrterminal.M,
		Writer:         buffer,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		WhiteChar:      qrterminal.WHITE_WHITE,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		QuietZone:      2,
	})

	r, err := ioutil.ReadAll(buffer)
	return string(r), err
}

// CreateAccount begins creation of a token account.
//
// This method is called when a user enters the appropriate command from their
// CLI as defined in accountCreateTokenCmd or accountCreateaDataCmd.
//
// The second parameter is the first user-supplied argument from the CLI
// and all subsequent arguments are arrayed in the third parameter.
//
// TODO: This function should be renamed to reflect that it only creates
// token accounts.
func CreateAccount(cmd *cobra.Command, origin string, args []string) (string, error) {

	// Resolve the first user-supplied argument to a lite account or
	// ADI key page, which is indicated by a URL.
	originURL, err := url2.Parse(origin)
	if err != nil {
		_ = cmd.Usage()
		return "", err
	}

	// Resolve the remaining user-supplied arguments to a signing key.
	// Note that the args used to do this are removed from the arg list.
	args, trxHeader, privKey, err := prepareSigner(originURL, args)
	if err != nil {
		return "", err
	}

	// Aside from the args used to define a signing key, there must be at least
	// two args supplied by the user: a URL at which to create the account and
	// TODO:
	if len(args) < 2 {
		return "", fmt.Errorf("not enough arguments")
	}

	// Resolve the first remaining user-supplied argument to a validated
	// Accumulate URL, which is the address at which to create a new account.
	newAccountURL, err := url2.Parse(args[0])
	if err != nil {
		_ = cmd.Usage()
		return "", fmt.Errorf("invalid account url %s", args[0])
	}
	if originURL.Authority != newAccountURL.Authority {
		return "", fmt.Errorf("account url to create (%s) doesn't match the authority adi (%s)", newAccountURL.Authority, originURL.Authority)
		// TODO: This requires some architecture explaining. Create a document doing that, and point to it here.
	}

	// Resolve the second remaining user-supplied argument to a validated
	// Accumulate URL.
	tokenAccountURL, err := url2.Parse(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid token url")
	}

	// If the user provided three or more arguments in addition to what was
	// required to resolve a signing key, then try to read the third argument
	// to a keybook URL.
	var keybook string
	if len(args) >= 3 {
		keyBookURL, err := url2.Parse(args[2])
		if err != nil {
			return "", fmt.Errorf("invalid key book url")
		}
		keybook = keyBookURL.String()
	}
	// else { keybook = "" }

	// Create a query request. We will be asking what, if anything, exists at
	// the user-supplied token account URL.
	queryRequest := new(api.GeneralQuery)
	queryRequest.Url = tokenAccountURL.String()
	queryResponse := new(api.ChainQueryResponse)
	token := protocol.TokenIssuer{}
	queryResponse.Data = &token
	// Execute the query.
	err = Client.RequestAPIv2(context.Background(), "query", queryRequest, queryResponse)
	// Determine what is actually at the token account URL. We expect a token
	// account.
	if err != nil || queryResponse.Type != types.AccountTypeTokenIssuer.String() {
		return "", fmt.Errorf("invalid token type %v", err)
	}

	// Assemble all the data we've gathered so far.
	newAccount := protocol.CreateTokenAccount{}
	newAccount.Url = newAccountURL.String()
	newAccount.TokenUrl = tokenAccountURL.String()
	newAccount.KeyBookUrl = keybook
	newAccount.Scratch = flagAccount.Scratch

	// Send the account creation request out to the Accumulate network.
	res, err := dispatchTxRequest("create-token-account", &newAccount, nil, originURL, trxHeader, privKey)
	if err != nil {
		return "", err
	}
	return ActionResponseFrom(res).Print()
}

func GenerateAccount() (string, error) {
	return GenerateKey("")
}

func ListAccounts() (string, error) {

	b, err := Db.GetBucket(BucketLabel)
	if err != nil {
		//no accounts so nothing to do...
		return "", err
	}
	var out string
	for _, v := range b.KeyValueList {
		lt, err := protocol.LiteTokenAddress(v.Value, protocol.AcmeUrl().String())
		if err != nil {
			continue
		}
		if lt.String() == string(v.Key) {
			out += fmt.Sprintf("%s\n", v.Key)
		}
	}
	//TODO: this probably should also list out adi accounts as well
	return out, nil
}

func RestoreAccounts() (out string, err error) {
	anon, err := Db.GetBucket(BucketAnon)
	if err != nil {
		//no anon accounts so nothing to do...
		return
	}
	for _, v := range anon.KeyValueList {
		u, err := url2.Parse(string(v.Key))
		if err != nil {
			out += fmt.Sprintf("%q is not a valid URL\n", v.Key)
		}
		key, _, err := protocol.ParseLiteTokenAddress(u)
		if err != nil {
			out += fmt.Sprintf("%q is not a valid lite account: %v\n", v.Key, err)
		} else if key == nil {
			out += fmt.Sprintf("%q is not a lite account\n", v.Key)
		}

		privKey := ed25519.PrivateKey(v.Value)
		pubKey := privKey.Public().(ed25519.PublicKey)
		out += fmt.Sprintf("Converting %s : %x\n", v.Key, pubKey)

		err = Db.Put(BucketLabel, v.Key, pubKey)
		if err != nil {
			log.Fatal(err)
		}
		err = Db.Put(BucketKeys, pubKey, privKey)
		if err != nil {
			return "", err
		}
		err = Db.DeleteBucket(BucketAnon)
		if err != nil {
			return "", err
		}
	}
	return out, nil
}
