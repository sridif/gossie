package gossie

import (
    //"fmt"
    //"cassandra"
    enc "encoding/binary"
    //"strings"
    "strconv"
    "os"
)

const (
    _ = iota
    BytesType
    AsciiType
    UTF8Type
    LongType
    IntegerType
    DecimalType
    UUIDType
    BooleanType
    FloatType
    DoubleType
    DateType
    CounterColumnType
)

var (
    ErrorUnsupportedMarshaling = os.NewError("Cannot marshal value")
    ErrorUnsupportedUnmarshaling = os.NewError("Cannot unmarshal value")
)

type TypeDesc int

/*
    to do:

    FloatType
    DoubleType
    IntegerType
    DecimalType
    UUIDType

    float32
    float64
    all uints

    maybe some more (un)marshalings, maybe something better for DateType?

    more error checking, pass along strconv errors
*/

func Marshal(value interface{}, typeDesc TypeDesc) ([]byte, os.Error) {
    switch v := value.(type) {
        case []byte:    return v, nil
        case bool:      return marshalBool(v, typeDesc)
        case int8:      return marshalInt(int64(v), 1, typeDesc)
        case int16:     return marshalInt(int64(v), 2, typeDesc)
        case int:       return marshalInt(int64(v), 4, typeDesc)
        case int32:     return marshalInt(int64(v), 4, typeDesc)
        case int64:     return marshalInt(v, 8, typeDesc)
        case string:    return marshalString(v, typeDesc)
    }
    return nil, ErrorUnsupportedMarshaling
}

func marshalBool(value bool, typeDesc TypeDesc) ([]byte, os.Error) {
    switch typeDesc {
        case BytesType, BooleanType:
            b := make([]byte, 1)
            if value {
                b[0] = 1
            }
            return b, nil

        case AsciiType, UTF8Type:
            b := make([]byte, 1)
            if value {
                b[0] = '1'
            } else {
                b[0] = '0'
            }
            return b, nil

        case LongType:
            b := make([]byte, 8)
            if value {
                b[7] = 1
            }
            return b, nil
    }
    return nil, ErrorUnsupportedMarshaling
}

func marshalInt(value int64, size int, typeDesc TypeDesc) ([]byte, os.Error) {
    switch typeDesc {

        case LongType:
            b := make([]byte, 8)
            enc.BigEndian.PutUint64(b, uint64(value))
            return b, nil

        case BytesType:
            b := make([]byte, 8)
            enc.BigEndian.PutUint64(b, uint64(value))
            return b[len(b)-size:], nil

        case DateType:
            if size != 8 {
                return nil, ErrorUnsupportedMarshaling
            }
            b := make([]byte, 8)
            enc.BigEndian.PutUint64(b, uint64(value))
            return b, nil

        case AsciiType, UTF8Type:
            return marshalString(strconv.Itoa64(value), UTF8Type)
    }
    return nil, ErrorUnsupportedMarshaling
}

func marshalString(value string, typeDesc TypeDesc) ([]byte, os.Error) {
    // let cassandra check the ascii-ness of the []byte
    switch typeDesc {
        case BytesType, AsciiType, UTF8Type:
            return []byte(value), nil

        case LongType:
            i, err := strconv.Atoi64(value)
            if err != nil {
                return nil, err
            }
            return marshalInt(i, 8, LongType)

/* fix this!
        case UUIDType:
            if len(value) != 36 {
                return nil, ErrorUnsupportedMarshaling
            }
            ints := strings.Split(value, "-")
            if len(ints) != 5 {
                return nil, ErrorUnsupportedMarshaling
            }
            b := marshalInt(strconv.Btoi64(ints[0], 16), 4, BytesType)
            b = append(b, marshalInt(strconv.Btoi64(ints[1], 16), 2, BytesType))
            b = append(b, marshalInt(strconv.Btoi64(ints[2], 16), 2, BytesType))
            b = append(b, marshalInt(strconv.Btoi64(ints[3], 16), 2, BytesType))
            b = append(b, marshalInt(strconv.Btoi64(ints[4], 16), 6, BytesType))
            return b, nil
*/

    }
    return nil, ErrorUnsupportedMarshaling
}

type Value interface {
    Bytes() []byte
    SetBytes([]byte)
}

type Bytes string
func (u *Bytes) Bytes() []byte {
    return []byte(string(*u))
}
func (u *Bytes) SetBytes(b []byte)  {
    *u = Bytes(string(b[0:(len(b))]))
}

type Long int64
func (l *Long) Bytes() []byte {
    b := make([]byte, 8)
    enc.BigEndian.PutUint64(b, uint64(*l))
    return b
}
func (l *Long) SetBytes(b []byte)  {
    *l = Long(enc.BigEndian.Uint64(b))
}

func makeTypeDesc(cassType string) TypeDesc {

    // not a simple class type, check for composite and parse it
    /* disable composite support for now...
    if (strings.HasPrefix(cassType, "org.apache.cassandra.db.marshal.CompositeType(")) {
        composite := &compositeTypeDesc{}
        componentsString := cassType[strings.Index(cassType, "(")+1:len(cassType)-1]
        componentsSlice := strings.Split(componentsString, ",")
        components := make([]TypeDesc, 0)
        for _, component := range componentsSlice {
            components = append(components, makeTypeDesc(component))
        }
        composite.components = components
        return composite
    }
    */

    // simple types
    switch cassType {
        case "org.apache.cassandra.db.marshal.BytesType":      return BytesType
        case "org.apache.cassandra.db.marshal.AsciiType":      return AsciiType
        case "org.apache.cassandra.db.marshal.UTF8Type":       return UTF8Type
        case "org.apache.cassandra.db.marshal.LongType":       return LongType
        case "org.apache.cassandra.db.marshal.IntegerType":    return IntegerType
        case "org.apache.cassandra.db.marshal.DecimalType":    return DecimalType
        case "org.apache.cassandra.db.marshal.UUIDType":       return UUIDType
        case "org.apache.cassandra.db.marshal.BooleanType":    return BooleanType
        case "org.apache.cassandra.db.marshal.FloatType":      return FloatType
        case "org.apache.cassandra.db.marshal.DoubleType":     return DoubleType
        case "org.apache.cassandra.db.marshal.DateType":       return DateType
    }

    // not a recognized type
    return BytesType
}