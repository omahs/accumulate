package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dc "github.com/docker/cli/cli/compose/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tendermint/tendermint/libs/log"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/types"
	"gitlab.com/accumulatenetwork/accumulate"
	"gitlab.com/accumulatenetwork/accumulate/config"
	cfg "gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/client"
	"gitlab.com/accumulatenetwork/accumulate/internal/genesis"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/node"
	"gitlab.com/accumulatenetwork/accumulate/networks"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	etcd "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize a named network",
	Run:   initNamedNetwork,
	Args:  cobra.NoArgs,
}

var cmdListNamedNetworkConfig = &cobra.Command{
	Use:   "list <network-name>",
	Short: "Initialize a named network",
	Run:   listNamedNetworkConfig,
	Args:  cobra.ExactArgs(1),
}

var cmdInitDualNode = &cobra.Command{
	Use:   "dual <url|ip> <dn base port> <bvn base port>",
	Short: "Initialize a dual run from seed IP, DN base port, and BVN base port",
	Run:   initDualNode,
	Args:  cobra.ExactArgs(2),
}

var cmdInitNode = &cobra.Command{
	Use:   "node <node nr> <network-name|url>",
	Short: "Initialize a node",
	Run:   initNode,
	Args:  cobra.ExactArgs(2),
}

var cmdInitNetwork = &cobra.Command{
	Use:   "network <network configuration file",
	Short: "Initialize a network",
	Run:   initNetwork,
	Args:  cobra.ExactArgs(1),
}

var cmdInitDevnet = &cobra.Command{
	Use:   "devnet",
	Short: "Initialize a DevNet",
	Run:   initDevNet,
	Args:  cobra.NoArgs,
}

var flagInit struct {
	Net           string
	NoEmptyBlocks bool
	NoWebsite     bool
	Reset         bool
	LogLevels     string
	Etcd          []string
}

var flagInitNode struct {
	GenesisDoc       string
	ListenIP         string
	Follower         bool
	SkipVersionCheck bool
}

var flagInitNodeFromSeed struct {
	GenesisDoc       string
	Follower         bool
	SkipVersionCheck bool
}

var flagInitDevnet struct {
	Name          string
	NumBvns       int
	NumValidators int
	NumFollowers  int
	BasePort      int
	IPs           []string
	Docker        bool
	DockerImage   string
	UseVolumes    bool
	Compose       bool
	DnsSuffix     string
}

var flagInitNetwork struct {
	GenesisDoc     string
	Docker         bool
	DockerImage    string
	UseVolumes     bool
	Compose        bool
	DnsSuffix      string
	FactomBalances string
}

