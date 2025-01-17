package ibft

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/0xPolygon/minimal/blockchain"
	"github.com/0xPolygon/minimal/chain"
	"github.com/0xPolygon/minimal/consensus"
	"github.com/0xPolygon/minimal/crypto"
	"github.com/0xPolygon/minimal/types"
	"github.com/stretchr/testify/assert"
)

type testerAccount struct {
	alias string
	priv  *ecdsa.PrivateKey
}

func (t *testerAccount) Address() types.Address {
	return crypto.PubKeyToAddress(&t.priv.PublicKey)
}

func (t *testerAccount) sign(h *types.Header) *types.Header {
	h, _ = writeSeal(t.priv, h)
	return h
}

type testerAccountPool struct {
	accounts []*testerAccount
}

func newTesterAccountPool(num ...int) *testerAccountPool {
	t := &testerAccountPool{
		accounts: []*testerAccount{},
	}
	if len(num) == 1 {
		for i := 0; i < num[0]; i++ {
			key, _ := crypto.GenerateKey()
			t.accounts = append(t.accounts, &testerAccount{
				alias: strconv.Itoa(i),
				priv:  key,
			})
		}
	}
	return t
}

func (ap *testerAccountPool) add(accounts ...string) {
	for _, account := range accounts {
		if acct := ap.get(account); acct != nil {
			continue
		}
		priv, err := crypto.GenerateKey()
		if err != nil {
			panic("BUG: Failed to generate crypto key")
		}
		ap.accounts = append(ap.accounts, &testerAccount{
			alias: account,
			priv:  priv,
		})
	}
}

func (ap *testerAccountPool) genesis() *chain.Genesis {
	genesis := &types.Header{
		MixHash: IstanbulDigest,
	}
	putIbftExtraValidators(genesis, ap.ValidatorSet())
	genesis.ComputeHash()

	c := &chain.Genesis{
		Mixhash:   genesis.MixHash,
		ExtraData: genesis.ExtraData,
	}
	return c
}

func (ap *testerAccountPool) get(name string) *testerAccount {
	for _, i := range ap.accounts {
		if i.alias == name {
			return i
		}
	}
	return nil
}

func (ap *testerAccountPool) ValidatorSet() ValidatorSet {
	v := ValidatorSet{}
	for _, i := range ap.accounts {
		v = append(v, i.Address())
	}
	return v
}

