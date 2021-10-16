package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

type encodeState struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "bencode: unsupported type: " + e.Type.String()
}

type bencodeError struct{ error }

type encOpts struct{}

type encoderFunc func(e *encodeState, v reflect.Value, opts encOpts)

func newEncodeState() *encodeState {
	return &encodeState{}
}

func (e *encodeState) marshal(v interface{}, opt encOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if be, ok := r.(bencodeError); ok {
				err = be
			} else {
				panic(r)
			}
		}
	}()

	e.reflectValue(reflect.ValueOf(v), opt)

	return nil
}

func (e *encodeState) error(err error) {
	panic(bencodeError{err})
}

func (e *encodeState) reflectValue(v reflect.Value, opt encOpts) {
	valueEncoder(v)(e, v, opt)
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func invalidValueEncoder(e *encodeState, v reflect.Value, opts encOpts) {

}

func typeEncoder(v reflect.Type) encoderFunc {
	// todo: cache ???
	f := newTypeEncoder(v)
	return f
}

func newTypeEncoder(t reflect.Type) encoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Struct:
		return newStructEncoder(t)
	// case reflect.Float32:
	// case reflect.Float64:
	default:
		return unsupportedTypeEncoder
	}
}

type structField struct {
}

type structEncoder struct {
	fields []structField
}

func (se structEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	_ = e.WriteByte('d')
	for i := range se.fields {
		f := &se.fields[i]

	}
	_ = e.WriteByte('e')
}

func getStructFields(t reflect.Type) []structField {
	//todo: !
	return []structField{}
}

func newStructEncoder(t reflect.Type) encoderFunc {

}

type arrayEncoder struct {
	elEncoder encoderFunc
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	e := arrayEncoder{typeEncoder(t.Elem())}
	return e.encode
}

func (ae arrayEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	_ = e.WriteByte('l')
	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			_ = e.WriteByte(',')
			ae.elEncoder(e, v.Index(i), opts)
		}
	}
	_ = e.WriteByte('e')
}

type sliceEncoder struct {
	arEncoder encoderFunc
}

func (se sliceEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		_ = e.WriteByte('l')
		_ = e.WriteByte('e')
		return
	}
	se.arEncoder(e, v, opts)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	s := sliceEncoder{newArrayEncoder(t)}
	return s.encode
}

type ptrEncoder struct {
	encoder encoderFunc
}

func (pe ptrEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.error(fmt.Errorf("json: could not encode nil value"))
	}
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	pe := ptrEncoder{typeEncoder(t.Elem())}
	return pe.encode
}

type mapEncoder struct {
	valEncoder encoderFunc
	keyEncoder encoderFunc
}

func (me mapEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		_ = e.WriteByte('d')
		_ = e.WriteByte('e')
		return
	}

	keys := make([]struct {
		kv reflect.Value
		vv reflect.Value
		ks string
	}, v.Len())

	mi := v.MapRange()
	for i := 0; mi.Next(); i++ {
		keys[i].kv = mi.Key()
		keys[i].vv = mi.Value()
		switch keys[i].kv.Kind() {
		case reflect.String:
			keys[i].ks = keys[i].kv.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			keys[i].ks = strconv.FormatInt(keys[i].kv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			keys[i].ks = strconv.FormatUint(keys[i].kv.Uint(), 10)
		default:
			panic("unexpected map key type")
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].ks < keys[j].ks })
	_ = e.WriteByte('d')
	for _, v := range keys {
		me.keyEncoder(e, v.kv, opts)
		me.valEncoder(e, v.vv, opts)
	}
	_ = e.WriteByte('e')
}
func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		return unsupportedTypeEncoder
	}
	me := mapEncoder{
		typeEncoder(t.Elem()),
		typeEncoder(t.Key()),
	}
	return me.encode
}

func interfaceEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		return
	}
	e.reflectValue(v.Elem(), opts)
}

func stringEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	s := v.String()
	uintEncoder(e, reflect.ValueOf(len(s)), opts)
	_ = e.WriteByte(':')
	_, _ = e.WriteString(s)
}

func intEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	buf := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
	_ = e.WriteByte('i')
	_, _ = e.Write(buf)
	_ = e.WriteByte('e')
}

func uintEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	buf := strconv.AppendUint(e.scratch[:0], v.Uint(), 10)
	_ = e.WriteByte('i')
	_, _ = e.Write(buf)
	_ = e.WriteByte('e')
}

func boolEncoder(e *encodeState, v reflect.Value, _ encOpts) {
	if v.Bool() {
		_, _ = e.WriteString("i1e")
	} else {
		_, _ = e.WriteString("i0e")
	}
}

func unsupportedTypeEncoder(e *encodeState, v reflect.Value, _ encOpts) {
	e.error(&UnsupportedTypeError{v.Type()})
}

func Marshal(v interface{}) ([]byte, error) {
	e := newEncodeState()
	err := e.marshal(v, encOpts{})
	if err != nil {
		return nil, err
	}
	buf := append([]byte(nil), e.Bytes()...)
	return buf, nil
}

func Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("not implemeted")
}