func init() {
	cmdMain.AddCommand(cmdInit)
	cmdInit.AddCommand(cmdInitNode, cmdInitDevnet, cmdInitNetwork, cmdListNamedNetworkConfig, cmdInitDualNode)

	cmdInitNetwork.Flags().StringVar(&flagInitNetwork.GenesisDoc, "genesis-doc", "", "Genesis doc for the target network")
	cmdInitNetwork.Flags().BoolVar(&flagInitNetwork.Docker, "docker", false, "Configure a network that will be deployed with Docker Compose")
	cmdInitNetwork.Flags().StringVar(&flagInitNetwork.DockerImage, "image", "registry.gitlab.com/accumulatenetwork/accumulate", "Docker image name (and tag)")
	cmdInitNetwork.Flags().BoolVar(&flagInitNetwork.UseVolumes, "use-volumes", false, "Use Docker volumes instead of a local directory")
	cmdInitNetwork.Flags().BoolVar(&flagInitNetwork.Compose, "compose", false, "Only write the Docker Compose file, do not write the configuration files")
	cmdInitNetwork.Flags().StringVar(&flagInitNetwork.DnsSuffix, "dns-suffix", "", "DNS suffix to add to hostnames used when initializing dockerized nodes")
	cmdInitNetwork.Flags().StringVar(&flagInitNetwork.FactomBalances, "factom-balances", "", "Factom addresses and balances file path for writing onto the genesis block")

	cmdInit.PersistentFlags().StringVarP(&flagInit.Net, "network", "n", "", "Node to build configs for")
	cmdInit.PersistentFlags().BoolVar(&flagInit.NoEmptyBlocks, "no-empty-blocks", false, "Do not create empty blocks")
	cmdInit.PersistentFlags().BoolVar(&flagInit.NoWebsite, "no-website", false, "Disable website")
	cmdInit.PersistentFlags().BoolVar(&flagInit.Reset, "reset", false, "Delete any existing directories within the working directory")
	cmdInit.PersistentFlags().StringVar(&flagInit.LogLevels, "log-levels", "", "Override the default log levels")
	cmdInit.PersistentFlags().StringSliceVar(&flagInit.Etcd, "etcd", nil, "Use etcd endpoint(s)")
	_ = cmdInit.MarkFlagRequired("network")

	cmdInitNode.Flags().BoolVarP(&flagInitNode.Follower, "follow", "f", false, "Do not participate in voting")
	cmdInitNode.Flags().StringVar(&flagInitNode.GenesisDoc, "genesis-doc", "", "Genesis doc for the target network")
	cmdInitNode.Flags().StringVarP(&flagInitNode.ListenIP, "listen", "l", "", "Address and port to listen on, e.g. tcp://1.2.3.4:5678")
	cmdInitNode.Flags().BoolVar(&flagInitNode.SkipVersionCheck, "skip-version-check", false, "Do not enforce the version check")
	_ = cmdInitNode.MarkFlagRequired("listen")

	cmdInitDualNode.Flags().BoolVarP(&flagInitNodeFromSeed.Follower, "follow", "f", false, "Do not participate in voting")
	cmdInitDualNode.Flags().StringVar(&flagInitNodeFromSeed.GenesisDoc, "genesis-doc", "", "Genesis doc for the target network")
	cmdInitDualNode.Flags().BoolVar(&flagInitNodeFromSeed.SkipVersionCheck, "skip-version-check", false, "Do not enforce the version check")

	cmdInitDevnet.Flags().StringVar(&flagInitDevnet.Name, "name", "DevNet", "Network name")
	cmdInitDevnet.Flags().IntVarP(&flagInitDevnet.NumBvns, "bvns", "b", 2, "Number of block validator networks to configure")
	cmdInitDevnet.Flags().IntVarP(&flagInitDevnet.NumValidators, "validators", "v", 2, "Number of validator nodes per subnet to configure")
	cmdInitDevnet.Flags().IntVarP(&flagInitDevnet.NumFollowers, "followers", "f", 1, "Number of follower nodes per subnet to configure")
	cmdInitDevnet.Flags().IntVar(&flagInitDevnet.BasePort, "port", 26656, "Base port to use for listeners")
	cmdInitDevnet.Flags().StringSliceVar(&flagInitDevnet.IPs, "ip", []string{"127.0.1.1"}, "IP addresses to use or base IP - must not end with .0")
	cmdInitDevnet.Flags().BoolVar(&flagInitDevnet.Docker, "docker", false, "Configure a network that will be deployed with Docker Compose")
	cmdInitDevnet.Flags().StringVar(&flagInitDevnet.DockerImage, "image", "registry.gitlab.com/accumulatenetwork/accumulate", "Docker image name (and tag)")
	cmdInitDevnet.Flags().BoolVar(&flagInitDevnet.UseVolumes, "use-volumes", false, "Use Docker volumes instead of a local directory")
	cmdInitDevnet.Flags().BoolVar(&flagInitDevnet.Compose, "compose", false, "Only write the Docker Compose file, do not write the configuration files")
	cmdInitDevnet.Flags().StringVar(&flagInitDevnet.DnsSuffix, "dns-suffix", "", "DNS suffix to add to hostnames used when initializing dockerized nodes")
}

func listNamedNetworkConfig(*cobra.Command, []string) {
	n := Network{}
	n.Network = "TestNet"
	for _, subnet := range networks.TestNet {
		s := Subnet{}

		s.Name = subnet.Name
		for i := range subnet.Nodes {
			s.Nodes = append(s.Nodes, Node{subnet.Nodes[i].IP, subnet.Nodes[i].Type})
		}
		s.Port = subnet.Port
		s.Type = subnet.Type
		n.Subnet = append(n.Subnet, s)
	}
	data, err := json.Marshal(&n)
	check(err)
	fmt.Printf("%s", data)
}

