package cmd

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/AccumulateNetwork/accumulated/types/api/transactions"
	"log"
	"strconv"
	"strings"
	"time"

	url2 "github.com/AccumulateNetwork/accumulated/internal/url"
	"github.com/AccumulateNetwork/accumulated/protocol"
	"github.com/AccumulateNetwork/accumulated/types"
	acmeapi "github.com/AccumulateNetwork/accumulated/types/api"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Create and manage Keys, Books, and Pages",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			switch arg := args[0]; arg {
			//case "get":
			//	if len(args) > 2 {
			//		if args[1] == "page" {
			//			GetKey(args[2], "sig-spec")
			//		} else if args[1] == "book" {
			//			GetKey(args[2], "sig-spec-group")
			//		} else {
			//			fmt.Println("Usage:")
			//			PrintKeyGet()
			//		}
			//	} else {
			//		fmt.Println("Usage:")
			//		PrintKeyGet()
			//	}
			//case "create":
			//	if len(args) > 3 {
			//		CreateKeyBookOrPage(args[1], args[2], args[3:])
			//	} else {
			//		fmt.Println("Usage:")
			//		PrintKeyCreate()
			//	}
			case "import":
				ImportMneumonic("seed", args[1:])
			case "book":
				if len(args) == 3 {
					if args[1] == "get" {
						GetKey(args[2], "sig-spec-group")
					} else {
						fmt.Println("Usage:")
						PrintKeyBookGet()
					}
				} else if len(args) > 3 {
					if args[1] == "create" {
						CreateKeyBook(args[2], args[3:])
					} else {
						fmt.Println("Usage:")
						PrintKeyBookCreate()
					}
				}
			case "page":
				if len(args) == 3 {
					if args[1] == "get" {
						GetKey(args[2], "sig-spec")
					} else {
						fmt.Println("Usage:")
						PrintKeyPageGet()
					}
				} else if len(args) > 3 {
					switch args[1] {
					case "create":
						CreateKeyPage(args[2], args[3:])
					case "update":
						KeyPageUpdate(args[2], protocol.UpdateKey, args[3:])
					case "add":
						KeyPageUpdate(args[2], protocol.AddKey, args[3:])
					case "remove":
						KeyPageUpdate(args[2], protocol.RemoveKey, args[3:])
					default:
						fmt.Println("Usage:")
						PrintKeyPageCreate()
						PrintKeyUpdate()
					}
				} else {
					fmt.Println("Usage:")
					PrintKeyPageCreate()
					PrintKeyUpdate()
				}
			case "list":
				ListKeyPublic()
			case "generate":
				if len(args) > 1 {
					GenerateKey(args[1])
				} else {
					PrintKeyGenerate()
				}
			default:
				fmt.Println("Usage:")
				PrintKey()
			}
		} else {
			fmt.Println("Usage:")
			PrintKey()
		}

	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
}

func PrintKeyPublic() {
	fmt.Println("  accumulate key list			List generated keys associated with the wallet")
}
func PrintKeyBookGet() {
	fmt.Println("  accumulate key get book [URL]			Get existing Key Book by URL")
}

func PrintKeyPageGet() {
	fmt.Println("  accumulate key get page [URL]			Get existing Key Page by URL")
}

