package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	cfg "gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/client"
)

var cmdInitDualNode = &cobra.Command{
	Use:   "dual <[partition.network] | [peer bvn url]>",
	Short: "Initialize a dual node using the either a bvn url as a peer, or by specifying the partition.network name and --seed https://seedproxy",
	Run:   initDualNode,
	Args:  cobra.ExactArgs(1),
}

func setFlagsForInit() error {
	var err error
	if flagInitDualNode.ResolvePublicIP && flagInitDualNode.PublicIP != "" {
		return fmt.Errorf("cannot specify both --resolve-public-ip and --public-ip flags")
	}
	if flagInitDualNode.ResolvePublicIP {
		flagInitNode.PublicIP, err = resolvePublicIp()
		if err != nil {
			return fmt.Errorf("cannot resolve public ip address, %v", err)
		}
	} else {
		flagInitNode.PublicIP = flagInitDualNode.PublicIP
	}

	flagInitNode.SkipVersionCheck = flagInitDualNode.SkipVersionCheck
	flagInitNode.GenesisDoc = flagInitDualNode.GenesisDoc
	flagInitNode.SeedProxy = flagInitDualNode.SeedProxy
	flagInitNode.Follower = false
	flagInitNode.NoPrometheus = flagInitDualNode.NoPrometheus
	var listen string
	if flagInitDualNode.ListenIP != "" {
		u, err := url.Parse(flagInitDualNode.ListenIP)
		if err != nil {
			return err
		}
		listen = fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())
	}
	flagInitNode.ListenIP = listen
	return nil
}

func initDualNodeFromSeed(cmd *cobra.Command, args []string) error {
	s := strings.Split(args[0], ".")
	if len(s) != 2 {
		fatalf("network must be in the form of <network-name>.<partition-name>, e.g. mainnet.bvn0")
	}
	partitionName := s[0]
	networkName := s[1]
	if partitionName == "Directory" {
		return fmt.Errorf("cannot specify \"Directory\" partition, please specify a block validator name for init dual node")
	}
	_ = networkName

	err := setFlagsForInit()
	if err != nil {
		return err
	}

	// configure the BVN first so we know how to setup the bvn.
	args = []string{args[0]}

	initNode(cmd, args)

	c, err := finalizeDnn()
	if err != nil {
		return err
	}

	partition, _, err := findInDescribe("", partitionName, &c.Accumulate.Network)
	if err != nil {
		return fmt.Errorf("cannot find partition %s in network configuration, %v", partitionName, err)
	}

	if partition.Type == cfg.NetworkTypeDirectory {
		return fmt.Errorf("network partition of second node configuration must be a block validator. Please specify {network-name}.{bvn-partition-id} first parameter to init dual")
	}

	bvnHost, err := findHealthyNodeOnPartition(partition)
	if err != nil {
		return fmt.Errorf("cannot find a healthy node on partition %s, %v", partitionName, err)
	}

	args = []string{fmt.Sprintf("tcp://%s:%d", bvnHost, partition.BasePort)}

	initNode(cmd, args)

	return finalizeBvnn()
}

func initDualNodeFromPeer(cmd *cobra.Command, args []string) error {
	u, err := url.Parse(args[0])
	check(err)

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		fatalf("cannot resolve host and port %v", args[0])
	}
	bvnHost := u.String()

	_, err = net.LookupIP(host)
	checkf(err, "unknown host %s", u.Hostname())

	bvnBasePort, err := strconv.ParseUint(port, 10, 16)
	checkf(err, "invalid DN port number")
	dnBasePort := bvnBasePort - uint64(cfg.PortOffsetBlockValidator)

	err = setFlagsForInit()
	if err != nil {
		return err
	}

	// configure the directory node
	dnnUrl := fmt.Sprintf("%s://%s:%d", u.Scheme, u.Hostname(), dnBasePort)
	args = []string{dnnUrl}

	initNode(cmd, args)

	_, err = finalizeDnn()
	if err != nil {
		return fmt.Errorf("error finalizing dnn configuration, %v", err)
	}

	args = []string{bvnHost}

	initNode(cmd, args)

	//finalize BVNN
	err = finalizeBvnn()
	if err != nil {
		return err
	}

	return nil
}