func initNamedNetwork(*cobra.Command, []string) {

	lclSubnet, err := networks.Resolve(flagInit.Net)
	checkf(err, "--network")

	subnets := make([]config.Subnet, len(lclSubnet.Network))
	i := uint(0)
	for _, s := range lclSubnet.Network {
		bvnNodes := make([]config.Node, len(s.Nodes))
		for i, n := range s.Nodes {
			bvnNodes[i] = config.Node{
				Type:    n.Type,
				Address: fmt.Sprintf("http://%s:%d", n.IP, s.Port),
			}
		}

		subnets[i] = config.Subnet{
			ID:    s.Name,
			Type:  s.Type,
			Nodes: bvnNodes,
		}
		i++
	}

	fmt.Printf("Building config for %s (%s)\n", lclSubnet.Name, lclSubnet.NetworkName)

	listenIP := make([]string, len(lclSubnet.Nodes))
	remoteIP := make([]string, len(lclSubnet.Nodes))
	config := make([]*cfg.Config, len(lclSubnet.Nodes))

	for i, node := range lclSubnet.Nodes {
		listenIP[i] = "0.0.0.0"
		remoteIP[i] = node.IP
		config[i] = cfg.Default(lclSubnet.NetworkName, lclSubnet.Type, node.Type, lclSubnet.Name)
		config[i].Accumulate.Network.LocalAddress = fmt.Sprintf("%s:%d", node.IP, lclSubnet.Port)
		config[i].Accumulate.Network.Subnets = subnets

		if flagInit.LogLevels != "" {
			_, _, err := logging.ParseLogLevel(flagInit.LogLevels, io.Discard)
			checkf(err, "--log-level")
			config[i].LogLevel = flagInit.LogLevels
		}

		if flagInit.NoEmptyBlocks {
			config[i].Consensus.CreateEmptyBlocks = false
		}

		if flagInit.NoWebsite {
			config[i].Accumulate.Website.Enabled = false
		}
	}

	if flagInit.Reset {
		nodeReset()
	}

	_, err = node.Init(node.InitOptions{
		WorkDir:  flagMain.WorkDir,
		Port:     lclSubnet.Port,
		Config:   config,
		RemoteIP: remoteIP,
		ListenIP: listenIP,
		Logger:   newLogger(),
	})
	check(err)
}

func nodeReset() {
	ent, err := os.ReadDir(flagMain.WorkDir)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	check(err)

	for _, ent := range ent {
		if !ent.IsDir() {
			continue
		}

		dir := path.Join(flagMain.WorkDir, ent.Name())
		fmt.Fprintf(os.Stderr, "Deleting %s\n", dir)
		err = os.RemoveAll(dir)
		check(err)
	}
}

