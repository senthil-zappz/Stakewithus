package proof

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/protoio"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

// TMSignature contains all details of validator signature for performing signer recovery for ECDSA
// secp256k1 signature. Note that this struct is written specifically for signature signed on
// Tendermint's precommit data, which includes the block hash and some additional information prepended
// and appended to the block hash. The prepended part (prefix) and the appended part (suffix) are
// different for each signer (including signature size, machine clock, validator index, etc).
type TMSignature struct {
	R                tmbytes.HexBytes `json:"r"`
	S                tmbytes.HexBytes `json:"s"`
	V                uint8            `json:"v"`
	EncodedTimestamp tmbytes.HexBytes `json:"encoded_timestamp"`
}

// TMSignatureEthereum is an Ethereum version of TMSignature for solidity ABI-encoding.
type TMSignatureEthereum struct {
	R                common.Hash
	S                common.Hash
	V                uint8
	EncodedTimestamp []byte
}

func (signature *TMSignature) encodeToEthFormat() TMSignatureEthereum {
	return TMSignatureEthereum{
		R:                common.BytesToHash(signature.R),
		S:                common.BytesToHash(signature.S),
		V:                signature.V,
		EncodedTimestamp: signature.EncodedTimestamp,
	}
}

func recoverETHAddress(msg, sig, signer []byte) ([]byte, uint8, error) {
	for i := uint8(0); i < 2; i++ {
		pubuc, err := crypto.SigToPub(tmhash.Sum(msg), append(sig, byte(i)))
		if err != nil {
			return nil, 0, err
		}
		pub := crypto.CompressPubkey(pubuc)
		var tmp [33]byte

		copy(tmp[:], pub)
		if string(signer) == string(secp256k1.PubKey(tmp[:]).Address()) {
			return crypto.PubkeyToAddress(*pubuc).Bytes(), 27 + i, nil
		}
	}
	return nil, 0, fmt.Errorf("No match address found")
}

func GetPrefix(t tmproto.SignedMsgType, height int64, round int64) ([]byte, error) {
	fmt.Println(t, height, round)
	prefix, err := protoio.MarshalDelimited(
		&tmproto.CanonicalVote{
			Type:   t,
			Height: height,
			Round:  round,
		},
	)
	if err != nil {
		return nil, err
	}
	length := int(prefix[0])
	// prefix should be X + default timestamp that equals to `2a0b088092b8c398feffffff01`, so we trim last 13 bytes
	return prefix[1 : length-12], nil
}

// GetSignaturesAndPrefix returns a list of TMSignature from Tendermint signed header.
func GetSignaturesAndPrefix(info *types.SignedHeader) ([]TMSignature, CommonEncodedVotePart, error) {
	addrs := []string{}
	mapAddrs := map[string]TMSignature{}

	prefix, err := GetPrefix(tmproto.SignedMsgType(info.Commit.Type()), info.Commit.Height, int64(info.Commit.Round))
	if err != nil {
		return nil, CommonEncodedVotePart{}, err
	}
	// Append with 4 fixed bytes
	// 34 is a key for the CanonicalBlockID ( 34 == (4 << 3) | 2 )
	// 72 is the length of the CanonicalBlockID message ( 72 == (1+1 + 32) + (1+1 + (1 + 1) + (1+1 + 32)) )
	// 10 is a key for the blockHash ( 10 == (1 << 3) | 2 )
	// 32 is the length of the blockHash
	prefix = append(prefix, []byte{34, 72, 10, 32}...)

	suffix, err := protoio.MarshalDelimited(
		&tmproto.CanonicalPartSetHeader{
			Total: info.Commit.BlockID.PartSetHeader.Total,
			Hash:  info.Commit.BlockID.PartSetHeader.Hash,
		},
	)
	if err != nil {
		return nil, CommonEncodedVotePart{}, err
	}
	// Append with 1 fixed byte
	// 18 is a key for the CanonicalPartSetHeader ( 18 == (2 << 3) | 2 )
	suffix = append([]byte{18}, suffix...)

	commonVote := CommonEncodedVotePart{SignedDataPrefix: prefix, SignedDataSuffix: suffix}

	commonPart := append(commonVote.SignedDataPrefix, info.Commit.BlockID.Hash...)
	commonPart = append(commonPart, commonVote.SignedDataSuffix...)

	chainIDBytes := []byte(info.ChainID)
	encodedChainIDConstant := append([]byte{50, uint8(len(chainIDBytes))}, chainIDBytes...)

	for _, vote := range info.Commit.Signatures {
		if !vote.ForBlock() {
			continue
		}

		encodedTimestamp := encodeTime(vote.Timestamp)

		msg := append(commonPart, []byte{42, uint8(len(encodedTimestamp))}...)
		msg = append(msg, encodedTimestamp...)
		msg = append(msg, encodedChainIDConstant...)
		msg = append([]byte{uint8(len(msg))}, msg...)

		addr, v, err := recoverETHAddress(msg, vote.Signature, vote.ValidatorAddress)
		if err != nil {
			return nil, CommonEncodedVotePart{}, err
		}
		addrs = append(addrs, string(addr))
		mapAddrs[string(addr)] = TMSignature{
			vote.Signature[:32],
			vote.Signature[32:],
			v,
			encodedTimestamp,
		}
	}
	if len(addrs) == 0 {
		return nil, CommonEncodedVotePart{}, fmt.Errorf("No valid precommit")
	}

	signatures := make([]TMSignature, len(addrs))
	sort.Strings(addrs)
	for i, addr := range addrs {
		signatures[i] = mapAddrs[addr]
	}

	return signatures, commonVote, nil
}
