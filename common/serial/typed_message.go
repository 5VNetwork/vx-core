package serial

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func GetInstanceOf(v *anypb.Any) (proto.Message, error) {
	
	return v.UnmarshalNew()
}

func ToTypedMessages(messages ...proto.Message) []*anypb.Any {
	if len(messages) == 0 {
		return nil
	}
	anyMessages := make([]*anypb.Any, 0, len(messages))
	for _, message := range messages {
		anyMessages = append(anyMessages, ToTypedMessage(message))
	}
	return anyMessages
}

func ToTypedMessage0(message proto.Message) (*anypb.Any, error) {
	if message == nil {
		panic("message is nil")
	}
	a := new(anypb.Any)
	err := a.MarshalFrom(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}
	return a, nil
}

// ToTypedMessage converts a proto Message into TypedMessage.
func ToTypedMessage(message proto.Message) *anypb.Any {
	if message == nil {
		panic("message is nil")
	}
	a := new(anypb.Any)
	err := a.MarshalFrom(message)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal message: %v", err))
	}
	return a
}
