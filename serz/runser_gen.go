package serz

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Interval16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zxvk uint32
	zxvk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zxvk > 0 {
		zxvk--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Start":
			z.Start, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "Last":
			z.Last, err = dc.ReadUint16()
			if err != nil {
				return
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
func (z Interval16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Start"
	err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.Start)
	if err != nil {
		return
	}
	// write "Last"
	err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.Last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Interval16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Start"
	o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint16(o, z.Start)
	// string "Last"
	o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
	o = msgp.AppendUint16(o, z.Last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Interval16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbzg uint32
	zbzg, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbzg > 0 {
		zbzg--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Start":
			z.Start, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "Last":
			z.Last, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
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
func (z Interval16) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint16Size + 5 + msgp.Uint16Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Interval32) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zbai uint32
	zbai, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zbai > 0 {
		zbai--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Start":
			z.Start, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "Last":
			z.Last, err = dc.ReadUint32()
			if err != nil {
				return
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
func (z Interval32) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Start"
	err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.Start)
	if err != nil {
		return
	}
	// write "Last"
	err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.Last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Interval32) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Start"
	o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint32(o, z.Start)
	// string "Last"
	o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
	o = msgp.AppendUint32(o, z.Last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Interval32) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Start":
			z.Start, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "Last":
			z.Last, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
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
func (z Interval32) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint32Size + 5 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RunContainer16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zwht uint32
	zwht, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zwht > 0 {
		zwht--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zhct uint32
			zhct, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zhct) {
				z.Iv = (z.Iv)[:zhct]
			} else {
				z.Iv = make([]Interval16, zhct)
			}
			for zajw := range z.Iv {
				var zcua uint32
				zcua, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zcua > 0 {
					zcua--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zajw].Start, err = dc.ReadUint16()
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zajw].Last, err = dc.ReadUint16()
						if err != nil {
							return
						}
					default:
						err = dc.Skip()
						if err != nil {
							return
						}
					}
				}
			}
		case "Card":
			z.Card, err = dc.ReadInt64()
			if err != nil {
				return
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
func (z *RunContainer16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Iv"
	err = en.Append(0x82, 0xa2, 0x49, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Iv)))
	if err != nil {
		return
	}
	for zajw := range z.Iv {
		// map header, size 2
		// write "Start"
		err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.Iv[zajw].Start)
		if err != nil {
			return
		}
		// write "Last"
		err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.Iv[zajw].Last)
		if err != nil {
			return
		}
	}
	// write "Card"
	err = en.Append(0xa4, 0x43, 0x61, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Card)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RunContainer16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Iv"
	o = append(o, 0x82, 0xa2, 0x49, 0x76)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Iv)))
	for zajw := range z.Iv {
		// map header, size 2
		// string "Start"
		o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint16(o, z.Iv[zajw].Start)
		// string "Last"
		o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint16(o, z.Iv[zajw].Last)
	}
	// string "Card"
	o = append(o, 0xa4, 0x43, 0x61, 0x72, 0x64)
	o = msgp.AppendInt64(o, z.Card)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RunContainer16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zxhx uint32
	zxhx, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zxhx > 0 {
		zxhx--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zlqf uint32
			zlqf, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zlqf) {
				z.Iv = (z.Iv)[:zlqf]
			} else {
				z.Iv = make([]Interval16, zlqf)
			}
			for zajw := range z.Iv {
				var zdaf uint32
				zdaf, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zdaf > 0 {
					zdaf--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zajw].Start, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zajw].Last, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					default:
						bts, err = msgp.Skip(bts)
						if err != nil {
							return
						}
					}
				}
			}
		case "Card":
			z.Card, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
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
func (z *RunContainer16) Msgsize() (s int) {
	s = 1 + 3 + msgp.ArrayHeaderSize + (len(z.Iv) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 5 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RunContainer32) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zjfb uint32
	zjfb, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zjfb > 0 {
		zjfb--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zcxo uint32
			zcxo, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zcxo) {
				z.Iv = (z.Iv)[:zcxo]
			} else {
				z.Iv = make([]Interval32, zcxo)
			}
			for zpks := range z.Iv {
				var zeff uint32
				zeff, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zeff > 0 {
					zeff--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zpks].Start, err = dc.ReadUint32()
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zpks].Last, err = dc.ReadUint32()
						if err != nil {
							return
						}
					default:
						err = dc.Skip()
						if err != nil {
							return
						}
					}
				}
			}
		case "Card":
			z.Card, err = dc.ReadInt64()
			if err != nil {
				return
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
func (z *RunContainer32) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Iv"
	err = en.Append(0x82, 0xa2, 0x49, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Iv)))
	if err != nil {
		return
	}
	for zpks := range z.Iv {
		// map header, size 2
		// write "Start"
		err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint32(z.Iv[zpks].Start)
		if err != nil {
			return
		}
		// write "Last"
		err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint32(z.Iv[zpks].Last)
		if err != nil {
			return
		}
	}
	// write "Card"
	err = en.Append(0xa4, 0x43, 0x61, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Card)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RunContainer32) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Iv"
	o = append(o, 0x82, 0xa2, 0x49, 0x76)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Iv)))
	for zpks := range z.Iv {
		// map header, size 2
		// string "Start"
		o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint32(o, z.Iv[zpks].Start)
		// string "Last"
		o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint32(o, z.Iv[zpks].Last)
	}
	// string "Card"
	o = append(o, 0xa4, 0x43, 0x61, 0x72, 0x64)
	o = msgp.AppendInt64(o, z.Card)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RunContainer32) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zrsw uint32
	zrsw, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zrsw > 0 {
		zrsw--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zxpk uint32
			zxpk, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zxpk) {
				z.Iv = (z.Iv)[:zxpk]
			} else {
				z.Iv = make([]Interval32, zxpk)
			}
			for zpks := range z.Iv {
				var zdnj uint32
				zdnj, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zdnj > 0 {
					zdnj--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zpks].Start, bts, err = msgp.ReadUint32Bytes(bts)
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zpks].Last, bts, err = msgp.ReadUint32Bytes(bts)
						if err != nil {
							return
						}
					default:
						bts, err = msgp.Skip(bts)
						if err != nil {
							return
						}
					}
				}
			}
		case "Card":
			z.Card, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
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
func (z *RunContainer32) Msgsize() (s int) {
	s = 1 + 3 + msgp.ArrayHeaderSize + (len(z.Iv) * (12 + msgp.Uint32Size + msgp.Uint32Size)) + 5 + msgp.Int64Size
	return
}
