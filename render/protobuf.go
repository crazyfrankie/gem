package render

import (
	"net/http"

	"google.golang.org/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Data any
}

var protobufContentType = []string{"application/x-protobuf"}

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (p ProtoBuf) Render(writer http.ResponseWriter) error {
	p.WriteContentType(writer)

	bytes, err := proto.Marshal(p.Data.(proto.Message))
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	return err
}

// WriteContentType (Protobuf) writes custom ContentType.
func (p ProtoBuf) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, protobufContentType)
}
