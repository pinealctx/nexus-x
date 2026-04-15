package nxproto

import (
	"encoding/json"
	"sync"

	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const redactedValue = "***"

// sensitiveIndex is a pre-built index of sensitive field numbers per
// message type, scanned once from the proto registry.
type sensitiveIndex struct {
	fields map[protoreflect.FullName][]protoreflect.FieldNumber
	hasAny map[protoreflect.FullName]bool
}

var (
	globalSensitiveIdx  *sensitiveIndex
	sensitiveIdxOnce    sync.Once
	redactedJSONEncoder = protojson.MarshalOptions{EmitUnpopulated: false}
)

// SensitiveIdx returns the global sensitive field index, building it
// once on first call. Safe for concurrent use.
func SensitiveIdx() *sensitiveIndex {
	sensitiveIdxOnce.Do(func() {
		globalSensitiveIdx = buildSensitiveIndex()
	})
	return globalSensitiveIdx
}

func buildSensitiveIndex() *sensitiveIndex {
	idx := &sensitiveIndex{
		fields: make(map[protoreflect.FullName][]protoreflect.FieldNumber),
		hasAny: make(map[protoreflect.FullName]bool),
	}
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		msgs := fd.Messages()
		for i := range msgs.Len() {
			idx.scanMessage(msgs.Get(i))
		}
		return true
	})
	return idx
}

func (idx *sensitiveIndex) scanMessage(md protoreflect.MessageDescriptor) {
	name := md.FullName()
	if _, ok := idx.fields[name]; ok {
		return
	}

	idx.fields[name] = nil

	var sensitive []protoreflect.FieldNumber
	fds := md.Fields()
	for i := range fds.Len() {
		fd := fds.Get(i)
		if checkSensitiveOption(fd) {
			sensitive = append(sensitive, fd.Number())
		}
		if fd.Kind() == protoreflect.MessageKind {
			idx.scanMessage(fd.Message())
		}
	}
	idx.fields[name] = sensitive

	has := len(sensitive) > 0
	if !has {
		for i := range fds.Len() {
			fd := fds.Get(i)
			if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() {
				if idx.hasAny[fd.Message().FullName()] {
					has = true
					break
				}
			}
		}
	}
	idx.hasAny[name] = has

	nested := md.Messages()
	for i := range nested.Len() {
		idx.scanMessage(nested.Get(i))
	}
}

// MarshalRedactedBytes serializes a proto message to JSON bytes with
// sensitive fields replaced by "***". Only clones when sensitive
// fields exist. Returns nil on error.
func MarshalRedactedBytes(msg proto.Message) []byte {
	if msg == nil {
		return nil
	}
	idx := SensitiveIdx()
	if idx.hasAny[msg.ProtoReflect().Descriptor().FullName()] {
		clone := proto.Clone(msg)
		redactFields(idx, clone.ProtoReflect())
		msg = clone
	}
	b, err := redactedJSONEncoder.Marshal(msg)
	if err != nil {
		return []byte("<marshal error>")
	}
	return b
}

// MarshalRedacted serializes a proto message to a JSON string with
// sensitive fields replaced by "***".
func MarshalRedacted(msg proto.Message) string {
	if msg == nil {
		return "<nil>"
	}
	return string(MarshalRedactedBytes(msg))
}

func redactFields(idx *sensitiveIndex, m protoreflect.Message) {
	md := m.Descriptor()
	for _, num := range idx.fields[md.FullName()] {
		fd := md.Fields().ByNumber(num)
		if fd != nil && fd.Kind() == protoreflect.StringKind && m.Has(fd) {
			m.Set(fd, protoreflect.ValueOfString(redactedValue))
		}
	}
	fds := md.Fields()
	for i := range fds.Len() {
		fd := fds.Get(i)
		if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() && m.Has(fd) {
			if idx.hasAny[fd.Message().FullName()] {
				redactFields(idx, m.Mutable(fd).Message())
			}
		}
	}
}

func checkSensitiveOption(fd protoreflect.FieldDescriptor) bool {
	opts := fd.Options()
	if opts == nil {
		return false
	}
	if !proto.HasExtension(opts, sharedv1.E_Sensitive) {
		return false
	}
	return proto.GetExtension(opts, sharedv1.E_Sensitive).(bool)
}

// --- ProtoJSON zap field ---

// ProtoJSON returns a zap.Field that embeds a proto message as raw
// JSON in the log output. Sensitive fields are redacted. Serialization
// is lazy — zero cost when the log level is not enabled.
func ProtoJSON(key string, msg proto.Message) zap.Field {
	if msg == nil {
		return zap.String(key, "<nil>")
	}
	return zap.Any(key, &lazyProtoJSON{msg: msg})
}

type lazyProtoJSON struct{ msg proto.Message }

func (p *lazyProtoJSON) MarshalJSON() ([]byte, error) {
	b := MarshalRedactedBytes(p.msg)
	if b == nil {
		return json.RawMessage("null"), nil
	}
	return b, nil
}
