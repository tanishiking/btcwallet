package message

import (
	"fmt"

	"github.com/tanishiking/btcwallet/protocol/common"
)

// Reject means reject message.
type Reject struct {
	Message *common.VarStr
	Code    uint8
	Reason  *common.VarStr
	Data    []byte
}

// DecodeReject decode byte slice to Reject.
func DecodeReject(b []byte) (*Reject, error) {
	message, err := common.DecodeVarStr(b)
	if err != nil {
		return nil, err
	}
	length := len(message.Encode())
	code := b[length]
	b = b[length+1:]

	reason, err := common.DecodeVarStr(b)
	if err != nil {
		return nil, err
	}
	reasonLen := len(reason.Encode())
	data := b[reasonLen:]
	return &Reject{
		Message: message,
		Code:    code,
		Reason:  reason,
		Data:    data,
	}, nil
}

// String stringify reject message.
func (reject *Reject) String() string {
	return fmt.Sprintf("ccode: %X, message: %s, reason: %s, data: %v", reject.Code, string(reject.Message.Data), string(reject.Reason.Data), reject.Data)
}
