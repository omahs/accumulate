package network

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2/query"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	client "gitlab.com/accumulatenetwork/accumulate/pkg/client/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/managed"
	"gitlab.com/accumulatenetwork/accumulate/tools/cmd/staking/app"
)

type Network struct {
	client             *client.Client  // Pointer to the client
	params             *app.Parameters // Parameters that drive staking
	Accounts           map[string]int  // Accounts to be queried when getting a block
	Blocks             []*app.Block    // The list of major blocks so far in the protocol
	missingMajorBlocks []int           // Any missing blocks
	period             time.Duration   // the time between
	start              time.Time       // When this app started, used to time block queries
}

var _ app.Accumulate = (*Network)(nil)

func New(server string) (*Network, error) {
	c, err := client.New(server)
	if err != nil {
		return nil, err
	}

	n := new(Network)
	n.client = c
	return n, nil
}

func (n *Network) Debug() { n.client.DebugRequest = true }

func (n *Network) Run() {

	n.period = time.Second * 8                   // Start testing 4 times a second.  We will auto adjust
	if p, err := n.GetParameters(); err != nil { // Try and get new parameters
		n.params = p
	}
	n.start = time.Now() // "Mon, 02 Jan 2006 15:04:05 MST"
	firstTimestamp, _ := time.Parse(time.RFC1123, "Mon, 24 Oct 2022 00:00:00 UTC")
	cnt := 0
	for {
		wait := false
		num := int64(len(n.Blocks))
		b, _ := n.getBlock(num + 1)
		if b == nil {
			fmt.Printf("%d ", cnt)
			cnt++
			time.Sleep(n.period)
			wait = true
			continue
		}
		cnt = 0
		b.Timestamp = firstTimestamp.Add(time.Hour * 12 * time.Duration(len(n.Blocks)-1)) // Get a -1 by calculating
		n.Blocks = append(n.Blocks, b)                                                  //   before this append

		// Calculate a sample time, even if the real world major block time is much faster.
		if num > 3 && wait {
			dt := time.Since(n.start) / time.Duration(num) // Get Duration of this block
			cp := dt / 8                                   // We want to sample so many times between blocks
			p := (9*n.period + cp) / 10                    // Weight current 9 times more than the current reading
			if p > time.Minute*5 {                         // Check at least every 5 minutes
				p = time.Minute * 5
			}
			n.period = time.Duration(p) // Set the duration to p
		}

		fmt.Println(len(n.Blocks))
	}

}

func (n *Network) Init() {

}
func (n *Network) TokensIssued(int64) {}

func (n *Network) GetParameters() (*app.Parameters, error) {
	if n.params == nil {
		n.params = new(app.Parameters)
		n.params.Init()
	}
	return n.params, nil
}

// QueryParameters()
// ToDo: Should be modified to extract the Parameters as a side effect of GetBlock() since
// the parameters at play are set dynamically, i.e. so we use the right parameters based on
// block height.
//
// For now, we will use the set of hard coded parameters, and the QueryParameters isn't used.
func (n *Network) QueryParameters() (*app.Parameters, error) {
	// Get the latest data entry and unmarshal it
	req1 := new(api.DataEntryQuery)
	req1.Url = app.ParametersUrl
	res1, err := n.client.QueryData(context.Background(), req1)
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "query parameters: %w", err)
	}

	b, err := json.Marshal(res1.Data)
	if err != nil {
		return nil, errors.Format(errors.StatusEncodingError, "marshal parameters data entry: %w", err)
	}

	entry, err := protocol.UnmarshalDataEntryJSON(b)
	if err != nil {
		return nil, errors.Format(errors.StatusEncodingError, "unmarshal parameters data entry: %w", err)
	}

	params := new(app.Parameters)
	err = params.UnmarshalBinary(entry.GetData()[0])
	if err != nil {
		return nil, errors.Format(errors.StatusEncodingError, "unmarshal parameters: %w", err)
	}

	n.params = params
	return params, nil
}

func (n *Network) GetTokensIssued() (int64, error) {
	req := new(api.GeneralQuery)
	req.Url = n.params.Account.TokenIssuance
	issuer := new(protocol.TokenIssuer)
	res := new(api.ChainQueryResponse)
	res.Data = issuer
	err := n.client.RequestAPIv2(context.Background(), "query", req, res)
	if err != nil {
		return 0, err
	}

	// TODO: deal with RC3 issued amount, remove for mainnet
	issued := issuer.Issued
	for issued.Cmp(issuer.SupplyLimit) > 0 {
		issued.Sub(&issued, issuer.SupplyLimit)
	}

	return issued.Int64(), nil
}

// GetBlock
// How the Application gets blocks
func (n *Network) GetBlock(index int64, accounts map[string]int) (*app.Block, error) {
	n.Accounts = accounts
	if index < 1 {
		return nil, fmt.Errorf("block numbers are one based")
	}
	if index-1 >= int64(len(n.Blocks)) {
		return nil, nil
	}
	return n.Blocks[index-1], nil
}