func initNode(cmd *cobra.Command, args []string) {
	nodeNr, err := strconv.ParseUint(args[0], 10, 16)
	checkf(err, "invalid node number")

	netAddr, netPort, err := networks.ResolveAddr(args[1], true)
	checkf(err, "invalid network name or URL")

	u, err := url.Parse(flagInitNode.ListenIP)
	checkf(err, "invalid --listen %q", flagInitNode.ListenIP)

	nodePort := 26656
	if u.Port() != "" {
		p, err := strconv.ParseInt(u.Port(), 10, 16)
		if err != nil {
			fatalf("invalid port number %q", u.Port())
		}
		nodePort = int(p)
	}

	accClient, err := client.New(fmt.Sprintf("http://%s:%d", netAddr, netPort+networks.AccApiPortOffset))
	checkf(err, "failed to create API client for %s", args[0])

	tmClient, err := rpchttp.New(fmt.Sprintf("tcp://%s:%d", netAddr, netPort+networks.TmRpcPortOffset))
	checkf(err, "failed to create Tendermint client for %s", args[0])

	version := getVersion(accClient)
	switch {
	case !accumulate.IsVersionKnown() && !version.VersionIsKnown:
		warnf("The version of this executable and %s is unknown. If there is a version mismatch, the node may fail.", args[0])

	case accumulate.Commit != version.Commit:
		if flagInitNode.SkipVersionCheck {
			warnf("This executable is version %s but %s is %s. This may cause the node to fail.", formatVersion(accumulate.Version, accumulate.IsVersionKnown()), args[0], formatVersion(version.Version, version.VersionIsKnown))
		} else {
			fatalf("wrong version: network is %s, we are %s", formatVersion(version.Version, version.VersionIsKnown), formatVersion(accumulate.Version, accumulate.IsVersionKnown()))
		}
	}

	description, err := accClient.Describe(context.Background())
	checkf(err, "failed to get description from %s", args[0])

	var genDoc *types.GenesisDoc
	if cmd.Flag("genesis-doc").Changed {
		genDoc, err = types.GenesisDocFromFile(flagInitNode.GenesisDoc)
		checkf(err, "failed to load genesis doc %q", flagInitNode.GenesisDoc)
	} else {
		warnf("You are fetching the Genesis document from %s! Only do this if you trust %[1]s and your connection to it!", args[0])
		rgen, err := tmClient.Genesis(context.Background())
		checkf(err, "failed to get genesis from %s", args[0])
		genDoc = rgen.Genesis
	}

	status, err := tmClient.Status(context.Background())
	checkf(err, "failed to get status of %s", args[0])

	nodeType := cfg.Validator
	if flagInitNode.Follower {
		nodeType = cfg.Follower
	}
	config := config.Default(description.Network.NetworkName, description.Network.Type, nodeType, description.Network.LocalSubnetID)
	config.P2P.PersistentPeers = fmt.Sprintf("%s@%s:%d", status.NodeInfo.NodeID, netAddr, netPort+networks.TmP2pPortOffset)
	config.Accumulate.Network = description.Network

	if flagInit.Net != "" {
		config.Accumulate.Network.LocalAddress = parseHost(flagInit.Net)
		//need to find the subnet and add the local address to it as well.
		localNode := cfg.Node{Address: flagInit.Net, Type: nodeType}
		for i, s := range config.Accumulate.Network.Subnets {
			if s.ID == config.Accumulate.Network.LocalSubnetID {
				//loop through all the nodes and add persistent peers
				for _, n := range s.Nodes {
					//don't bother to fetch if we already have it.
					nodeHost, _, err := net.SplitHostPort(parseHost(n.Address))
					if err != nil {
						warnf("invalid host from node %s", n.Address)
						continue
					}
					if netAddr != nodeHost {
						tmClient, err := rpchttp.New(fmt.Sprintf("tcp://%s:%d", nodeHost, netPort+networks.TmRpcPortOffset))
						if err != nil {
							warnf("failed to create Tendermint client for %s with error %v", n.Address, err)
							continue
						}

						status, err := tmClient.Status(context.Background())
						if err != nil {
							warnf("failed to get status of %s with error %v", n.Address, err)
							continue
						}

						peers := config.P2P.PersistentPeers
						config.P2P.PersistentPeers = fmt.Sprintf("%s,%s@%s:%d", peers,
							status.NodeInfo.NodeID, nodeHost, netPort+networks.TmP2pPortOffset)
					}
				}

				//prepend the new node to the list
				config.Accumulate.Network.Subnets[i].Nodes = append([]cfg.Node{localNode}, config.Accumulate.Network.Subnets[i].Nodes...)
				break
			}
		}
	}

	if flagInit.LogLevels != "" {
		_, _, err := logging.ParseLogLevel(flagInit.LogLevels, io.Discard)
		checkf(err, "--log-level")
		config.LogLevel = flagInit.LogLevels
	}

	if len(flagInit.Etcd) > 0 {
		config.Accumulate.Storage.Type = cfg.EtcdStorage
		config.Accumulate.Storage.Etcd = new(etcd.Config)
		config.Accumulate.Storage.Etcd.Endpoints = flagInit.Etcd
		config.Accumulate.Storage.Etcd.DialTimeout = 5 * time.Second
	}

	if flagInit.Reset {
		nodeReset()
	}

	_, err = node.Init(node.InitOptions{
		NodeNr:     &nodeNr,
		Version:    1,
		WorkDir:    flagMain.WorkDir,
		Port:       nodePort,
		GenesisDoc: genDoc,
		Config:     []*cfg.Config{config},
		RemoteIP:   []string{u.Hostname()},
		ListenIP:   []string{u.Hostname()},
		Logger:     newLogger(),
	})
	check(err)
}

var baseIP net.IP
var ipCount byte

