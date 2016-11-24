package roaring

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *bitmapContainer) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Cardinality":
			z.Cardinality, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Bitmap":
			var zbai uint32
			zbai, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Bitmap) >= int(zbai) {
				z.Bitmap = (z.Bitmap)[:zbai]
			} else {
				z.Bitmap = make([]uint64, zbai)
			}
			for zxvk := range z.Bitmap {
				z.Bitmap[zxvk], err = dc.ReadUint64()
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
func (z *bitmapContainer) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Cardinality"
	err = en.Append(0x82, 0xab, 0x43, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Cardinality)
	if err != nil {
		return
	}
	// write "Bitmap"
	err = en.Append(0xa6, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Bitmap)))
	if err != nil {
		return
	}
	for zxvk := range z.Bitmap {
		err = en.WriteUint64(z.Bitmap[zxvk])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *bitmapContainer) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Cardinality"
	o = append(o, 0x82, 0xab, 0x43, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79)
	o = msgp.AppendInt(o, z.Cardinality)
	// string "Bitmap"
	o = append(o, 0xa6, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Bitmap)))
	for zxvk := range z.Bitmap {
		o = msgp.AppendUint64(o, z.Bitmap[zxvk])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *bitmapContainer) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Cardinality":
			z.Cardinality, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Bitmap":
			var zajw uint32
			zajw, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Bitmap) >= int(zajw) {
				z.Bitmap = (z.Bitmap)[:zajw]
			} else {
				z.Bitmap = make([]uint64, zajw)
			}
			for zxvk := range z.Bitmap {
				z.Bitmap[zxvk], bts, err = msgp.ReadUint64Bytes(bts)
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
func (z *bitmapContainer) Msgsize() (s int) {
	s = 1 + 12 + msgp.IntSize + 7 + msgp.ArrayHeaderSize + (len(z.Bitmap) * (msgp.Uint64Size))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *bitmapContainerShortIterator) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zhct uint32
	zhct, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ptr":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.ptr = nil
			} else {
				if z.ptr == nil {
					z.ptr = new(bitmapContainer)
				}
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
					case "Cardinality":
						z.ptr.Cardinality, err = dc.ReadInt()
						if err != nil {
							return
						}
					case "Bitmap":
						var zxhx uint32
						zxhx, err = dc.ReadArrayHeader()
						if err != nil {
							return
						}
						if cap(z.ptr.Bitmap) >= int(zxhx) {
							z.ptr.Bitmap = (z.ptr.Bitmap)[:zxhx]
						} else {
							z.ptr.Bitmap = make([]uint64, zxhx)
						}
						for zwht := range z.ptr.Bitmap {
							z.ptr.Bitmap[zwht], err = dc.ReadUint64()
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
			}
		case "i":
			z.i, err = dc.ReadInt()
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
func (z *bitmapContainerShortIterator) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "ptr"
	err = en.Append(0x82, 0xa3, 0x70, 0x74, 0x72)
	if err != nil {
		return err
	}
	if z.ptr == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		// map header, size 2
		// write "Cardinality"
		err = en.Append(0x82, 0xab, 0x43, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79)
		if err != nil {
			return err
		}
		err = en.WriteInt(z.ptr.Cardinality)
		if err != nil {
			return
		}
		// write "Bitmap"
		err = en.Append(0xa6, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70)
		if err != nil {
			return err
		}
		err = en.WriteArrayHeader(uint32(len(z.ptr.Bitmap)))
		if err != nil {
			return
		}
		for zwht := range z.ptr.Bitmap {
			err = en.WriteUint64(z.ptr.Bitmap[zwht])
			if err != nil {
				return
			}
		}
	}
	// write "i"
	err = en.Append(0xa1, 0x69)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.i)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *bitmapContainerShortIterator) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "ptr"
	o = append(o, 0x82, 0xa3, 0x70, 0x74, 0x72)
	if z.ptr == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 2
		// string "Cardinality"
		o = append(o, 0x82, 0xab, 0x43, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79)
		o = msgp.AppendInt(o, z.ptr.Cardinality)
		// string "Bitmap"
		o = append(o, 0xa6, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70)
		o = msgp.AppendArrayHeader(o, uint32(len(z.ptr.Bitmap)))
		for zwht := range z.ptr.Bitmap {
			o = msgp.AppendUint64(o, z.ptr.Bitmap[zwht])
		}
	}
	// string "i"
	o = append(o, 0xa1, 0x69)
	o = msgp.AppendInt(o, z.i)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *bitmapContainerShortIterator) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zlqf uint32
	zlqf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zlqf > 0 {
		zlqf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ptr":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.ptr = nil
			} else {
				if z.ptr == nil {
					z.ptr = new(bitmapContainer)
				}
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
					case "Cardinality":
						z.ptr.Cardinality, bts, err = msgp.ReadIntBytes(bts)
						if err != nil {
							return
						}
					case "Bitmap":
						var zpks uint32
						zpks, bts, err = msgp.ReadArrayHeaderBytes(bts)
						if err != nil {
							return
						}
						if cap(z.ptr.Bitmap) >= int(zpks) {
							z.ptr.Bitmap = (z.ptr.Bitmap)[:zpks]
						} else {
							z.ptr.Bitmap = make([]uint64, zpks)
						}
						for zwht := range z.ptr.Bitmap {
							z.ptr.Bitmap[zwht], bts, err = msgp.ReadUint64Bytes(bts)
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
			}
		case "i":
			z.i, bts, err = msgp.ReadIntBytes(bts)
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
func (z *bitmapContainerShortIterator) Msgsize() (s int) {
	s = 1 + 4
	if z.ptr == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 12 + msgp.IntSize + 7 + msgp.ArrayHeaderSize + (len(z.ptr.Bitmap) * (msgp.Uint64Size))
	}
	s += 2 + msgp.IntSize
	return
}
