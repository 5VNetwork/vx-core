package protocol

import "errors"

var ErrProtoNeedMoreData = errors.New("protocol matches, but need more data to complete sniffing")
var ErrNoClue = errors.New("not enough information for making a decision")
