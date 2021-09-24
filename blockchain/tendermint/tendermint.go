package tendermint

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	cfg "github.com/AccumulateNetwork/accumulated/config"
	"github.com/AccumulateNetwork/accumulated/networks"
	abcicli "github.com/tendermint/tendermint/abci/client"
	abciserver "github.com/tendermint/tendermint/abci/server"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	coregrpc "github.com/tendermint/tendermint/rpc/grpc"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	nodeDirPerm = 0755
)

func Initialize(shardname string, index int, WorkingDir string) {
	network := networks.Networks[index]

	listenIP := make([]string, len(network.Ip))
	config := make([]*cfg.Config, len(network.Ip))

	for i := range network.Ip {
		listenIP[i] = "tcp://0.0.0.0"
		config[i] = new(cfg.Config)
		config[i].Config = *tmcfg.DefaultConfig()
	}

	err := InitWithConfig(WorkingDir, shardname, network.Name, network.Port, config, network.Ip, listenIP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
	}
}

func InitWithConfig(workDir, shardName, chainID string, port int, config []*cfg.Config, remoteIP []string, listenIP []string) (err error) {
	defer func() {
		if err != nil {
			_ = os.RemoveAll(workDir)
		}
	}()

	fmt.Println("Tendermint Initialize")

	nValidators := len(config)
	genVals := make([]types.GenesisValidator, nValidators)

	for i, config := range config {
		nodeDirName := fmt.Sprintf("Node%d", i)
		nodeDir := path.Join(workDir, nodeDirName)
		config.SetRoot(nodeDir)

		config.Instrumentation.Namespace = shardName

		// config.ProxyApp = fmt.Sprintf("%s:%d", IPs[i], port)
		config.ProxyApp = ""
		config.P2P.ListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port)
		config.RPC.ListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port+1)
		config.RPC.GRPCListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port+2)
		config.Instrumentation.PrometheusListenAddr = fmt.Sprintf(":%d", port)

		config.Consensus.CreateEmptyBlocks = false
		err = os.MkdirAll(path.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			return fmt.Errorf("failed to create config dir: %v", err)
		}

		err = os.MkdirAll(path.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			return fmt.Errorf("failed to create data dir: %v", err)
		}

		if err := initFilesWithConfig(config, &chainID); err != nil {
			return err
		}

		pvKeyFile := path.Join(nodeDir, config.PrivValidatorKey)
		pvStateFile := path.Join(nodeDir, config.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("failed to get public key: %v", err)
		}
		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    nodeDirName,
		}
	}

	// Generate genesis doc from generated validators
	genDoc := &types.GenesisDoc{
		ChainID:         "chain-" + tmrand.Str(6),
		GenesisTime:     tmtime.Now(),
		InitialHeight:   0,
		Validators:      genVals,
		ConsensusParams: types.DefaultConsensusParams(),
	}

	// Write genesis file.
	for _, config := range config {
		if err := genDoc.SaveAs(path.Join(config.RootDir, config.BaseConfig.Genesis)); err != nil {
			return fmt.Errorf("failed to save gen doc: %v", err)
		}
	}

	// Gather persistent peer addresses.
	persistentPeers := make([]string, nValidators)
	for i, config := range config {
		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return fmt.Errorf("failed to load node key: %v", err)
		}
		persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("%s:%d", remoteIP[i], port))
	}

	// Overwrite default config.
	for i, config := range config {
		config.LogLevel = "main:info,state:info,statesync:info,*:error"
		if nValidators > 1 {
			config.P2P.AddrBookStrict = false
			config.P2P.AllowDuplicateIP = true
			config.P2P.PersistentPeers = ""
			for j, peer := range persistentPeers {
				if j != i {
					config.P2P.PersistentPeers += "," + peer
				}
			}
			config.P2P.PersistentPeers = config.P2P.PersistentPeers[1:]
		} else {
			config.P2P.AddrBookStrict = true
			config.P2P.AllowDuplicateIP = false
		}
		config.Moniker = fmt.Sprintf("Node%d", i)

		config.Accumulate.RPC.ListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port+3)
		config.Accumulate.Router.JSONListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port+4)
		config.Accumulate.Router.RESTListenAddress = fmt.Sprintf("%s:%d", listenIP[i], port+5)

		err := cfg.Store(config)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Successfully initialized %v node directories\n", nValidators)
	return nil
}

func initFilesWithConfig(config *cfg.Config, chainid *string) error {

	logger := tmlog.NewNopLogger()

	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if tmos.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			fmt.Printf("Can't load or gen node key\n")
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}
	// genesis file
	genFile := config.GenesisFile()
	if tmos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         *chainid,
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		pubKey, err := pv.GetPubKey()
		if err != nil {
			fmt.Printf("can't get pubkey: %v\n", err)
			return err
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			fmt.Printf("Can't save genFile: %s\n", genFile)
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}

func makeGRPCClient(addr string) (abcicli.Client, error) { //grpccore.BroadcastAPIClient, error) {//abcicli.Client, error) {
	// Start the listener
	socket := addr                 //fmt.Sprintf("unix://%s.sock", addr)
	logger := tmlog.NewNopLogger() //TestingLogger()

	//client := grpccore.StartGRPCClient(addr)
	client := abcicli.NewGRPCClient(socket, true)

	client.SetLogger(logger.With("module", "abci-client"))
	if err := client.Start(); err != nil {
		return nil, err
	}
	return client, nil
}

//
//func makeRPCClient(addr string) rpcclient.ABCIClient {
//}
//

func makeGRPCServer(app abcitypes.Application, name string) (service.Service, error) {
	// Start the listener
	socket := name                 // fmt.Sprintf("unix://%s.sock", name)
	logger := tmlog.NewNopLogger() //TestingLogger()

	gapp := abcitypes.NewGRPCApplication(app)
	server := abciserver.NewGRPCServer(socket, gapp)
	server.SetLogger(logger.With("module", "abci-server"))
	if err := server.Start(); err != nil {
		return nil, err
	}

	return server, nil
}
func WaitForRPC(laddr string) {
	//laddr := GetConfig().RPC.ListenAddress
	client, err := rpcclient.New(laddr)
	if err != nil {
		panic(err)
	}
	result := new(ctypes.ResultStatus)
	for {
		_, err := client.Call(context.Background(), "status", map[string]interface{}{}, result)
		if err == nil {
			return
		}

		// fmt.Println("error", err)
		time.Sleep(time.Millisecond)
	}
}
func GetGRPCClient(grpcAddr string) coregrpc.BroadcastAPIClient {
	return coregrpc.StartGRPCClient(grpcAddr)
}

func GetRPCClient(rpcAddr string) *rpcclient.Client {
	client, _ := rpcclient.New(rpcAddr)
	//b := client.NewRequestBatch()
	//
	//result := new(ctypes.ResultStatus)
	//_, err := client.Call(context.Background(), "status", map[string]interface{}{}, result)
	//b.Call()
	return client
}

func WaitForGRPC(grpcAddr string) {
	client := GetGRPCClient(grpcAddr)
	for {
		_, err := client.Ping(context.Background(), &coregrpc.RequestPing{})
		if err == nil {
			return
		}
	}
}
