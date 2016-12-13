package roaring

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *RunIterator16) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "rc":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.rc = nil
			} else {
				if z.rc == nil {
					z.rc = new(runContainer16)
				}
				err = z.rc.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "curIndex":
			z.curIndex, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "curPosInIndex":
			z.curPosInIndex, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "curSeq":
			z.curSeq, err = dc.ReadInt64()
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
func (z *RunIterator16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "rc"
	err = en.Append(0x84, 0xa2, 0x72, 0x63)
	if err != nil {
		return err
	}
	if z.rc == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.rc.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "curIndex"
	err = en.Append(0xa8, 0x63, 0x75, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.curIndex)
	if err != nil {
		return
	}
	// write "curPosInIndex"
	err = en.Append(0xad, 0x63, 0x75, 0x72, 0x50, 0x6f, 0x73, 0x49, 0x6e, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.curPosInIndex)
	if err != nil {
		return
	}
	// write "curSeq"
	err = en.Append(0xa6, 0x63, 0x75, 0x72, 0x53, 0x65, 0x71)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.curSeq)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RunIterator16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "rc"
	o = append(o, 0x84, 0xa2, 0x72, 0x63)
	if z.rc == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.rc.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "curIndex"
	o = append(o, 0xa8, 0x63, 0x75, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78)
	o = msgp.AppendInt64(o, z.curIndex)
	// string "curPosInIndex"
	o = append(o, 0xad, 0x63, 0x75, 0x72, 0x50, 0x6f, 0x73, 0x49, 0x6e, 0x49, 0x6e, 0x64, 0x65, 0x78)
	o = msgp.AppendUint16(o, z.curPosInIndex)
	// string "curSeq"
	o = append(o, 0xa6, 0x63, 0x75, 0x72, 0x53, 0x65, 0x71)
	o = msgp.AppendInt64(o, z.curSeq)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RunIterator16) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "rc":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.rc = nil
			} else {
				if z.rc == nil {
					z.rc = new(runContainer16)
				}
				bts, err = z.rc.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "curIndex":
			z.curIndex, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "curPosInIndex":
			z.curPosInIndex, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "curSeq":
			z.curSeq, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z *RunIterator16) Msgsize() (s int) {
	s = 1 + 3
	if z.rc == nil {
		s += msgp.NilSize
	} else {
		s += z.rc.Msgsize()
	}
	s += 9 + msgp.Int64Size + 14 + msgp.Uint16Size + 7 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *addHelper16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zcmr uint32
	zcmr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zcmr > 0 {
		zcmr--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "runstart":
			z.runstart, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "runlen":
			z.runlen, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "actuallyAdded":
			z.actuallyAdded, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "m":
			var zajw uint32
			zajw, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.m) >= int(zajw) {
				z.m = (z.m)[:zajw]
			} else {
				z.m = make([]interval16, zajw)
			}
			for zbai := range z.m {
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
					case "start":
						z.m[zbai].start, err = dc.ReadUint16()
						if err != nil {
							return
						}
					case "last":
						z.m[zbai].last, err = dc.ReadUint16()
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
		case "rc":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.rc = nil
			} else {
				if z.rc == nil {
					z.rc = new(runContainer16)
				}
				err = z.rc.DecodeMsg(dc)
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
func (z *addHelper16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "runstart"
	err = en.Append(0x85, 0xa8, 0x72, 0x75, 0x6e, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.runstart)
	if err != nil {
		return
	}
	// write "runlen"
	err = en.Append(0xa6, 0x72, 0x75, 0x6e, 0x6c, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.runlen)
	if err != nil {
		return
	}
	// write "actuallyAdded"
	err = en.Append(0xad, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x6c, 0x79, 0x41, 0x64, 0x64, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.actuallyAdded)
	if err != nil {
		return
	}
	// write "m"
	err = en.Append(0xa1, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.m)))
	if err != nil {
		return
	}
	for zbai := range z.m {
		// map header, size 2
		// write "start"
		err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.m[zbai].start)
		if err != nil {
			return
		}
		// write "last"
		err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.m[zbai].last)
		if err != nil {
			return
		}
	}
	// write "rc"
	err = en.Append(0xa2, 0x72, 0x63)
	if err != nil {
		return err
	}
	if z.rc == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.rc.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *addHelper16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "runstart"
	o = append(o, 0x85, 0xa8, 0x72, 0x75, 0x6e, 0x73, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint16(o, z.runstart)
	// string "runlen"
	o = append(o, 0xa6, 0x72, 0x75, 0x6e, 0x6c, 0x65, 0x6e)
	o = msgp.AppendUint16(o, z.runlen)
	// string "actuallyAdded"
	o = append(o, 0xad, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x6c, 0x79, 0x41, 0x64, 0x64, 0x65, 0x64)
	o = msgp.AppendUint16(o, z.actuallyAdded)
	// string "m"
	o = append(o, 0xa1, 0x6d)
	o = msgp.AppendArrayHeader(o, uint32(len(z.m)))
	for zbai := range z.m {
		// map header, size 2
		// string "start"
		o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint16(o, z.m[zbai].start)
		// string "last"
		o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint16(o, z.m[zbai].last)
	}
	// string "rc"
	o = append(o, 0xa2, 0x72, 0x63)
	if z.rc == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.rc.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *addHelper16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zhct uint32
	zhct, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "runstart":
			z.runstart, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "runlen":
			z.runlen, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "actuallyAdded":
			z.actuallyAdded, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "m":
			var zcua uint32
			zcua, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.m) >= int(zcua) {
				z.m = (z.m)[:zcua]
			} else {
				z.m = make([]interval16, zcua)
			}
			for zbai := range z.m {
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
					case "start":
						z.m[zbai].start, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					case "last":
						z.m[zbai].last, bts, err = msgp.ReadUint16Bytes(bts)
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
		case "rc":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.rc = nil
			} else {
				if z.rc == nil {
					z.rc = new(runContainer16)
				}
				bts, err = z.rc.UnmarshalMsg(bts)
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
func (z *addHelper16) Msgsize() (s int) {
	s = 1 + 9 + msgp.Uint16Size + 7 + msgp.Uint16Size + 14 + msgp.Uint16Size + 2 + msgp.ArrayHeaderSize + (len(z.m) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 3
	if z.rc == nil {
		s += msgp.NilSize
	} else {
		s += z.rc.Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *interval16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zlqf uint32
	zlqf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zlqf > 0 {
		zlqf--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "start":
			z.start, err = dc.ReadUint16()
			if err != nil {
				return
			}
		case "last":
			z.last, err = dc.ReadUint16()
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
func (z interval16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "start"
	err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.start)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint16(z.last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z interval16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "start"
	o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint16(o, z.start)
	// string "last"
	o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
	o = msgp.AppendUint16(o, z.last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *interval16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
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
		case "start":
			z.start, bts, err = msgp.ReadUint16Bytes(bts)
			if err != nil {
				return
			}
		case "last":
			z.last, bts, err = msgp.ReadUint16Bytes(bts)
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
func (z interval16) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint16Size + 5 + msgp.Uint16Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *runContainer16) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "iv":
			var zcxo uint32
			zcxo, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.iv) >= int(zcxo) {
				z.iv = (z.iv)[:zcxo]
			} else {
				z.iv = make([]interval16, zcxo)
			}
			for zpks := range z.iv {
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
					case "start":
						z.iv[zpks].start, err = dc.ReadUint16()
						if err != nil {
							return
						}
					case "last":
						z.iv[zpks].last, err = dc.ReadUint16()
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
		case "card":
			z.card, err = dc.ReadInt64()
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
func (z *runContainer16) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "iv"
	err = en.Append(0x82, 0xa2, 0x69, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.iv)))
	if err != nil {
		return
	}
	for zpks := range z.iv {
		// map header, size 2
		// write "start"
		err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.iv[zpks].start)
		if err != nil {
			return
		}
		// write "last"
		err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.iv[zpks].last)
		if err != nil {
			return
		}
	}
	// write "card"
	err = en.Append(0xa4, 0x63, 0x61, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.card)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *runContainer16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "iv"
	o = append(o, 0x82, 0xa2, 0x69, 0x76)
	o = msgp.AppendArrayHeader(o, uint32(len(z.iv)))
	for zpks := range z.iv {
		// map header, size 2
		// string "start"
		o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint16(o, z.iv[zpks].start)
		// string "last"
		o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint16(o, z.iv[zpks].last)
	}
	// string "card"
	o = append(o, 0xa4, 0x63, 0x61, 0x72, 0x64)
	o = msgp.AppendInt64(o, z.card)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *runContainer16) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "iv":
			var zxpk uint32
			zxpk, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.iv) >= int(zxpk) {
				z.iv = (z.iv)[:zxpk]
			} else {
				z.iv = make([]interval16, zxpk)
			}
			for zpks := range z.iv {
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
					case "start":
						z.iv[zpks].start, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					case "last":
						z.iv[zpks].last, bts, err = msgp.ReadUint16Bytes(bts)
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
		case "card":
			z.card, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z *runContainer16) Msgsize() (s int) {
	s = 1 + 3 + msgp.ArrayHeaderSize + (len(z.iv) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 5 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *uint16Slice) DecodeMsg(dc *msgp.Reader) (err error) {
	var zkgt uint32
	zkgt, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zkgt) {
		(*z) = (*z)[:zkgt]
	} else {
		(*z) = make(uint16Slice, zkgt)
	}
	for zsnv := range *z {
		(*z)[zsnv], err = dc.ReadUint16()
		if err != nil {
			return
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z uint16Slice) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zema := range z {
		err = en.WriteUint16(z[zema])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z uint16Slice) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zema := range z {
		o = msgp.AppendUint16(o, z[zema])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *uint16Slice) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zqke uint32
	zqke, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zqke) {
		(*z) = (*z)[:zqke]
	} else {
		(*z) = make(uint16Slice, zqke)
	}
	for zpez := range *z {
		(*z)[zpez], bts, err = msgp.ReadUint16Bytes(bts)
		if err != nil {
			return
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z uint16Slice) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (len(z) * (msgp.Uint16Size))
	return
}