func PrintKeyBookCreate() {
	fmt.Println("  accumulate key book create [actor adi url] [signing key label] [key index (optional)] [key height (optional)] [new key book url] [key page url 1] ... [key page url n] Create new key page with 1 to N public keys")
	fmt.Println("\t\t example usage: accumulate key book create acc://RedWagon redKey5 acc://RedWagon/RedBook acc://RedWagon/RedPage1")
}
func PrintKeyPageCreate() {
	fmt.Println("  accumulate key page create [actor adi url] [signing key label] [key index (optional)] [key height (optional)] [new key page url] [public key label 1] ... [public key label n] Create new key page with 1 to N public keys within the wallet")
	fmt.Println("\t\t example usage: accumulate key page create acc://RedWagon redKey5 acc://RedWagon/RedPage1 redKey1 redKey2 redKey3")
}
func PrintKeyUpdate() {
	fmt.Println("  accumulate key page update [key page url] [signing key label] [key index (optional)] [key height (optional)] [old key label] [new public key or label] Update key page with a new public key")
	fmt.Println("\t\t example usage: accumulate key update page  acc://RedWagon redKey5 acc://RedWagon/RedPage1 redKey1 redKey2 redKey3")
	fmt.Println("  accumulate key page add [key page url] [signing key label] [key index (optional)] [key height (optional)] [new key label] Add key to key page")
	fmt.Println("\t\t example usage: accumulate key add page acc://RedWagon redKey5 acc://RedWagon/RedPage1 redKey1 redKey2 redKey3")
	fmt.Println("  accumulate key page remove [key page url] [signing key label] [key index (optional)] [key height (optional)] [old key label] Remove key in key page")
	fmt.Println("\t\t example usage: accumulate key add page acc://RedWagon redKey5 acc://RedWagon/RedPage1 redKey1 redKey2 redKey3")

}

func PrintKeyGenerate() {
	fmt.Println("  accumulate key generate [label]     Generate a new key and give it a label in the wallet")
}

func PrintKey() {
	PrintKeyBookGet()
	PrintKeyPageGet()
	PrintKeyBookCreate()
	PrintKeyPageCreate()
	PrintKeyGenerate()
	PrintKeyPublic()
}

func GetKeyPage(book string, keyLabel string) (*protocol.SigSpec, int, error) {

	b, err := url2.Parse(book)
	if err != nil {
		log.Fatal(err)
	}

	privKey, err := LookupByLabel(keyLabel)
	if err != nil {
		log.Fatal(err)
	}

	kb, err := GetKeyBook(b.String())
	if err != nil {
		log.Fatal(err)
	}

	for i := range kb.SigSpecs {
		v := kb.SigSpecs[i]
		//we have a match so go fetch the ssg
		s, err := GetByChainId(v[:])
		if err != nil {
			log.Fatal(err)
		}
		if *s.Type.AsString() != types.ChainTypeSigSpec.Name() {
			log.Fatal(fmt.Errorf("expecting key page, received %s", s.Type))
		}
		ss := protocol.SigSpec{}
		err = ss.UnmarshalBinary(*s.Data)
		if err != nil {
			log.Fatal(err)
		}

		for j := range ss.Keys {
			_, err := LookupByPubKey(ss.Keys[j].PublicKey)
			if err == nil && bytes.Equal(privKey[32:], v[:]) {
				return &ss, j, nil
			}
		}
	}

	return nil, 0, fmt.Errorf("key page not found in book %s for key label %s", book, keyLabel)
}

func GetKeyBook(url string) (*protocol.SigSpecGroup, error) {
	s, err := GetKey(url, "sig-spec-group")
	if err != nil {
		log.Fatal(err)
	}

	ssg := protocol.SigSpecGroup{}
	err = json.Unmarshal([]byte(s), &ssg)
	if err != nil {
		log.Fatal(err)
	}

	return &ssg, nil
}