func nextIP() string {
	if len(flagInitDevnet.IPs) > 1 {
		ipCount++
		if len(flagInitDevnet.IPs) < int(ipCount) {
			fatalf("not enough IPs")
		}
		return flagInitDevnet.IPs[ipCount-1]
	}

	if baseIP == nil {
		baseIP = net.ParseIP(flagInitDevnet.IPs[0])
		if baseIP == nil {
			fatalf("invalid IP: %q", flagInitDevnet.IPs[0])
		}
		if baseIP[15] == 0 {
			fatalf("invalid IP: base IP address must not end with .0")
		}
	}

	ip := make(net.IP, len(baseIP))
	copy(ip, baseIP)
	ip[15] += ipCount
	ipCount++
	return ip.String()
}

func initDevNet(cmd *cobra.Command, _ []string) {
	count := flagInitDevnet.NumValidators + flagInitDevnet.NumFollowers
	verifyInitFlags(cmd, count)

	compose := new(dc.Config)
	compose.Version = "3"
	compose.Services = make([]dc.ServiceConfig, 0, 1+count*(flagInitDevnet.NumBvns+1))
	compose.Volumes = make(map[string]dc.VolumeConfig, 1+count*(flagInitDevnet.NumBvns+1))

	subnets := make([]config.Subnet, flagInitDevnet.NumBvns+1)
	dnConfig := make([]*cfg.Config, count)
	dnRemote := make([]string, count)
	dnListen := make([]string, count)
	initDNs(count, dnConfig, dnRemote, dnListen, compose, subnets)

	bvnConfig := make([][]*cfg.Config, flagInitDevnet.NumBvns)
	bvnRemote := make([][]string, flagInitDevnet.NumBvns)
	bvnListen := make([][]string, flagInitDevnet.NumBvns)
	initBVNs(bvnConfig, count, bvnRemote, bvnListen, compose, subnets)

	handleDNSSuffix(dnRemote, bvnRemote)

	if flagInit.Reset {
		nodeReset()
	}

	if !flagInitDevnet.Compose {
		createInLocalFS(dnConfig, dnRemote, dnListen, bvnConfig, bvnRemote, bvnListen)
		return
	}
	createDockerCompose(cmd, dnRemote, compose)
}

func verifyInitFlags(cmd *cobra.Command, count int) {
	if cmd.Flag("network").Changed {
		fatalf("--network is not applicable to devnet")
	}

	if flagInitDevnet.Compose {
		flagInitDevnet.Docker = true
	}

	if flagInitDevnet.Docker && cmd.Flag("ip").Changed {
		fatalf("--ip and --docker are mutually exclusive")
	}

	if flagInitDevnet.NumBvns == 0 {
		fatalf("Must have at least one block validator network")
	}

	if flagInitDevnet.NumValidators == 0 {
		fatalf("Must have at least one block validator node")
	}

	switch len(flagInitDevnet.IPs) {
	case 1:
		// Generate a sequence from the base IP
	case count * (flagInitDevnet.NumBvns + 1):
		// One IP per node
	default:
		fatalf("not enough IPs - you must specify one base IP or one IP for each node")
	}
}

func initDNs(count int, dnConfig []*cfg.Config, dnRemote []string, dnListen []string, compose *dc.Config, subnets []cfg.Subnet) {
	dnNodes := make([]config.Node, count)
	for i := 0; i < count; i++ {
		nodeType := cfg.Validator
		if i >= flagInitDevnet.NumValidators {
			nodeType = cfg.Follower
		}
		dnConfig[i], dnRemote[i], dnListen[i] = initDevNetNode(cfg.Directory, nodeType, 0, i, compose)
		dnNodes[i] = config.Node{
			Type:    nodeType,
			Address: fmt.Sprintf("http://%s:%d", dnRemote[i], flagInitDevnet.BasePort),
		}
		dnConfig[i].Accumulate.Network.Subnets = subnets
		dnConfig[i].Accumulate.Network.LocalAddress = parseHost(dnNodes[i].Address)
	}

	subnets[0] = config.Subnet{
		ID:    protocol.Directory,
		Type:  config.Directory,
		Nodes: dnNodes,
	}
}

