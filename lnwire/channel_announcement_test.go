package lnwire

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/roasbeef/btcd/btcec"
	"github.com/roasbeef/btcd/chaincfg/chainhash"
)

func TestChannelAnnoucementEncodeDecode(t *testing.T) {
	ca := &ChannelAnnouncement{
		NodeSig1:       someSig,
		NodeSig2:       someSig,
		ShortChannelID: someShortChannelID,
		BitcoinSig1:    someSig,
		BitcoinSig2:    someSig,
		NodeID1:        pubKey,
		NodeID2:        pubKey,
		BitcoinKey1:    pubKey,
		BitcoinKey2:    pubKey,
	}

	// Next encode the CA message into an empty bytes buffer.
	var b bytes.Buffer
	if err := ca.Encode(&b, 0); err != nil {
		t.Fatalf("unable to encode ChannelAnnouncement: %v", err)
	}

	// Ensure the max payload estimate is correct.
	serializedLength := uint32(b.Len())
	if serializedLength != ca.MaxPayloadLength(0) {
		t.Fatalf("payload length estimate is incorrect: expected %v "+
			"got %v", serializedLength, ca.MaxPayloadLength(0))
	}

	// Deserialize the encoded CA message into a new empty struct.
	ca2 := &ChannelAnnouncement{}
	if err := ca2.Decode(&b, 0); err != nil {
		t.Fatalf("unable to decode ChannelAnnouncement: %v", err)
	}

	// Assert equality of the two instances.
	if !reflect.DeepEqual(ca, ca2) {
		t.Fatalf("encode/decode error messages don't match %#v vs %#v",
			ca, ca2)
	}
}

func TestChannelAnnoucementValidation(t *testing.T) {
	getKeys := func(s string) (*btcec.PrivateKey, *btcec.PublicKey) {
		return btcec.PrivKeyFromBytes(btcec.S256(), []byte(s))
	}

	firstNodePrivKey, firstNodePubKey := getKeys("node-id-1")
	secondNodePrivKey, secondNodePubKey := getKeys("node-id-2")
	firstBitcoinPrivKey, firstBitcoinPubKey := getKeys("bitcoin-key-1")
	secondBitcoinPrivKey, secondBitcoinPubKey := getKeys("bitcoin-key-2")

	hash := chainhash.DoubleHashB(firstNodePubKey.SerializeCompressed())
	firstBitcoinSig, _ := firstBitcoinPrivKey.Sign(hash)

	hash = chainhash.DoubleHashB(secondNodePubKey.SerializeCompressed())
	secondBitcoinSig, _ := secondBitcoinPrivKey.Sign(hash)

	ca := &ChannelAnnouncement{
		ShortChannelID: someShortChannelID,
		BitcoinSig1:    firstBitcoinSig,
		BitcoinSig2:    secondBitcoinSig,
		NodeID1:        firstNodePubKey,
		NodeID2:        secondNodePubKey,
		BitcoinKey1:    firstBitcoinPubKey,
		BitcoinKey2:    secondBitcoinPubKey,
	}

	dataToSign, _ := ca.DataToSign()
	hash = chainhash.DoubleHashB(dataToSign)

	firstNodeSign, _ := firstNodePrivKey.Sign(hash)
	ca.NodeSig1 = firstNodeSign

	secondNodeSign, _ := secondNodePrivKey.Sign(hash)
	ca.NodeSig2 = secondNodeSign

	if err := ca.Validate(); err != nil {
		t.Fatal(err)
	}
}
