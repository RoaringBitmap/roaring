package roaring

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *arrayContainer) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zbzg uint32
	zbzg, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zbzg > 0 {
		zbzg--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Content":
			var zbai uint32
			zbai, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Content) >= int(zbai) {
				z.Content = (z.Content)[:zbai]
			} else {
				z.Content = make([]uint16, zbai)
			}
			for zxvk := range z.Content {
				z.Content[zxvk], err = dc.ReadUint16()
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *arrayContainer) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Content"
	err = en.Append(0x81, 0xa7, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Content)))
	if err != nil {
		return
	}
	for zxvk := range z.Content {
		err = en.WriteUint16(z.Content[zxvk])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *arrayContainer) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Content"
	o = append(o, 0x81, 0xa7, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Content)))
	for zxvk := range z.Content {
		o = msgp.AppendUint16(o, z.Content[zxvk])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *arrayContainer) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zcmr uint32
	zcmr, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zcmr > 0 {
		zcmr--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Content":
			var zajw uint32
			zajw, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Content) >= int(zajw) {
				z.Content = (z.Content)[:zajw]
			} else {
				z.Content = make([]uint16, zajw)
			}
			for zxvk := range z.Content {
				z.Content[zxvk], bts, err = msgp.ReadUint16Bytes(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *arrayContainer) Msgsize() (s int) {
	s = 1 + 8 + msgp.ArrayHeaderSize + (len(z.Content) * (msgp.Uint16Size))
	return
}