func initBVNs(bvnConfigs [][]*cfg.Config, count int, bvnRemotes [][]string, bvnListen [][]string, compose *dc.Config, subnets []cfg.Subnet) {
	for bvn := range bvnConfigs {
		subnetID := fmt.Sprintf("BVN%d", bvn)
		bvnConfigs[bvn] = make([]*cfg.Config, count)
		bvnRemotes[bvn] = make([]string, count)
		bvnListen[bvn] = make([]string, count)
		bvnNodes := make([]config.Node, count)

		for i := 0; i < count; i++ {
			nodeType := cfg.Validator
			if i >= flagInitDevnet.NumValidators {
				nodeType = cfg.Follower
			}
			bvnConfigs[bvn][i], bvnRemotes[bvn][i], bvnListen[bvn][i] = initDevNetNode(cfg.BlockValidator, nodeType, bvn, i, compose)
			bvnNodes[i] = config.Node{
				Type:    nodeType,
				Address: fmt.Sprintf("http://%s:%d", bvnRemotes[bvn][i], flagInitDevnet.BasePort),
			}
			bvnConfigs[bvn][i].Accumulate.Network.Subnets = subnets
			bvnConfigs[bvn][i].Accumulate.Network.LocalAddress = parseHost(bvnNodes[i].Address)
		}
		subnets[bvn+1] = config.Subnet{
			ID:    subnetID,
			Type:  config.BlockValidator,
			Nodes: bvnNodes,
		}
	}
}

func handleDNSSuffix(dnRemote []string, bvnRemote [][]string) {
	if flagInitDevnet.Docker && flagInitDevnet.DnsSuffix != "" {
		for i := range dnRemote {
			dnRemote[i] += flagInitDevnet.DnsSuffix
		}
		for _, remotes := range bvnRemote {
			for i := range remotes {
				remotes[i] += flagInitDevnet.DnsSuffix
			}
		}
	}
}

func createInLocalFS(dnConfig []*cfg.Config, dnRemote []string, dnListen []string, bvnConfig [][]*cfg.Config, bvnRemote [][]string, bvnListen [][]string) {
	logger := newLogger()
	netValMap := make(genesis.NetworkValidatorMap)
	genInit, err := node.Init(node.InitOptions{
		WorkDir:             filepath.Join(flagMain.WorkDir, "dn"),
		Port:                flagInitDevnet.BasePort,
		Config:              dnConfig,
		RemoteIP:            dnRemote,
		ListenIP:            dnListen,
		NetworkValidatorMap: netValMap,
		Logger:              logger.With("subnet", protocol.Directory),
	})
	check(err)
	genList := []genesis.Bootstrap{genInit}

	for bvn := range bvnConfig {
		bvnConfig, bvnRemote, bvnListen := bvnConfig[bvn], bvnRemote[bvn], bvnListen[bvn]
		genesis, err := node.Init(node.InitOptions{
			WorkDir:             filepath.Join(flagMain.WorkDir, fmt.Sprintf("bvn%d", bvn)),
			Port:                flagInitDevnet.BasePort,
			Config:              bvnConfig,
			RemoteIP:            bvnRemote,
			ListenIP:            bvnListen,
			NetworkValidatorMap: netValMap,
			Logger:              logger.With("subnet", fmt.Sprintf("BVN%d", bvn)),
		})
		check(err)
		if genesis != nil {
			genList = append(genList, genesis)
		}
	}

	// Execute bootstrap after the entire network is known
	for _, genesis := range genList {
		err := genesis.Bootstrap()
		if err != nil {
			panic(fmt.Errorf("could not execute genesis: %v", err))
		}
	}
}

