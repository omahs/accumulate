package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"

	"github.com/AccumulateNetwork/accumulated/client"
)

var (
	Client = client.NewAPIClient()
	Db     = initDB()
)

var currentUser = func() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}()

var defaultWorkDir = filepath.Join(currentUser.HomeDir, ".accumulate")

// rootCmd represents the base command when called without any subcommands
var rootCmd = func() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "accumulate",
		Short: "CLI for Accumulate Network",
	}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&Client.Server, "server", "s", "http://localhost:35554/v1", "Accumulated server")
	flags.DurationVarP(&Client.Timeout, "timeout", "t", 5*time.Second, "Timeout for all API requests (i.e. 10s, 1m)")
	flags.BoolVarP(&Client.DebugRequest, "debug", "d", false, "Print accumulated API calls")

	return cmd

}()

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize()
}

func initDB() *bolt.DB {
	//make sure working directory exists
	if _, err := os.Stat(defaultWorkDir); os.IsNotExist(err) {
		err := os.MkdirAll(defaultWorkDir, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	wallet := filepath.Join(defaultWorkDir, "wallet.db")
	db, err := bolt.Open(wallet, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("anon"))
		if err != nil {
			return fmt.Errorf("DB: %s", err)
		}
		return nil
	})

	return db
}