func TestSnapshot_ProcessHeaders(t *testing.T) {
	type mockVote struct {
		validator string
		voted     string
		auth      bool
	}
	mine := func(validator string) mockVote {
		return mockVote{validator: validator}
	}
	vote := func(validator, voted string, auth bool) mockVote {
		return mockVote{
			validator: validator,
			voted:     voted,
			auth:      auth,
		}
	}

	type mockSnapshot struct {
		validators []string
		votes      []mockVote
	}
	type mockHeader struct {
		action   mockVote
		snapshot *mockSnapshot
	}

	var cases = []struct {
		epoch      uint64
		validators []string
		headers    []mockHeader
	}{
		{
			// single validator casts no vote
			validators: []string{
				"A",
			},
			headers: []mockHeader{
				{
					action: mine("A"),
					snapshot: &mockSnapshot{
						validators: []string{"A"},
					},
				},
			},
		},
		{
			// single validator votes to add two peers
			validators: []string{"A"},
			headers: []mockHeader{
				{
					// one vote from A is enough to promote B.
					// the vote is not even shown on the result
					action: vote("A", "B", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
					},
				},
				{
					action: mine("B"),
				},
				{
					// one vote from A is NOT enough to promote C
					// since now B is also a validator
					action: vote("A", "C", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
						votes: []mockVote{
							vote("A", "C", true),
						},
					},
				},
			},
		},
		{
			// single validator dropping himself
			validators: []string{"A"},
			headers: []mockHeader{
				{
					action: vote("A", "A", false),
					snapshot: &mockSnapshot{
						validators: []string{},
					},
				},
			},
		},
		{
			// two validators, dropping requires consensus
			validators: []string{"A", "B"},
			headers: []mockHeader{
				{
					action: vote("A", "B", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
						votes: []mockVote{
							vote("A", "B", false),
						},
					},
				},
				{
					action: vote("B", "B", false),
					snapshot: &mockSnapshot{
						validators: []string{"A"},
					},
				},
			},
		},
		{
			// adding votes are only counted once per validator and target
			validators: []string{"A", "B"},
			headers: []mockHeader{
				{
					action: vote("A", "C", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
						votes: []mockVote{
							vote("A", "C", true),
						},
					},
				},
				{
					action: vote("A", "C", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
						votes: []mockVote{
							vote("A", "C", true),
						},
					},
				},
			},
		},
		{
			// delete votes are only counted once per validator and target
			validators: []string{"A", "B", "C"},
			headers: []mockHeader{
				{
					action: vote("A", "C", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "C", false),
						},
					},
				},
				{
					action: vote("A", "C", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "C", false),
						},
					},
				},
			},
		},
		{
			// multiple (add, delete) votes are possible
			validators: []string{"A", "B", "C"},
			headers: []mockHeader{
				{
					action: vote("A", "D", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "D", true),
						},
					},
				},
				{
					action: vote("A", "E", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "D", true),
							vote("A", "E", true),
						},
					},
				},
				{
					action: vote("A", "B", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "D", true),
							vote("A", "E", true),
							vote("A", "B", false),
						},
					},
				},
			},
		},
		{
			// votes from desauthorized nodes are discarded inmediately
			validators: []string{"A", "B", "C"},
			headers: []mockHeader{
				// validator C makes two votes (add and delete)
				{
					action: vote("C", "D", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("C", "D", true),
						},
					},
				},
				{
					action: vote("C", "B", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("C", "D", true),
							vote("C", "B", false),
						},
					},
				},
				// A and B remove C
				{
					action: vote("A", "C", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("C", "D", true),
							vote("C", "B", false),
							vote("A", "C", false),
						},
					},
				},
				// B vote is enough to discard C and clean all the votes
				{
					action: vote("B", "C", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B"},
						votes:      []mockVote{},
					},
				},
			},
		},
		{
			// epoch transition resets all votes
			epoch:      3,
			validators: []string{"A", "B", "C"},
			headers: []mockHeader{
				{
					// block 1
					action: vote("A", "D", true),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "D", true),
						},
					},
				},
				{
					// block 2
					action: vote("B", "C", false),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes: []mockVote{
							vote("A", "D", true),
							vote("B", "C", false),
						},
					},
				},
				{
					// block 3 (do not vote)
					action: mine("B"),
					snapshot: &mockSnapshot{
						validators: []string{"A", "B", "C"},
						votes:      []mockVote{},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			epoch := c.epoch
			if epoch == 0 {
				epoch = 1000
			}

			pool := newTesterAccountPool()
			pool.add(c.validators...)

			genesis := pool.genesis()

			// create votes
			parentHash := types.Hash{}

			headers := []*types.Header{}
			for num, header := range c.headers {
				v := header.action
				pool.add(v.validator)

				h := &types.Header{
					Number:     uint64(num + 1),
					ParentHash: parentHash,
					Miner:      types.ZeroAddress,
					MixHash:    IstanbulDigest,
					ExtraData:  genesis.ExtraData,
				}
				if v.voted != "" {
					// if voted is empty, we are just creating a new block
					// without votes
					pool.add(v.voted)
					h.Miner = pool.get(v.voted).Address()
				}
				if v.auth {
					// add auth to the vote
					h.Nonce = nonceAuthVote
				} else {
					h.Nonce = nonceDropVote
				}

				// sign the vote
				h = pool.get(v.validator).sign(h)
				h.ComputeHash()

				parentHash = h.Hash
				headers = append(headers, h)
			}

			// process the headers independently
			ibft := &Ibft{
				epochSize:  epoch,
				blockchain: blockchain.TestBlockchain(t, genesis),
				config:     &consensus.Config{},
			}
			assert.NoError(t, ibft.setupSnapshot())
			for indx, header := range headers {
				if err := ibft.processHeaders([]*types.Header{header}); err != nil {
					t.Fatal(err)
				}

				// get latest snapshot
				snap, err := ibft.getSnapshot(header.Number)
				assert.NoError(t, err)

				result := c.headers[indx].snapshot
				if result != nil {
					resSnap := &Snapshot{
						Votes: []*Vote{},
						Set:   ValidatorSet{},
					}
					// check validators
					for _, i := range result.validators {
						resSnap.Set.Add(pool.get(i).Address())
					}
					// build result votes
					for _, v := range result.votes {
						resSnap.Votes = append(resSnap.Votes, &Vote{
							Validator: pool.get(v.validator).Address(),
							Address:   pool.get(v.voted).Address(),
							Authorize: v.auth,
						})
					}
					if !resSnap.Equal(snap) {
						fmt.Println("-- wrong result --")
						fmt.Println(resSnap.Set)
						fmt.Println(snap.Set)
						fmt.Println(resSnap.Votes)
						fmt.Println(snap.Votes)
						t.Fatal("bad")
					}
				}
			}

			// check the metadata
			meta, err := ibft.getSnapshotMetadata()
			assert.NoError(t, err)

			if meta.LastBlock != headers[len(headers)-1].Number {
				t.Fatal("incorrect meta")
			}

			// Process headers all at the same time should have the same result
			ibft1 := &Ibft{
				epochSize:  epoch,
				blockchain: blockchain.TestBlockchain(t, genesis),
				config:     &consensus.Config{},
			}
			assert.NoError(t, ibft1.setupSnapshot())
			if err := ibft1.processHeaders(headers); err != nil {
				t.Fatal(err)
			}

			// from 0 to last header check that all the snapshots match
			for i := uint64(0); i < headers[len(headers)-1].Number; i++ {
				snap0, err := ibft.getSnapshot(i)
				assert.NoError(t, err)

				snap1, err := ibft1.getSnapshot(i)
				assert.NoError(t, err)

				if !snap0.Equal(snap1) {
					t.Fatal("bad")
				}
			}
		})
	}
}