func finalizeDnn() (*cfg.Config, error) {
	c, err := cfg.Load(filepath.Join(flagMain.WorkDir, "dnn"))
	if err != nil {
		return nil, err
	}

	//make sure we have a block validator type
	if c.Accumulate.NetworkType != cfg.Directory {
		return nil, fmt.Errorf("expecting directory but received %v", c.Accumulate.NetworkType)
	}

	if flagInit.NoEmptyBlocks {
		c.Consensus.CreateEmptyBlocks = false
	}
	if flagInit.NoWebsite {
		c.Accumulate.Website.Enabled = false
	}

	if len(c.P2P.PersistentPeers) > 0 {
		c.P2P.BootstrapPeers = c.P2P.PersistentPeers
		c.P2P.PersistentPeers = ""
	}

	err = cfg.Store(c)
	if err != nil {
		return nil, fmt.Errorf("cannot store configuration file for node, %v", err)
	}

	return c, nil
}

func finalizeBvnn() error {
	c, err := cfg.Load(filepath.Join(flagMain.WorkDir, "bvnn"))
	if err != nil {
		return fmt.Errorf("cannot load configuration file for node, %v", err)
	}

	if c.Accumulate.NetworkType != cfg.NetworkTypeBlockValidator {
		return fmt.Errorf("network partition of second node configuration must be a block validator. Please specify {network-name}.{bvn-partition-id} first parameter to init dual")
	}

	if flagInit.NoEmptyBlocks {
		c.Consensus.CreateEmptyBlocks = false
	}
	if flagInit.NoWebsite {
		c.Accumulate.Website.Enabled = false
	}

	//in dual mode, the key between bvn and dn is shared.
	//This will be cleaned up when init system is overhauled with AC-1263
	if len(c.P2P.PersistentPeers) > 0 {
		c.P2P.BootstrapPeers = c.P2P.PersistentPeers
		c.P2P.PersistentPeers = ""
	}

	return cfg.Store(c)
}

// initDualNode accumulate `init dual http://ip:bvnport` or `init dual partition.network --seed https://seednode
func initDualNode(cmd *cobra.Command, args []string) {
	var err error
	if flagInitDualNode.SeedProxy != "" {
		err = initDualNodeFromSeed(cmd, args)
	} else {
		err = initDualNodeFromPeer(cmd, args)
	}
	check(err)
}