func createDockerCompose(cmd *cobra.Command, dnRemote []string, compose *dc.Config) {
	var svc dc.ServiceConfig
	api := fmt.Sprintf("http://%s:%d/v2", dnRemote[0], flagInitDevnet.BasePort+networks.AccApiPortOffset)
	svc.Name = "tools"
	svc.ContainerName = "devnet-init"
	svc.Image = flagInitDevnet.DockerImage
	svc.Environment = map[string]*string{"ACC_API": &api}
	extras := make(map[string]interface{})
	extras["profiles"] = [...]string{"init"}
	svc.Extras = extras

	svc.Command = dc.ShellCommand{"init", "devnet", "-w", "/nodes", "--docker"}
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "work-dir", "docker", "compose", "reset":
			return
		}

		s := fmt.Sprintf("--%s=%v", flag.Name, flag.Value)
		svc.Command = append(svc.Command, s)
	})

	if flagInitDevnet.UseVolumes {
		svc.Volumes = make([]dc.ServiceVolumeConfig, len(compose.Services))
		for i, node := range compose.Services {
			bits := strings.SplitN(node.Name, "-", 2)
			svc.Volumes[i] = dc.ServiceVolumeConfig{Type: "volume", Source: node.Name, Target: path.Join("/nodes", bits[0], "Node"+bits[1])}
		}
	} else {
		svc.Volumes = []dc.ServiceVolumeConfig{
			{Type: "bind", Source: ".", Target: "/nodes"},
		}
	}

	compose.Services = append(compose.Services, svc)

	dn0svc := compose.Services[0]
	dn0svc.Ports = make([]dc.ServicePortConfig, networks.MaxPortOffset+1)
	for i := range dn0svc.Ports {
		port := uint32(flagInitDevnet.BasePort + i)
		dn0svc.Ports[i] = dc.ServicePortConfig{
			Mode: "host", Protocol: "tcp", Target: port, Published: port,
		}
	}

	f, err := os.Create(filepath.Join(flagMain.WorkDir, "docker-compose.yml"))
	check(err)
	defer f.Close()

	err = yaml.NewEncoder(f).Encode(compose)
	check(err)
}

func initDevNetNode(netType cfg.NetworkType, nodeType cfg.NodeType, bvn, node int, compose *dc.Config) (config *cfg.Config, remote, listen string) {
	if netType == cfg.Directory {
		config = cfg.Default(cfg.DevNet, netType, nodeType, protocol.Directory)
	} else {
		config = cfg.Default(cfg.DevNet, netType, nodeType, fmt.Sprintf("BVN%d", bvn))
	}
	if flagInit.LogLevels != "" {
		_, _, err := logging.ParseLogLevel(flagInit.LogLevels, io.Discard)
		checkf(err, "--log-level")
		config.LogLevel = flagInit.LogLevels
	}

	if flagInit.NoEmptyBlocks {
		config.Consensus.CreateEmptyBlocks = false
	}
	if flagInit.NoWebsite {
		config.Accumulate.Website.Enabled = false
	}

	if len(flagInit.Etcd) > 0 {
		config.Accumulate.Storage.Type = cfg.EtcdStorage
		config.Accumulate.Storage.Etcd = new(etcd.Config)
		config.Accumulate.Storage.Etcd.Endpoints = flagInit.Etcd
		config.Accumulate.Storage.Etcd.DialTimeout = 5 * time.Second
	}

	if !flagInitDevnet.Docker {
		ip := nextIP()
		return config, ip, ip
	}

	var name, dir string
	if netType == cfg.Directory {
		name, dir = fmt.Sprintf("dn-%d", node), fmt.Sprintf("./dn/Node%d", node)
	} else {
		name, dir = fmt.Sprintf("bvn%d-%d", bvn, node), fmt.Sprintf("./bvn%d/Node%d", bvn, node)
	}

	var svc dc.ServiceConfig
	svc.Name = name
	svc.ContainerName = "devnet-" + name
	svc.Image = flagInitDevnet.DockerImage

	if flagInitDevnet.UseVolumes {
		svc.Volumes = []dc.ServiceVolumeConfig{
			{Type: "volume", Source: name, Target: "/node"},
		}
		compose.Volumes[name] = dc.VolumeConfig{}
	} else {
		svc.Volumes = []dc.ServiceVolumeConfig{
			{Type: "bind", Source: dir, Target: "/node"},
		}
	}

	compose.Services = append(compose.Services, svc)
	return config, svc.Name, "0.0.0.0"
}

func parseHost(address string) string {
	nodeUrl, err := url.Parse(address)
	if err == nil {
		return nodeUrl.Host
	}
	return address
}

func newLogger() log.Logger {
	levels := config.DefaultLogLevels
	if flagInit.LogLevels != "" {
		levels = flagInit.LogLevels
	}

	writer, err := logging.NewConsoleWriter("plain")
	check(err)
	level, writer, err := logging.ParseLogLevel(levels, writer)
	check(err)
	logger, err := logging.NewTendermintLogger(zerolog.New(writer), level, false)
	check(err)
	return logger
}
