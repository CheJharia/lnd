package lnwire

import (
	"fmt"
	"io"
	"math"

	"google.golang.org/grpc/codes"
)

// ErrorCode represents the short error code for each of the defined errors
// within the Lightning Network protocol spec.
type ErrorCode uint16

// ToGrpcCode is used to generate gRPC specific code which will be propagated
// to the ln rpc client. This code is used to have more detailed view of what
// goes wrong and also in order to have the ability pragmatically determine
// the error and take specific actions on the client side.
func (e ErrorCode) ToGrpcCode() codes.Code {
	return (codes.Code)(e) + 100
}

const (
	// ErrMaxPendingChannels is returned by remote peer when the number of
	// active pending channels exceeds their maximum policy limit.
	ErrMaxPendingChannels ErrorCode = 1

	// ErrSynchronizingChain is returned by a remote peer that receives a
	// channel update or a funding request while their still syncing to the
	// latest state of the blockchain.
	ErrSynchronizingChain ErrorCode = 2
)

// ErrorData is a set of bytes associated with a particular sent error. A
// receiving node SHOULD only print out data verbatim if the string is composed
// solely of printable ASCII characters. For reference, the printable character
// set includes byte values 32 through 127 inclusive.
type ErrorData []byte

// Error represents a generic error bound to an exact channel. The message
// format is purposefully general in order to allow expression of a wide array
// of possible errors. Each Error message is directed at a particular open
// channel referenced by ChannelPoint.
type Error struct {
	// ChanID references the active channel in which the error occurred
	// within. If the ChanID is all zeroes, then this error applies to the
	// entire established connection.
	ChanID ChannelID

	// Code is the short error code that succinctly identifies the error
	// code. This is similar field is similar to HTTP error codes.
	//
	// TODO(roasbeef): make PR to repo to add error codes, in addition to
	// what's there atm
	Code ErrorCode

	// Data is the attached error data that describes the exact failure
	// which caused the error message to be sent.
	Data ErrorData
}

// NewError creates a new Error message.
func NewError() *Error {
	return &Error{}
}

// A compile time check to ensure Error implements the lnwire.Message
// interface.
var _ Message = (*Error)(nil)

// Decode deserializes a serialized Error message stored in the passed
// io.Reader observing the specified protocol version.
//
// This is part of the lnwire.Message interface.
func (c *Error) Decode(r io.Reader, pver uint32) error {
	return readElements(r,
		&c.ChanID,
		&c.Code,
		&c.Data,
	)
}

// Encode serializes the target Error into the passed io.Writer observing the
// protocol version specified.
//
// This is part of the lnwire.Message interface.
func (c *Error) Encode(w io.Writer, pver uint32) error {
	return writeElements(w,
		c.ChanID,
		c.Code,
		c.Data,
	)
}

// Command returns the integer uniquely identifying an Error message on the
// wire.
//
// This is part of the lnwire.Message interface.
func (c *Error) Command() uint32 {
	return CmdError
}

// MaxPayloadLength returns the maximum allowed payload size for a Error
// complete message observing the specified protocol version.
//
// This is part of the lnwire.Message interface.
func (c *Error) MaxPayloadLength(uint32) uint32 {
	// 32 + 2 + 655326
	return 65536
}

// Validate performs any necessary sanity checks to ensure all fields present
// on the Error are valid.
//
// This is part of the lnwire.Message interface.
func (c *Error) Validate() error {
	if len(c.Data) > math.MaxUint16 {
		return fmt.Errorf("problem string length too long")
	}

	// We're good!
	return nil
}