//
//// initDualNode accumulate init dual partition.network http://ip:bvnport
//func initDualNode(cmd *cobra.Command, args []string) {
//	s := strings.Split(args[0], ".")
//	if len(s) != 2 {
//		fatalf("network must be in the form of <network-name>.<partition-name>, e.g. mainnet.bvn0")
//	}
//	partitionName := s[0]
//	networkName := s[1]
//	if partitionName == "Directory" {
//		fatalf("cannot specify \"Directory\" partition, please specify a bvn for init dual node")
//	}
//	_ = networkName
//
//	u, err := url.Parse(args[1])
//	check(err)
//
//	host := u.Hostname()
//	port := u.Port()
//	if port == "" {
//		fatalf("cannot resolve host and port %v", args[1])
//	}
//
//	addr, err := net.LookupIP(host)
//	checkf(err, "unknown host %s", u.Hostname())
//	netAddr := addr[0].String()
//
//	dnBasePort, err := strconv.ParseUint(port, 10, 16)
//	checkf(err, "invalid DN port number")
//	dnBasePort -= uint64(cfg.PortOffsetBlockValidator)
//
//	flagInitNode.ListenIP = fmt.Sprintf("http://0.0.0.0:%d", dnBasePort)
//	if flagInitDualNode.PublicIP == "" {
//		flagInitNode.PublicIP, err = resolvePublicIp()
//		checkf(err, "cannot resolve public ip address")
//	} else {
//		flagInitNode.PublicIP = flagInitDualNode.PublicIP
//	}
//	flagInitNode.SkipVersionCheck = flagInitDualNode.SkipVersionCheck
//	flagInitNode.GenesisDoc = flagInitDualNode.GenesisDoc
//	flagInitNode.SeedProxy = flagInitDualNode.SeedProxy
//	flagInitNode.Follower = false
//
//	// configure the BVN first so we know how to setup the bvn.
//	args = []string{u.String()}
//
//	initNode(cmd, args)
//
//	c, err := cfg.Load(filepath.Join(flagMain.WorkDir, "dnn"))
//	check(err)
//
//	//make sure we have a block validator type
//	if c.Accumulate.NetworkType != cfg.Directory {
//		fatalf("expecting directory but received %v", c.Accumulate.NetworkType)
//	}
//
//	//now find out what bvn we are on then let
//	_ = netAddr
//
//	if flagInit.NoEmptyBlocks {
//		c.Consensus.CreateEmptyBlocks = false
//	}
//	if flagInit.NoWebsite {
//		c.Accumulate.Website.Enabled = false
//	}
//
//	if len(c.P2P.PersistentPeers) > 0 {
//		c.P2P.BootstrapPeers = c.P2P.PersistentPeers
//		c.P2P.PersistentPeers = ""
//	}
//	dnWebHostUrl, err := url.Parse(c.Accumulate.Website.ListenAddress)
//	checkf(err, "cannot parse website listen address (%v) for node", c.Accumulate.Website.ListenAddress)
//
//	err = cfg.Store(c)
//	checkf(err, "cannot store configuration file for node")
//
//	flagInitNode.ListenIP = fmt.Sprintf("http://0.0.0.0:%v", dnBasePort+cfg.PortOffsetBlockValidator)
//
//	partition, _, err := findInDescribe("", partitionName, &c.Accumulate.Network)
//	checkf(err, "cannot find partition %s in network configuration", partitionName)
//
//	if partition.Type == cfg.NetworkTypeDirectory {
//		fatalf("network partition of second node configuration must be a block validator. Please specify {network-name}.{bvn-partition-id} first parameter to init dual")
//	}
//	bvnHost, err := findHealthyNodeOnPartition(partition)
//	checkf(err, "cannot find a healthy node on partition %s", partitionName)
//
//	args = []string{fmt.Sprintf("tcp://%s:%d", bvnHost, dnBasePort+cfg.PortOffsetBlockValidator)}
//
//	initNode(cmd, args)
//
//	c, err = cfg.Load(filepath.Join(flagMain.WorkDir, "bvnn"))
//
//	checkf(err, "cannot load configuration file for node")
//
//	if flagInit.NoEmptyBlocks {
//		c.Consensus.CreateEmptyBlocks = false
//	}
//	if flagInit.NoWebsite {
//		c.Accumulate.Website.Enabled = false
//	}
//	webPort, err := strconv.ParseUint(dnWebHostUrl.Port(), 10, 16)
//	checkf(err, "invalid port for bvn website (%v)", dnWebHostUrl.Port())
//	c.Accumulate.Website.ListenAddress = fmt.Sprintf("http://%s:%d", dnWebHostUrl.Hostname(), webPort+1)
//
//	//in dual mode, the key between bvn and dn is shared.
//	//This will be cleaned up when init system is overhauled with AC-1263
//
//	if len(c.P2P.PersistentPeers) > 0 {
//		c.P2P.BootstrapPeers = c.P2P.PersistentPeers
//		c.P2P.PersistentPeers = ""
//	}
//
//	err = cfg.Store(c)
//	checkf(err, "cannot store configuration file for node")
//}

func resolvePublicIp() (string, error) {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	ip := struct {
		Query string
	}{}
	err = json.Unmarshal(body, &ip)
	if err != nil {
		return "", err
	}
	return ip.Query, nil
}

func findHealthyNodeOnPartition(partition *cfg.Partition) (string, error) {
	for _, p := range partition.Nodes {
		addr, _, err := resolveAddr(p.Address)
		if err != nil {
			continue
		}

		accClient, err := client.New(fmt.Sprintf("http://%s:%d", addr, partition.BasePort+int64(cfg.PortOffsetAccumulateApi)))
		if err != nil {
			continue
		}
		tmClient, err := rpchttp.New(fmt.Sprintf("tcp://%s:%d", addr, partition.BasePort+int64(cfg.PortOffsetTendermintRpc)))
		if err != nil {
			continue
		}

		_, err = accClient.Describe(context.Background())
		if err != nil {
			continue
		}

		_, err = tmClient.Status(context.Background())
		if err != nil {
			continue
		}
		//if we get here, assume we have a viable node
		return addr, nil
	}
	return "", fmt.Errorf("no viable node found on partition %s", partition.Id)
}