func GetKey(url string, method string) ([]byte, error) {

	var res interface{}
	var str []byte

	u, err := url2.Parse(url)
	params := acmeapi.APIRequestURL{}
	params.URL = types.String(u.String())

	if err := Client.Request(context.Background(), method, params, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

	return str, nil
}

//func CreateKeyBookOrPage(createType string, pageUrl string, args []string) {
//	u, err := url2.Parse(pageUrl)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	args, si, privKey, err := prepareSigner(u, args)
//	if err != nil {
//		PrintKeyBookCreate()
//		log.Fatal(err)
//	}
//	if len(args) < 2 {
//		PrintKeyCreate()
//		log.Fatal(fmt.Errorf("invalid number of arguments"))
//	}
//	newUrl, err := url2.Parse(args[0])
//	switch createType {
//	case "page":
//		CreateKeyPage(u, si, privKey, newUrl, args[1:])
//	case "book":
//		CreateKeyBook(u, si, privKey, newUrl, args[1:])
//	}
//}

// CreateKeyPage create a new key page
func CreateKeyPage(page string, args []string) {

	pageUrl, err := url2.Parse(page)
	if err != nil {
		PrintKeyPageCreate()
		log.Fatal(err)
	}

	args, si, privKey, err := prepareSigner(pageUrl, args)
	if err != nil {
		PrintKeyBookCreate()
		log.Fatal(err)
	}
	if len(args) < 2 {
		PrintKeyPageCreate()
		log.Fatal(fmt.Errorf("invalid number of arguments"))
	}
	newUrl, err := url2.Parse(args[0])
	keyLabels := args[1:]
	//when creating a key page you need to have the keys already generated and labeled.
	if newUrl.Authority != pageUrl.Authority {
		PrintKeyPageCreate()
		log.Fatalf("page url to create (%s) doesn't match the authority adi (%s)", newUrl.Authority, pageUrl.Authority)
	}

	css := protocol.CreateSigSpec{}
	ksp := make([]*protocol.KeySpecParams, len(keyLabels))
	css.Url = newUrl.String()
	css.Keys = ksp
	for i := range keyLabels {
		ksp := protocol.KeySpecParams{}

		pk, err := LookupByLabel(keyLabels[i])
		if err != nil {
			//now check to see if it is a valid key hex, if so we can assume that is the public key.
			ksp.PublicKey, err = pubKeyFromString(keyLabels[i])
			if err != nil {
				PrintKeyPageCreate()
				log.Fatal(fmt.Errorf("key %s, does not exist in wallet, nor is it a valid public key", keyLabels[i]))
			}
		} else {
			ksp.PublicKey = pk[32:]
		}

		css.Keys[i] = &ksp
	}

	data, err := json.Marshal(css)
	if err != nil {
		PrintKeyPageCreate()
		log.Fatal(err)
	}

	dataBinary, err := css.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(time.Now().Unix())
	params, err := prepareGenTx(data, dataBinary, pageUrl, si, privKey, nonce)
	if err != nil {
		log.Fatal(err)
	}

	var res interface{}
	var str []byte
	if err := Client.Request(context.Background(), "create-sig-spec", params, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

}

func resolveKey(key string) ([]byte, error) {
	ret, err := getPublicKey(key)
	if err != nil {
		ret, err = pubKeyFromString(key)
		if err != nil {
			PrintKeyUpdate()
			return nil, fmt.Errorf("key %s, does not exist in wallet, nor is it a valid public key", key)
		}
	}
	return ret, err
}

func KeyPageUpdate(actorUrl string, op protocol.KeyPageOperation, args []string) {

	u, err := url2.Parse(actorUrl)
	if err != nil {
		log.Fatal(err)
	}

	args, si, privKey, err := prepareSigner(u, args)
	if err != nil {
		PrintKeyUpdate()
		log.Fatal(err)
	}

	var newKey []byte
	var oldKey []byte

	ukp := protocol.UpdateKeyPage{}
	ukp.Operation = op
	switch op {
	case protocol.UpdateKey:
		if len(args) < 2 {
			PrintKeyUpdate()
			log.Fatal(fmt.Errorf("invalid number of arguments"))
		}
		oldKey, err = resolveKey(args[0])
		if err != nil {
			PrintKeyUpdate()
			log.Fatal(err)
		}
		newKey, err = resolveKey(args[1])
		if err != nil {
			PrintKeyUpdate()
			log.Fatal(err)
		}
	case protocol.AddKey:
		if len(args) < 1 {
			PrintKeyUpdate()
			log.Fatal(fmt.Errorf("invalid number of arguments"))
		}
		newKey, err = resolveKey(args[0])
		if err != nil {
			PrintKeyUpdate()
			log.Fatal(err)
		}
	case protocol.RemoveKey:
		if len(args) < 1 {
			PrintKeyUpdate()
			log.Fatal(fmt.Errorf("invalid number of arguments"))
		}
		oldKey, err = resolveKey(args[0])
		if err != nil {
			PrintKeyUpdate()
			log.Fatal(err)
		}
	}

	ukp.Key = oldKey[:]
	ukp.NewKey = newKey[:]
	data, err := json.Marshal(&ukp)
	if err != nil {
		log.Fatal(err)
	}

	dataBinary, err := ukp.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(time.Now().Unix())
	params, err := prepareGenTx(data, dataBinary, u, si, privKey, nonce)
	if err != nil {
		log.Fatal(err)
	}

	var res interface{}
	var str []byte
	if err := Client.Request(context.Background(), "key-page-update", params, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

}
func dispatchRequest(action string, payload interface{}, actor *url2.URL, si *transactions.SignatureInfo, privKey []byte) (interface{}, error) {
	json.Marshal(payload)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	dataBinary, err := payload.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(time.Now().Unix())
	params, err := prepareGenTx(data, dataBinary, actor, si, privKey, nonce)
	if err != nil {
		log.Fatal(err)
	}

	var res interface{}
	if err := Client.Request(context.Background(), "create-sig-spec-group", params, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// CreateKeyBook create a new key page
func CreateKeyBook(book string, args []string) {

	bookUrl, err := url2.Parse(book)
	if err != nil {
		log.Fatal(err)
	}

	args, si, privKey, err := prepareSigner(bookUrl, args)
	if err != nil {
		PrintKeyBookCreate()
		log.Fatal(err)
	}
	if len(args) < 2 {
		PrintKeyBookCreate()
		log.Fatal(fmt.Errorf("invalid number of arguments"))
	}

	newUrl, err := url2.Parse(args[0])

	if newUrl.Authority != bookUrl.Authority {
		log.Fatalf("book url to create (%s) doesn't match the authority adi (%s)", newUrl.Authority, bookUrl.Authority)
	}

	ssg := protocol.CreateSigSpecGroup{}
	ssg.Url = newUrl.String()

	var chainId types.Bytes32
	pageUrls := args[1:]
	for i := range pageUrls {
		u2, err := url2.Parse(pageUrls[i])
		if err != nil {
			log.Fatalf("invalid page url %s, %v", pageUrls[i], err)
		}
		chainId.FromBytes(u2.ResourceChain())
		ssg.SigSpecs = append(ssg.SigSpecs, chainId)
	}

	//res, err := dispatchRequest("create-sig-spec-group", &ssg, bookUrl, si, privKey)
	//if err != nil {
	//	log.Fatalf("error dispatching request %v", err)
	//}
	//str, err := json.Marshal(res)
	//if err != nil {
	//	log.Fatal(err)
	//}

	data, err := json.Marshal(&ssg)
	if err != nil {
		log.Fatal(err)
	}

	dataBinary, err := ssg.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	nonce := uint64(time.Now().Unix())
	params, err := prepareGenTx(data, dataBinary, bookUrl, si, privKey, nonce)
	if err != nil {
		log.Fatal(err)
	}

	var res interface{}
	var str []byte
	if err := Client.Request(context.Background(), "create-sig-spec-group", params, &res); err != nil {
		log.Fatal(err)
	}

	str, err = json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(str))

}

func pubKeyFromString(s string) ([]byte, error) {
	var pubKey types.Bytes32
	if len(s) != 64 {
		return nil, fmt.Errorf("invalid public key or wallet key label")
	}
	i, err := hex.Decode(pubKey[:], []byte(s))

	if err != nil {
		return nil, err
	}

	if i != 32 {
		return nil, fmt.Errorf("invalid public key")
	}

	return pubKey[:], nil
}

func getPublicKey(s string) ([]byte, error) {
	var pubKey types.Bytes32
	privKey, err := LookupByLabel(s)

	if err != nil {
		b, err := pubKeyFromString(s)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve public key %s,%v", s, err)
		}
		pubKey.FromBytes(b)
	} else {
		pubKey.FromBytes(privKey[32:])
	}

	return pubKey[:], nil
}

func LookupByAnon(anon string) (privKey []byte, err error) {
	err = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("anon"))
		privKey = b.Get([]byte(anon))
		if len(privKey) == 0 {
			err = fmt.Errorf("valid key not found for %s", anon)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return
}

func LookupByLabel(label string) (asData []byte, err error) {
	err = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("label"))
		asData = b.Get([]byte(label))
		if len(asData) == 0 {
			err = fmt.Errorf("valid key not found for %s", label)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return LookupByPubKey(asData)
}

func LookupByPubKey(pubKey []byte) (asData []byte, err error) {
	err = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("keys"))
		asData = b.Get(pubKey)
		return err
	})
	return
}

