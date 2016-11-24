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
	var zajw uint32
	zajw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
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
			var zwht uint32
			zwht, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.m) >= int(zwht) {
				z.m = (z.m)[:zwht]
			} else {
				z.m = make([]interval16, zwht)
			}
			for zbai := range z.m {
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
					case "Start":
						z.m[zbai].Start, err = dc.ReadUint16()
						if err != nil {
							return
						}
					case "Last":
						z.m[zbai].Last, err = dc.ReadUint16()
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
					case "Iv":
						var zxhx uint32
						zxhx, err = dc.ReadArrayHeader()
						if err != nil {
							return
						}
						if cap(z.rc.Iv) >= int(zxhx) {
							z.rc.Iv = (z.rc.Iv)[:zxhx]
						} else {
							z.rc.Iv = make([]interval16, zxhx)
						}
						for zcmr := range z.rc.Iv {
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
								case "Start":
									z.rc.Iv[zcmr].Start, err = dc.ReadUint16()
									if err != nil {
										return
									}
								case "Last":
									z.rc.Iv[zcmr].Last, err = dc.ReadUint16()
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
						z.rc.Card, err = dc.ReadInt64()
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
		// write "Start"
		err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.m[zbai].Start)
		if err != nil {
			return
		}
		// write "Last"
		err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.m[zbai].Last)
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
		// map header, size 2
		// write "Iv"
		err = en.Append(0x82, 0xa2, 0x49, 0x76)
		if err != nil {
			return err
		}
		err = en.WriteArrayHeader(uint32(len(z.rc.Iv)))
		if err != nil {
			return
		}
		for zcmr := range z.rc.Iv {
			// map header, size 2
			// write "Start"
			err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
			if err != nil {
				return err
			}
			err = en.WriteUint16(z.rc.Iv[zcmr].Start)
			if err != nil {
				return
			}
			// write "Last"
			err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
			if err != nil {
				return err
			}
			err = en.WriteUint16(z.rc.Iv[zcmr].Last)
			if err != nil {
				return
			}
		}
		// write "Card"
		err = en.Append(0xa4, 0x43, 0x61, 0x72, 0x64)
		if err != nil {
			return err
		}
		err = en.WriteInt64(z.rc.Card)
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
		// string "Start"
		o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint16(o, z.m[zbai].Start)
		// string "Last"
		o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint16(o, z.m[zbai].Last)
	}
	// string "rc"
	o = append(o, 0xa2, 0x72, 0x63)
	if z.rc == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 2
		// string "Iv"
		o = append(o, 0x82, 0xa2, 0x49, 0x76)
		o = msgp.AppendArrayHeader(o, uint32(len(z.rc.Iv)))
		for zcmr := range z.rc.Iv {
			// map header, size 2
			// string "Start"
			o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
			o = msgp.AppendUint16(o, z.rc.Iv[zcmr].Start)
			// string "Last"
			o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
			o = msgp.AppendUint16(o, z.rc.Iv[zcmr].Last)
		}
		// string "Card"
		o = append(o, 0xa4, 0x43, 0x61, 0x72, 0x64)
		o = msgp.AppendInt64(o, z.rc.Card)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *addHelper16) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
			var zpks uint32
			zpks, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.m) >= int(zpks) {
				z.m = (z.m)[:zpks]
			} else {
				z.m = make([]interval16, zpks)
			}
			for zbai := range z.m {
				var zjfb uint32
				zjfb, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zjfb > 0 {
					zjfb--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.m[zbai].Start, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					case "Last":
						z.m[zbai].Last, bts, err = msgp.ReadUint16Bytes(bts)
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
				var zcxo uint32
				zcxo, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zcxo > 0 {
					zcxo--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Iv":
						var zeff uint32
						zeff, bts, err = msgp.ReadArrayHeaderBytes(bts)
						if err != nil {
							return
						}
						if cap(z.rc.Iv) >= int(zeff) {
							z.rc.Iv = (z.rc.Iv)[:zeff]
						} else {
							z.rc.Iv = make([]interval16, zeff)
						}
						for zcmr := range z.rc.Iv {
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
								case "Start":
									z.rc.Iv[zcmr].Start, bts, err = msgp.ReadUint16Bytes(bts)
									if err != nil {
										return
									}
								case "Last":
									z.rc.Iv[zcmr].Last, bts, err = msgp.ReadUint16Bytes(bts)
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
						z.rc.Card, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z *addHelper16) Msgsize() (s int) {
	s = 1 + 9 + msgp.Uint16Size + 7 + msgp.Uint16Size + 14 + msgp.Uint16Size + 2 + msgp.ArrayHeaderSize + (len(z.m) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 3
	if z.rc == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 3 + msgp.ArrayHeaderSize + (len(z.rc.Iv) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 5 + msgp.Int64Size
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *interval16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zxpk uint32
	zxpk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zxpk > 0 {
		zxpk--
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
func (z interval16) EncodeMsg(en *msgp.Writer) (err error) {
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
func (z interval16) MarshalMsg(b []byte) (o []byte, err error) {
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
func (z *interval16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
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
func (z interval16) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint16Size + 5 + msgp.Uint16Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *runContainer16) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zsnv uint32
	zsnv, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zsnv > 0 {
		zsnv--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zkgt uint32
			zkgt, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zkgt) {
				z.Iv = (z.Iv)[:zkgt]
			} else {
				z.Iv = make([]interval16, zkgt)
			}
			for zobc := range z.Iv {
				var zema uint32
				zema, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zema > 0 {
					zema--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zobc].Start, err = dc.ReadUint16()
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zobc].Last, err = dc.ReadUint16()
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
func (z *runContainer16) EncodeMsg(en *msgp.Writer) (err error) {
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
	for zobc := range z.Iv {
		// map header, size 2
		// write "Start"
		err = en.Append(0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.Iv[zobc].Start)
		if err != nil {
			return
		}
		// write "Last"
		err = en.Append(0xa4, 0x4c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint16(z.Iv[zobc].Last)
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
func (z *runContainer16) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Iv"
	o = append(o, 0x82, 0xa2, 0x49, 0x76)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Iv)))
	for zobc := range z.Iv {
		// map header, size 2
		// string "Start"
		o = append(o, 0x82, 0xa5, 0x53, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint16(o, z.Iv[zobc].Start)
		// string "Last"
		o = append(o, 0xa4, 0x4c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint16(o, z.Iv[zobc].Last)
	}
	// string "Card"
	o = append(o, 0xa4, 0x43, 0x61, 0x72, 0x64)
	o = msgp.AppendInt64(o, z.Card)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *runContainer16) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zpez uint32
	zpez, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zpez > 0 {
		zpez--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Iv":
			var zqke uint32
			zqke, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Iv) >= int(zqke) {
				z.Iv = (z.Iv)[:zqke]
			} else {
				z.Iv = make([]interval16, zqke)
			}
			for zobc := range z.Iv {
				var zqyh uint32
				zqyh, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zqyh > 0 {
					zqyh--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Start":
						z.Iv[zobc].Start, bts, err = msgp.ReadUint16Bytes(bts)
						if err != nil {
							return
						}
					case "Last":
						z.Iv[zobc].Last, bts, err = msgp.ReadUint16Bytes(bts)
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
func (z *runContainer16) Msgsize() (s int) {
	s = 1 + 3 + msgp.ArrayHeaderSize + (len(z.Iv) * (12 + msgp.Uint16Size + msgp.Uint16Size)) + 5 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *uint16Slice) DecodeMsg(dc *msgp.Reader) (err error) {
	var zjpj uint32
	zjpj, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zjpj) {
		(*z) = (*z)[:zjpj]
	} else {
		(*z) = make(uint16Slice, zjpj)
	}
	for zywj := range *z {
		(*z)[zywj], err = dc.ReadUint16()
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
	for zzpf := range z {
		err = en.WriteUint16(z[zzpf])
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
	for zzpf := range z {
		o = msgp.AppendUint16(o, z[zzpf])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *uint16Slice) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zgmo uint32
	zgmo, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zgmo) {
		(*z) = (*z)[:zgmo]
	} else {
		(*z) = make(uint16Slice, zgmo)
	}
	for zrfe := range *z {
		(*z)[zrfe], bts, err = msgp.ReadUint16Bytes(bts)
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
