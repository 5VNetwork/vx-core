package serial

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// const V2RayTypeURLHeader = "type.googleapis.com/"

func GetInstanceOf(v *anypb.Any) (proto.Message, error) {
	// ret, err :=
	return v.UnmarshalNew()
	// instance, err := GetInstance(V2TypeFromURL(v.TypeUrl))
	// if err != nil {
	// 	return nil, err
	// }
	// protoMessage := instance.(proto.Message)
	// if err := proto.Unmarshal(v.Value, protoMessage); err != nil {
	// 	return nil, err
	// }
	// return protoMessage, nil
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
	// settings, _ := proto.Marshal(message)
	// return &anypb.Any{
	// 	TypeUrl: V2RayTypeURLHeader + GetMessageType(message),
	// 	Value:   settings,
	// }
}

// // GetMessageType returns the name of this proto Message.
// func GetMessageType(message proto.Message) string {
// 	return proto.MessageName(message)
// }

// // GetInstance creates a new instance of the message with messageType.
// func GetInstance(messageType string) (interface{}, error) {
// 	mType := proto.MessageType(messageType)
// 	if mType == nil || mType.Elem() == nil {
// 		return nil, errors.New("Serial: Unknown type: " + messageType)
// 	}
// 	return reflect.New(mType.Elem()).Interface(), nil
// }

// func V2Type(v *anypb.Any) string {
// 	return V2TypeFromURL(v.TypeUrl)
// }

// func V2TypeFromURL(string2 string) string {
// 	return strings.TrimPrefix(string2, V2RayTypeURLHeader)
// }

// func V2TypeHumanReadable(v *anypb.Any) string {
// 	return v.TypeUrl
// }

// func V2URLFromV2Type(readableType string) string {
// 	return readableType
// }

// const V2RayTypeURLHeader = "type.googleapis.com/"

// func GetInstanceOf(v *anypb.Any) (proto.Message, error) {
// 	instance, err := GetInstance(V2TypeFromURL(v.TypeUrl))
// 	if err != nil {
// 		return nil, err
// 	}
// 	protoMessage := instance.(proto.Message)
// 	if err := proto.Unmarshal(v.Value, protoMessage); err != nil {
// 		return nil, err
// 	}
// 	return protoMessage, nil
// }

// // ToTypedMessage converts a proto Message into TypedMessage.
// func ToTypedMessage(message proto.Message) *anypb.Any {
// 	if message == nil {
// 		return nil
// 	}
// 	settings, _ := proto.Marshal(message)
// 	return &anypb.Any{
// 		TypeUrl: V2RayTypeURLHeader + GetMessageType(message),
// 		Value:   settings,
// 	}
// }

// // GetMessageType returns the name of this proto Message.
// func GetMessageType(message proto.Message) string {
// 	return string(proto.MessageV2(message).ProtoReflect().Descriptor().FullName())
// }

// // GetInstance creates a new instance of the message with messageType.
// func GetInstance(messageType string) (interface{}, error) {
// 	mType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(messageType))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to find message type %s: %w", messageType, err)
// 	}
// 	return mType.New().Interface(), nil
// }

// func V2Type(v *anypb.Any) string {
// 	return V2TypeFromURL(v.TypeUrl)
// }

// func V2TypeFromURL(string2 string) string {
// 	return strings.TrimPrefix(string2, V2RayTypeURLHeader)
// }

// func V2TypeHumanReadable(v *anypb.Any) string {
// 	return v.TypeUrl
// }

// func V2URLFromV2Type(readableType string) string {
// 	return readableType
// }