func GenerateKey(label string) {

	if _, err := strconv.ParseInt(label, 10, 64); err == nil {
		log.Fatal("label cannot be a number")
	}
	_, err := LookupByLabel(label)
	if err == nil {
		log.Fatal(fmt.Errorf("key already exists for label %s", label))
	}
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("keys"))
		err := b.Put(pubKey, privKey)
		return err
	})

	err = Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("label"))
		err := b.Put([]byte(label), pubKey)
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Public Key %s : %x", label, pubKey)
}

func ListKeyPublic() {

	err := Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("label"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("%s \t\t %x\n", k, v)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func ImportKey(label string, pkhex string) {
	var pk ed25519.PrivateKey

	token, err := hex.DecodeString(pkhex)
	if err != nil {
		log.Fatal(err)
	}

	if len(token) == 32 {
		pk = ed25519.NewKeyFromSeed(token)
	}
	pk = token

	err = Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("keys"))
		err := b.Put([]byte(label), pk)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("{\"label\":\"%s\",\"publicKey\":\"%x\"}\n", label, token[32:])
}

func ExportKey(label string) {
	err := Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("keys"))
		pk := b.Get([]byte(label))
		fmt.Println(hex.EncodeToString(pk))
		fmt.Printf("{\"label\":\"%s\",\"privateKey\":\"%x\",\"publicKey\":\"%x\"}\n", label, pk[:32], pk[32:])
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func GeneratePrivateKey(label string) (privKey []byte, err error) {
	seed, err := lookupSeed(label)

	if err != nil {
		//if private key seed doesn't exist, just create a key
		_, privKey, err = ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}
	} else {
		//if we do have a seed, then create a new key
		masterKey, _ := bip32.NewMasterKey(seed)

		newKey, err := masterKey.NewChildKey(uint32(getKeyCount()))
		if err != nil {
			return nil, err
		}
		privKey = ed25519.NewKeyFromSeed(newKey.Key)
	}
	return
}

func getKeyCount() (count int64) {
	_ = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("label"))
		if b != nil {
			count = b.Tx().Size()
		}

		return nil
	})
	return count
}

func lookupSeed(label string) (seed []byte, err error) {

	err = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("seed"))
		seed = b.Get([]byte(label))
		if len(seed) == 0 {
			err = fmt.Errorf("seed for the label %s doesn't exist", label)
		}
		return err
	})

	return
}

func ImportMneumonic(label string, mnemonic []string) {
	mns := strings.Join(mnemonic, " ")

	if !bip39.IsMnemonicValid(mns) {
		log.Fatal("invalid mnemonic provided")
	}

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mns, "")

	var err error

	err = Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("seed"))
		seed := b.Get([]byte(label))
		if len(seed) != 0 {
			err = fmt.Errorf("seed for the label %s already exists", label)
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	err = Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("seed"))
		if b != nil {
			b.Put([]byte(label), seed)
		} else {
			return fmt.Errorf("DB: %s", err)
		}
		return nil
	})

}