func TestSnapshot_PurgeSnapshots(t *testing.T) {
	pool := newTesterAccountPool()
	pool.add("a", "b", "c")

	genesis := pool.genesis()
	ibft1 := &Ibft{
		epochSize:  10,
		blockchain: blockchain.TestBlockchain(t, genesis),
		config:     &consensus.Config{},
	}
	assert.NoError(t, ibft1.setupSnapshot())

	// write a header that creates a snapshot
	headers := []*types.Header{}
	for i := 1; i < 51; i++ {
		id := strconv.Itoa(i)
		pool.add(id)

		h := &types.Header{
			Number:     uint64(i),
			ParentHash: ibft1.blockchain.Header().Hash,
			Miner:      types.ZeroAddress,
			MixHash:    IstanbulDigest,
			ExtraData:  genesis.ExtraData,
		}

		h.Miner = pool.get(id).Address()
		h.Nonce = nonceAuthVote

		h = pool.get("a").sign(h)
		h.ComputeHash()
		headers = append(headers, h)
	}

	err := ibft1.processHeaders(headers)
	assert.NoError(t, err)

	assert.Equal(t, len(ibft1.store.list), 21)
}

func TestSnapshot_Store_SaveLoad(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "snapshot-store")
	assert.NoError(t, err)

	store0 := newSnapshotStore()
	for i := 0; i < 10; i++ {
		store0.add(&Snapshot{
			Number: uint64(i),
		})
	}
	assert.NoError(t, store0.saveToPath(tmpDir))

	store1 := newSnapshotStore()
	assert.NoError(t, store1.loadFromPath(tmpDir))

	assert.Equal(t, store0, store1)
}

func TestSnapshot_Store_Find(t *testing.T) {
	store := newSnapshotStore()

	for i := 0; i <= 100; i++ {
		if i%10 == 0 {
			store.add(&Snapshot{
				Number: uint64(i),
			})
		}
	}

	check := func(num, expected uint64) {
		assert.Equal(t, store.find(num).Number, expected)
	}

	check(0, 0)
	check(19, 10)
	check(20, 20)
	check(21, 20)
	check(1000, 100)
}
