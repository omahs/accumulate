package main

import (
	"github.com/AccumulateNetwork/accumulated/blockchain/accumulate"
	"github.com/AccumulateNetwork/accumulated/networks"
	"github.com/AccumulateNetwork/accumulated/router"
	"github.com/spf13/viper"

	//	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"os/signal"
	"os/user"
	"syscall"

	"github.com/AccumulateNetwork/accumulated/blockchain/tendermint"
)

var ConfigFile string
var WorkingDir string
var RouterNodeName string
var whichNode int

const DBVCIndex = 0

var (
	BuildTag string = "v0.0.1"
)

//var SpecialModeHeight int64 = 99999999999

func init() {

	usr,err := user.Current()
	if err != nil {
		log.Fatal( err )
		os.Exit(1)
	}

	initdir := path.Join(usr.HomeDir , "/.accumulate" )
	nodeName := "Acadia"

	version := flag.Bool("v", false, "prints the current version")
	flag.StringVar(&initdir, "workingdir", usr.HomeDir +  "/.accumulate", "Path to data directory")
        flag.StringVar(&nodeName, "n", "Acadia", "Node to build configs for")
	node := flag.Int("i", -1, "Which Node are we?  Required (0-n)")

 	wantInit := flag.Bool("init", false, "Initialize")
 	wantDBVC := flag.Bool("dbvc", false, "DBVC Initialize")

	flag.Parse()

	if *version {
		fmt.Printf("Accumulate BVC %s\n",BuildTag)
		os.Exit(0)
	}

	WorkingDir = initdir
	fmt.Printf("Working dir: %v\n", WorkingDir)

	if *wantInit {
	   for j := range router.Networks {
	       if router.Networks[j].Name == nodeName {
	          fmt.Printf("Building configs for %s\n",nodeName)
		  tendermint.Initialize("accumulate.", j, WorkingDir)
		  break;
	    	}
           }
	   os.Exit(0)
	}
	if *wantDBVC {
	   os.Exit(0)
	}
        if *node == -1 {
	   fmt.Printf("Must specify which node we are running, 0-n");
	   os.Exit(0); 
	}
	whichNode = *node
}


func main() {
	nodeDir := fmt.Sprintf("Node%d", whichNode)

	WorkingDir = path.Join(WorkingDir, nodeDir)
	ConfigFile = path.Join(WorkingDir,"/config/config.toml")

	fmt.Printf("%s\n", ConfigFile)

	//First create a router
	viper.SetConfigFile(ConfigFile)
	viper.AddConfigPath(WorkingDir)
	viper.ReadInConfig()
	urlrouter := router.NewRouter(viper.GetString("accumulate.RouterAddress"))

	//Next create a BVC
	accvm, err := accnode.CreateAccumulateBVC(ConfigFile, WorkingDir)
	if err != nil {
		panic(fmt.Errorf("unable to create accumulate BVC"))
	}

	///we really need to open up ports to ALL shards in the system.  Maybe this should be a query to the DBVC blockchain.
	accvmapi, _ := accvm.GetAPIClient()
	urlrouter.AddBVCClient(accvm.GetName(), accvmapi)

	//temporary server for each vm.  will be replaced by url router.
	//go router.Jsonrpcserver2(accvmapi)


	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}
