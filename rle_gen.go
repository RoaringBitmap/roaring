package roaring

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *RunIterator32) DecodeMsg(dc *msgp.Reader) (err error) {
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
					z.rc = new(runContainer32)
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
			z.curPosInIndex, err = dc.ReadUint32()
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
func (z *RunIterator32) EncodeMsg(en *msgp.Writer) (err error) {
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
	err = en.WriteUint32(z.curPosInIndex)
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
func (z *RunIterator32) MarshalMsg(b []byte) (o []byte, err error) {
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
	o = msgp.AppendUint32(o, z.curPosInIndex)
	// string "curSeq"
	o = append(o, 0xa6, 0x63, 0x75, 0x72, 0x53, 0x65, 0x71)
	o = msgp.AppendInt64(o, z.curSeq)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RunIterator32) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
					z.rc = new(runContainer32)
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
			z.curPosInIndex, bts, err = msgp.ReadUint32Bytes(bts)
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
func (z *RunIterator32) Msgsize() (s int) {
	s = 1 + 3
	if z.rc == nil {
		s += msgp.NilSize
	} else {
		s += z.rc.Msgsize()
	}
	s += 9 + msgp.Int64Size + 14 + msgp.Uint32Size + 7 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *addHelper32) DecodeMsg(dc *msgp.Reader) (err error) {
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
			z.runstart, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "runlen":
			z.runlen, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "actuallyAdded":
			z.actuallyAdded, err = dc.ReadUint32()
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
				z.m = make([]interval32, zwht)
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
					case "start":
						z.m[zbai].start, err = dc.ReadUint32()
						if err != nil {
							return
						}
					case "last":
						z.m[zbai].last, err = dc.ReadUint32()
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
					z.rc = new(runContainer32)
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
					case "iv":
						var zxhx uint32
						zxhx, err = dc.ReadArrayHeader()
						if err != nil {
							return
						}
						if cap(z.rc.iv) >= int(zxhx) {
							z.rc.iv = (z.rc.iv)[:zxhx]
						} else {
							z.rc.iv = make([]interval32, zxhx)
						}
						for zcmr := range z.rc.iv {
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
									z.rc.iv[zcmr].start, err = dc.ReadUint32()
									if err != nil {
										return
									}
								case "last":
									z.rc.iv[zcmr].last, err = dc.ReadUint32()
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
						z.rc.card, err = dc.ReadInt64()
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
func (z *addHelper32) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "runstart"
	err = en.Append(0x85, 0xa8, 0x72, 0x75, 0x6e, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.runstart)
	if err != nil {
		return
	}
	// write "runlen"
	err = en.Append(0xa6, 0x72, 0x75, 0x6e, 0x6c, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.runlen)
	if err != nil {
		return
	}
	// write "actuallyAdded"
	err = en.Append(0xad, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x6c, 0x79, 0x41, 0x64, 0x64, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.actuallyAdded)
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
		err = en.WriteUint32(z.m[zbai].start)
		if err != nil {
			return
		}
		// write "last"
		err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint32(z.m[zbai].last)
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
		// write "iv"
		err = en.Append(0x82, 0xa2, 0x69, 0x76)
		if err != nil {
			return err
		}
		err = en.WriteArrayHeader(uint32(len(z.rc.iv)))
		if err != nil {
			return
		}
		for zcmr := range z.rc.iv {
			// map header, size 2
			// write "start"
			err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
			if err != nil {
				return err
			}
			err = en.WriteUint32(z.rc.iv[zcmr].start)
			if err != nil {
				return
			}
			// write "last"
			err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
			if err != nil {
				return err
			}
			err = en.WriteUint32(z.rc.iv[zcmr].last)
			if err != nil {
				return
			}
		}
		// write "card"
		err = en.Append(0xa4, 0x63, 0x61, 0x72, 0x64)
		if err != nil {
			return err
		}
		err = en.WriteInt64(z.rc.card)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *addHelper32) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "runstart"
	o = append(o, 0x85, 0xa8, 0x72, 0x75, 0x6e, 0x73, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint32(o, z.runstart)
	// string "runlen"
	o = append(o, 0xa6, 0x72, 0x75, 0x6e, 0x6c, 0x65, 0x6e)
	o = msgp.AppendUint32(o, z.runlen)
	// string "actuallyAdded"
	o = append(o, 0xad, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x6c, 0x79, 0x41, 0x64, 0x64, 0x65, 0x64)
	o = msgp.AppendUint32(o, z.actuallyAdded)
	// string "m"
	o = append(o, 0xa1, 0x6d)
	o = msgp.AppendArrayHeader(o, uint32(len(z.m)))
	for zbai := range z.m {
		// map header, size 2
		// string "start"
		o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint32(o, z.m[zbai].start)
		// string "last"
		o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint32(o, z.m[zbai].last)
	}
	// string "rc"
	o = append(o, 0xa2, 0x72, 0x63)
	if z.rc == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 2
		// string "iv"
		o = append(o, 0x82, 0xa2, 0x69, 0x76)
		o = msgp.AppendArrayHeader(o, uint32(len(z.rc.iv)))
		for zcmr := range z.rc.iv {
			// map header, size 2
			// string "start"
			o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
			o = msgp.AppendUint32(o, z.rc.iv[zcmr].start)
			// string "last"
			o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
			o = msgp.AppendUint32(o, z.rc.iv[zcmr].last)
		}
		// string "card"
		o = append(o, 0xa4, 0x63, 0x61, 0x72, 0x64)
		o = msgp.AppendInt64(o, z.rc.card)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *addHelper32) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
			z.runstart, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "runlen":
			z.runlen, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "actuallyAdded":
			z.actuallyAdded, bts, err = msgp.ReadUint32Bytes(bts)
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
				z.m = make([]interval32, zpks)
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
					case "start":
						z.m[zbai].start, bts, err = msgp.ReadUint32Bytes(bts)
						if err != nil {
							return
						}
					case "last":
						z.m[zbai].last, bts, err = msgp.ReadUint32Bytes(bts)
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
					z.rc = new(runContainer32)
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
					case "iv":
						var zeff uint32
						zeff, bts, err = msgp.ReadArrayHeaderBytes(bts)
						if err != nil {
							return
						}
						if cap(z.rc.iv) >= int(zeff) {
							z.rc.iv = (z.rc.iv)[:zeff]
						} else {
							z.rc.iv = make([]interval32, zeff)
						}
						for zcmr := range z.rc.iv {
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
								case "start":
									z.rc.iv[zcmr].start, bts, err = msgp.ReadUint32Bytes(bts)
									if err != nil {
										return
									}
								case "last":
									z.rc.iv[zcmr].last, bts, err = msgp.ReadUint32Bytes(bts)
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
						z.rc.card, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z *addHelper32) Msgsize() (s int) {
	s = 1 + 9 + msgp.Uint32Size + 7 + msgp.Uint32Size + 14 + msgp.Uint32Size + 2 + msgp.ArrayHeaderSize + (len(z.m) * (12 + msgp.Uint32Size + msgp.Uint32Size)) + 3
	if z.rc == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 3 + msgp.ArrayHeaderSize + (len(z.rc.iv) * (12 + msgp.Uint32Size + msgp.Uint32Size)) + 5 + msgp.Int64Size
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *interval32) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "start":
			z.start, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "last":
			z.last, err = dc.ReadUint32()
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
func (z interval32) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "start"
	err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.start)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z interval32) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "start"
	o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	o = msgp.AppendUint32(o, z.start)
	// string "last"
	o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
	o = msgp.AppendUint32(o, z.last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *interval32) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "start":
			z.start, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "last":
			z.last, bts, err = msgp.ReadUint32Bytes(bts)
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
func (z interval32) Msgsize() (s int) {
	s = 1 + 6 + msgp.Uint32Size + 5 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *runContainer32) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "iv":
			var zkgt uint32
			zkgt, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.iv) >= int(zkgt) {
				z.iv = (z.iv)[:zkgt]
			} else {
				z.iv = make([]interval32, zkgt)
			}
			for zobc := range z.iv {
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
					case "start":
						z.iv[zobc].start, err = dc.ReadUint32()
						if err != nil {
							return
						}
					case "last":
						z.iv[zobc].last, err = dc.ReadUint32()
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
func (z *runContainer32) EncodeMsg(en *msgp.Writer) (err error) {
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
	for zobc := range z.iv {
		// map header, size 2
		// write "start"
		err = en.Append(0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint32(z.iv[zobc].start)
		if err != nil {
			return
		}
		// write "last"
		err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteUint32(z.iv[zobc].last)
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
func (z *runContainer32) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "iv"
	o = append(o, 0x82, 0xa2, 0x69, 0x76)
	o = msgp.AppendArrayHeader(o, uint32(len(z.iv)))
	for zobc := range z.iv {
		// map header, size 2
		// string "start"
		o = append(o, 0x82, 0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
		o = msgp.AppendUint32(o, z.iv[zobc].start)
		// string "last"
		o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
		o = msgp.AppendUint32(o, z.iv[zobc].last)
	}
	// string "card"
	o = append(o, 0xa4, 0x63, 0x61, 0x72, 0x64)
	o = msgp.AppendInt64(o, z.card)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *runContainer32) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "iv":
			var zqke uint32
			zqke, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.iv) >= int(zqke) {
				z.iv = (z.iv)[:zqke]
			} else {
				z.iv = make([]interval32, zqke)
			}
			for zobc := range z.iv {
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
					case "start":
						z.iv[zobc].start, bts, err = msgp.ReadUint32Bytes(bts)
						if err != nil {
							return
						}
					case "last":
						z.iv[zobc].last, bts, err = msgp.ReadUint32Bytes(bts)
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
func (z *runContainer32) Msgsize() (s int) {
	s = 1 + 3 + msgp.ArrayHeaderSize + (len(z.iv) * (12 + msgp.Uint32Size + msgp.Uint32Size)) + 5 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *uint32Slice) DecodeMsg(dc *msgp.Reader) (err error) {
	var zjpj uint32
	zjpj, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zjpj) {
		(*z) = (*z)[:zjpj]
	} else {
		(*z) = make(uint32Slice, zjpj)
	}
	for zywj := range *z {
		(*z)[zywj], err = dc.ReadUint32()
		if err != nil {
			return
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z uint32Slice) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zzpf := range z {
		err = en.WriteUint32(z[zzpf])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z uint32Slice) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zzpf := range z {
		o = msgp.AppendUint32(o, z[zzpf])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *uint32Slice) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zgmo uint32
	zgmo, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zgmo) {
		(*z) = (*z)[:zgmo]
	} else {
		(*z) = make(uint32Slice, zgmo)
	}
	for zrfe := range *z {
		(*z)[zrfe], bts, err = msgp.ReadUint32Bytes(bts)
		if err != nil {
			return
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z uint32Slice) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (len(z) * (msgp.Uint32Size))
	return
}