// getBlock
// How the Network struct queries the protocol
func (n *Network) getBlock(index int64) (*app.Block, error) {
	// Get the block metadata
	block, err := n.getMajorBlockMetadata(uint64(index))
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "get anchor hash: %w", err)
	}

	block.Transactions = map[[32]byte][]*protocol.Transaction{}

	for k, _ := range n.Accounts {
		block.Transactions[protocol.AccountUrl(k).AccountID32()] = nil
	}

	// Describe the network
	desc, err := n.client.Describe(context.Background())
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "describe: %w", err)
	}

	// For each partition
	for _, part := range desc.Values.Network.Partitions {
		// Get the major block
		major, err := n.getMajorBlock(part.ID, uint64(index))
		if err != nil {
			return nil, errors.Format(errors.StatusUnknownError, "query %s major block %d: %w", part.ID, index, err)
		}

		// Get all the corresponding minor blocks
		minor, err := n.getMinorBlocks(part.ID, major)
		if err != nil {
			return nil, errors.Format(errors.StatusUnknownError, "query %s minor blocks for major block %d: %w", part.ID, index, err)
		}

		if part.Type == protocol.PartitionTypeDirectory {
			if major.MajorBlockTime == nil {
				fmt.Print("Missing BlockTime")
			} else {
				block.Timestamp = *major.MajorBlockTime
			}
		}

		for _, b := range minor {
			for _, txn := range b.Transactions {
				// Ignore pending and failed transactions
				if txn.Status.Code != errors.StatusDelivered {
					continue
				}

				id := txn.Transaction.Header.Principal.AccountID32()
				txns, ok := block.Transactions[id]
				if ok {
					block.Transactions[id] = append(txns, txn.Transaction)
				}
			}
		}
	}

	return block, nil
}

func (n *Network) getMajorBlockMetadata(blockIndex uint64) (*app.Block, error) {
	offset := uint64(len(n.missingMajorBlocks) + 1)
	major := new(protocol.IndexEntry)
	_, err := n.queryChainEntry(major, protocol.DnUrl().JoinPath(protocol.AnchorPool).WithFragment(fmt.Sprintf("chain/major-block/%d", blockIndex-offset)))
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "query major block %d: %w", blockIndex, err)
	}
	if major.BlockIndex < blockIndex {
		panic("This should not be possible")
	}
	if major.BlockIndex > blockIndex {
		fmt.Printf("Major block %d is missing\n", blockIndex)
		n.missingMajorBlocks = append(n.missingMajorBlocks, int(blockIndex))
		block := new(app.Block)
		block.MajorHeight = int64(blockIndex)
		return block, nil
	}

	index := new(protocol.IndexEntry)
	_, err = n.queryChainEntry(index, protocol.DnUrl().JoinPath(protocol.Ledger).WithFragment(fmt.Sprintf("chain/root-index/%d", major.RootIndexIndex)))
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "query root index entry %d: %w", major.RootIndexIndex, err)
	}

	entry, err := n.queryChainEntry(nil, protocol.DnUrl().JoinPath(protocol.Ledger).WithFragment(fmt.Sprintf("chain/root/%d", index.Source)))
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "query root entry %d: %w", index.Source, err)
	}

	ms := new(managed.MerkleState)
	ms.Count = int64(entry.Height)
	ms.Pending = entry.State

	block := new(app.Block)
	block.MajorHeight = int64(blockIndex)
	block.BlockHash = *(*[32]byte)(ms.GetMDRoot())
	block.Timestamp = *major.BlockTime
	return block, nil
}

func (n *Network) queryChainEntry(value any, url *url.URL) (*api.ChainEntry, error) {
	entry := new(api.ChainEntry)
	entry.Value = value
	res := new(api.ChainQueryResponse)
	res.Data = entry
	req := new(api.GeneralQuery)
	req.Url = url
	err := n.client.RequestAPIv2(context.Background(), "query", req, res)
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "query major block: %w", err)
	}
	return entry, nil
}

func (n *Network) getMajorBlock(partition string, index uint64) (*api.MajorQueryResponse, error) {
	// Query
	req := new(api.MajorBlocksQuery)
	req.Url = protocol.PartitionUrl(partition)
	req.Start = index
	req.Count = 1
	resp, err := n.client.QueryMajorBlocks(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, errors.NotFound("major block %d of %s not found", index, partition)
	}

	// Remarshal map to struct
	b, err := json.Marshal(resp.Items[0])
	if err != nil {
		return nil, err
	}
	block := new(api.MajorQueryResponse)
	err = json.Unmarshal(b, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (n *Network) getMinorBlocks(partition string, major *api.MajorQueryResponse) ([]*api.MinorQueryResponse, error) {
	if len(major.MinorBlocks) == 0 {
		return nil, nil
	}

	// Query from the first minor block in the major block to the last
	req := new(api.MinorBlocksQuery)
	req.Url = protocol.PartitionUrl(partition)
	req.Start = major.MinorBlocks[0].BlockIndex
	if req.Start == 1 {
		req.Start++ // Skip Genesis
	}
	req.Count = major.MinorBlocks[len(major.MinorBlocks)-1].BlockIndex - req.Start + 1
	req.BlockFilterMode = query.BlockFilterModeExcludeEmpty
	req.TxFetchMode = query.TxFetchModeExpand
	resp, err := n.client.QueryMinorBlocks(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Remarshal []map to []struct
	var blocks []*api.MinorQueryResponse
	b, err := json.Marshal(resp.Items)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &blocks)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
