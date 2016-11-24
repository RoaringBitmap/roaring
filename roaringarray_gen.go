package roaring

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *containerSerz) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "t":
			{
				var zbzg int16
				zbzg, err = dc.ReadInt16()
				z.T = contype(zbzg)
			}
			if err != nil {
				return
			}
		case "r":
			err = z.R.DecodeMsg(dc)
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
func (z *containerSerz) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "t"
	err = en.Append(0x82, 0xa1, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt16(int16(z.T))
	if err != nil {
		return
	}
	// write "r"
	err = en.Append(0xa1, 0x72)
	if err != nil {
		return err
	}
	err = z.R.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *containerSerz) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "t"
	o = append(o, 0x82, 0xa1, 0x74)
	o = msgp.AppendInt16(o, int16(z.T))
	// string "r"
	o = append(o, 0xa1, 0x72)
	o, err = z.R.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *containerSerz) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbai uint32
	zbai, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbai > 0 {
		zbai--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "t":
			{
				var zcmr int16
				zcmr, bts, err = msgp.ReadInt16Bytes(bts)
				z.T = contype(zcmr)
			}
			if err != nil {
				return
			}
		case "r":
			bts, err = z.R.UnmarshalMsg(bts)
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
func (z *containerSerz) Msgsize() (s int) {
	s = 1 + 2 + msgp.Int16Size + 2 + z.R.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *contype) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zajw int16
		zajw, err = dc.ReadInt16()
		(*z) = contype(zajw)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z contype) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt16(int16(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z contype) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt16(o, int16(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *contype) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zwht int16
		zwht, bts, err = msgp.ReadInt16Bytes(bts)
		(*z) = contype(zwht)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z contype) Msgsize() (s int) {
	s = msgp.Int16Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *roaringArray) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Keys":
			var zdaf uint32
			zdaf, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Keys) >= int(zdaf) {
				z.Keys = (z.Keys)[:zdaf]
			} else {
				z.Keys = make([]uint16, zdaf)
			}
			for zhct := range z.Keys {
				z.Keys[zhct], err = dc.ReadUint16()
				if err != nil {
					return
				}
			}
		case "NeedCopyOnWrite":
			var zpks uint32
			zpks, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.NeedCopyOnWrite) >= int(zpks) {
				z.NeedCopyOnWrite = (z.NeedCopyOnWrite)[:zpks]
			} else {
				z.NeedCopyOnWrite = make([]bool, zpks)
			}
			for zcua := range z.NeedCopyOnWrite {
				z.NeedCopyOnWrite[zcua], err = dc.ReadBool()
				if err != nil {
					return
				}
			}
		case "CopyOnWrite":
			z.CopyOnWrite, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "Conserz":
			var zjfb uint32
			zjfb, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Conserz) >= int(zjfb) {
				z.Conserz = (z.Conserz)[:zjfb]
			} else {
				z.Conserz = make([]containerSerz, zjfb)
			}
			for zxhx := range z.Conserz {
				var zcxo uint32
				zcxo, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zcxo > 0 {
					zcxo--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "t":
						{
							var zeff int16
							zeff, err = dc.ReadInt16()
							z.Conserz[zxhx].T = contype(zeff)
						}
						if err != nil {
							return
						}
					case "r":
						err = z.Conserz[zxhx].R.DecodeMsg(dc)
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
func (z *roaringArray) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "Keys"
	err = en.Append(0x84, 0xa4, 0x4b, 0x65, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Keys)))
	if err != nil {
		return
	}
	for zhct := range z.Keys {
		err = en.WriteUint16(z.Keys[zhct])
		if err != nil {
			return
		}
	}
	// write "NeedCopyOnWrite"
	err = en.Append(0xaf, 0x4e, 0x65, 0x65, 0x64, 0x43, 0x6f, 0x70, 0x79, 0x4f, 0x6e, 0x57, 0x72, 0x69, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.NeedCopyOnWrite)))
	if err != nil {
		return
	}
	for zcua := range z.NeedCopyOnWrite {
		err = en.WriteBool(z.NeedCopyOnWrite[zcua])
		if err != nil {
			return
		}
	}
	// write "CopyOnWrite"
	err = en.Append(0xab, 0x43, 0x6f, 0x70, 0x79, 0x4f, 0x6e, 0x57, 0x72, 0x69, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.CopyOnWrite)
	if err != nil {
		return
	}
	// write "Conserz"
	err = en.Append(0xa7, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x72, 0x7a)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Conserz)))
	if err != nil {
		return
	}
	for zxhx := range z.Conserz {
		// map header, size 2
		// write "t"
		err = en.Append(0x82, 0xa1, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteInt16(int16(z.Conserz[zxhx].T))
		if err != nil {
			return
		}
		// write "r"
		err = en.Append(0xa1, 0x72)
		if err != nil {
			return err
		}
		err = z.Conserz[zxhx].R.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *roaringArray) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Keys"
	o = append(o, 0x84, 0xa4, 0x4b, 0x65, 0x79, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Keys)))
	for zhct := range z.Keys {
		o = msgp.AppendUint16(o, z.Keys[zhct])
	}
	// string "NeedCopyOnWrite"
	o = append(o, 0xaf, 0x4e, 0x65, 0x65, 0x64, 0x43, 0x6f, 0x70, 0x79, 0x4f, 0x6e, 0x57, 0x72, 0x69, 0x74, 0x65)
	o = msgp.AppendArrayHeader(o, uint32(len(z.NeedCopyOnWrite)))
	for zcua := range z.NeedCopyOnWrite {
		o = msgp.AppendBool(o, z.NeedCopyOnWrite[zcua])
	}
	// string "CopyOnWrite"
	o = append(o, 0xab, 0x43, 0x6f, 0x70, 0x79, 0x4f, 0x6e, 0x57, 0x72, 0x69, 0x74, 0x65)
	o = msgp.AppendBool(o, z.CopyOnWrite)
	// string "Conserz"
	o = append(o, 0xa7, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x72, 0x7a)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Conserz)))
	for zxhx := range z.Conserz {
		// map header, size 2
		// string "t"
		o = append(o, 0x82, 0xa1, 0x74)
		o = msgp.AppendInt16(o, int16(z.Conserz[zxhx].T))
		// string "r"
		o = append(o, 0xa1, 0x72)
		o, err = z.Conserz[zxhx].R.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *roaringArray) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Keys":
			var zxpk uint32
			zxpk, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Keys) >= int(zxpk) {
				z.Keys = (z.Keys)[:zxpk]
			} else {
				z.Keys = make([]uint16, zxpk)
			}
			for zhct := range z.Keys {
				z.Keys[zhct], bts, err = msgp.ReadUint16Bytes(bts)
				if err != nil {
					return
				}
			}
		case "NeedCopyOnWrite":
			var zdnj uint32
			zdnj, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.NeedCopyOnWrite) >= int(zdnj) {
				z.NeedCopyOnWrite = (z.NeedCopyOnWrite)[:zdnj]
			} else {
				z.NeedCopyOnWrite = make([]bool, zdnj)
			}
			for zcua := range z.NeedCopyOnWrite {
				z.NeedCopyOnWrite[zcua], bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					return
				}
			}
		case "CopyOnWrite":
			z.CopyOnWrite, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "Conserz":
			var zobc uint32
			zobc, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Conserz) >= int(zobc) {
				z.Conserz = (z.Conserz)[:zobc]
			} else {
				z.Conserz = make([]containerSerz, zobc)
			}
			for zxhx := range z.Conserz {
				var zsnv uint32
				zsnv, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zsnv > 0 {
					zsnv--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "t":
						{
							var zkgt int16
							zkgt, bts, err = msgp.ReadInt16Bytes(bts)
							z.Conserz[zxhx].T = contype(zkgt)
						}
						if err != nil {
							return
						}
					case "r":
						bts, err = z.Conserz[zxhx].R.UnmarshalMsg(bts)
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
func (z *roaringArray) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize + (len(z.Keys) * (msgp.Uint16Size)) + 16 + msgp.ArrayHeaderSize + (len(z.NeedCopyOnWrite) * (msgp.BoolSize)) + 12 + msgp.BoolSize + 8 + msgp.ArrayHeaderSize
	for zxhx := range z.Conserz {
		s += 1 + 2 + msgp.Int16Size + 2 + z.Conserz[zxhx].R.Msgsize()
	}
	return
}
