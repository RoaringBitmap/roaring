package roaring

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	randv2 "math/rand/v2"

	"strconv"
	"testing"

	"github.com/bits-and-blooms/bitset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuzzerRepro_1761183632825411000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAJAAAADwAAABgAAAAaAAAAVeD8/w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	bm.Describe()
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAEAAAAJAAAAEAAAAFXg")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm2.Describe()
	bm.AndNot(bm2)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

func TestFuzzerRepro_1761181918459062000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAJAAAADwAAABgAAAAaAAAAVeD8/w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAEAAAAJAAAAEAAAAFXg")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm.AndNot(bm2)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

func TestFuzzerRepro_1761177588768443000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEAAAMAAQADAAMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	bm.Describe()
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAEAAAAAAAAAEAAAAAEA")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm2.Describe()
	bm.Or(bm2)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

func TestFuzzerPanicRepro_1761177447422379000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAJAAAADwAAABgAAAAaAAAAVeD8/w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAEAAAAJAAAAEAAAAFXg")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm.AndNot(bm2)
}

func TestFuzzerPanicRepro_1761174531957958000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm.Xor(bm2)
}

func TestFuzzerPanicRepro_1761174003952142000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm.Xor(bm2)
}

func TestFuzzerPanicRepro_1761173763060614000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	bm.Describe()
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAIAAAAAAAAADwAAABgAAAAaAAAAAAD//w==")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm2.Describe()
	bm.Or(bm2)
	bm.Describe()
}

func TestFuzzerPanicRepro_1761171725501558000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEPAAMAAQD1/wMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	bm.Describe()
	b2, _ := base64.StdEncoding.DecodeString("OjAAAAEAAAAPAAEAEAAAAPb/9/8=")
	bm2 := NewBitmap()
	bm2.UnmarshalBinary(b2)
	bm2.Describe()
	bm.AndNot(bm2)
	bm.Describe()
}

func TestFuzzerRepro_1761171217612329001(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEPAAMAAQDo/wMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	bm.Flip(uint64(1048555), uint64(1048555)+1)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

func TestFuzzerRepro_1761171217612329000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEPAAMAAQDo/wMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	bm.Flip(uint64(1048555), uint64(1048555)+1)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

func TestFuzzerRepro_1761171056770369000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEAAAMAAQACAAMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	bm.Remove(3)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}
func TestFuzzerRepro_1761170740641699000(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("OzAAAAEAAAMAAQA2AAMA")
	bm := NewBitmap()
	bm.UnmarshalBinary(b)
	if err := bm.Validate(); err != nil {
		t.Errorf("Initial Validate failed: %v", err)
	}
	bm.AddInt(112)
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	} else {
		t.Logf("Validate succeeded")
	}
}

var bm1Arr = []uint32{279981785, 279982923, 279988809, 279995913}

const bm2Bytes = "OzAAAAGwEO2EbwQBdgoADXZFAFR2BABadgYAYnaiAAZ3GwAjdx0AQndaAJ533AB8eA4AjHgUAKJ4IQDFeA8B1nkLAON5FgD7eRQAEXoDABZ6AAAYeggAInpRAHV6BwB+egMAg3oFAIp6NgDCegYAynoVAOF6AQDkehYA/HpCAEF7ZACnewMArXsBALB7CQC7ewUAw3sJAM97DwDgewIA5HsQAPZ7BQD9exEAEHwDABV8CgAhfBMANnwFAD18EwBTfA0AY3w4AJ18HwC+fHwAPH0QAE99YgCzfQMAuH1DAP19sACvfhkAyn4dAOl+SQA0f3sAsX+lAFiAhgDggB0A/4AIAAmBBQAQgRQAJoEcAESBpgDsgQ8A/YGUAJOCNgDLgjUAAoNKAE6DJwB3g6gAIYQBACSEIQBHhAsAVIQSAGiEIQCLhNMAYIVVALeFBAC9hSQA44V8AGGGDwByhhAAhIYZAJ+GRgDnhhsABIdJAE+HGQBqhxgAhIcJAI+HOQDKhxYA4oc1ABmIDQAoiAQALogbAEuIcAC9iA0AzIgYAOaIEgD6iCoAJokaAEKJawKvizkA6ouLAHeMJQCejFIA8owGAPqMFAAQjRIAJI0MADKNFgBKjQEATY0IAFiNFwBxjQoAfY0EAIONYQDmjQAA6I06ACSOKgFQj8YAGJAzAE2QFQBkkHQA2pA7ABeRPwBYkQUAX5FdAL6RQwADkhYAG5IMACmSfgCpkhQAv5JlACaTEAA4k5wA1pMzAAuUBgATlEoAX5ROAa+VTQD+lRMAE5YzAEiWRACOlmwA/JYOAAyXAAAOlw8AH5cHACiXAAArlwkANpcBADmXBQBAlxYAWJcxAIuXEwCglxUAt5ccANWXAADXlxoA85cIAP2XMQAwmIAAsphyACaZdACdmU4A7ZkXAAaaVABcmhcAdZocAJOaFgCrmhYAw5osAPGaGgANmwEAEJsRACObBwAsmxwASpsDAE+bBQBWmw0AZZtmAM2bOQAInFIAXJwCAGCcOgCcnMkAZ50FAG6dGgCKnRQAoJ08AN6dAADgnQ0A750hABKeUABknhIAeJ42ALCebAAenzEAUZ8AAFOfCABdn4UA5J8AAOafQwAroCMAUKAEAFagQACYoBUAsKAxAOOgIgAHoQQADaEkADOhCQA+oQQARKEJAE+hPQCOoZ4ALqKMALyiKAHmo7oAoqQgAMSkUQAXpUEAWqW/ABumNgBTphkAbqZXAMemMwH8pwkAB6gMABWoTQBkqCgAjqg6AMqoHwDrqGwAWalEAJ+pEgCzqRIAx6kmAO+pPwAwqioAXKoSAHCqLQCfqkYA56pBACqrEgA+qyYAZqsrAJOrKAC9qwAAv6sdAN6rBwDnqxIA+6sCAP+rBQAGrBMAG6wRAC6sAAAwrA8AQawSAFWsEwBqrAQAcKwTAIWsJwCurBAAwKwMAM6sAADQrAMA1awLAOKsEwD3rAgAAa0oACutCwA5rQgAQ60qAG+tBAB1rQIAea0JAIStJQCrrQgAta0hANitCgDlrQsA8q0fABOuJgA8rhgAV64NAGauHQCFrioAsa4UAMeuGQDirhoA/q4PAA+vBQAWryYAP68TAFSvAwBZrwQAX68AAGGvGgB9rxAAj68DAJSvMADGrxQA3K8uAA2wBwAWsAsAI7AAACWwKABPsBUAZrABAGmwDQB4sCAAm7AIAKWwBwCusAEAsbATAMewIADpsBMA/rADAAOxGQAesQAAILEUADaxAAA4sRUAT7EfAHCxBgB4sRUAj7EXAKixCgC0sQEAt7EqAOOxGwAAsiEAI7ITADiyFgBQsgEAU7IOAGOyEwB4sgQAfrIAAIGyGgCdsgMAorIMALCyAwC1sgEAuLITAM2yFwDmshQA/LIEAAKzAQAFsxQAG7MNACqzCAA0swgAPrMAAEGzAABDswAARbMBAEizGQBjswYAa7MEAHGzCwB+swEAgbMEAIezEwCcsw0Aq7MHALSzCAC+swMAw7MDAMizAwDNswYA1bMUAOuzCQD2swEA+bMAAPuzDAAJtAcAErQAABS0AQAYtAsAJbQEACy0DQA7tBAATbQZAGi0IACKtAQAkLQCAJS0AQCXtAQAnbQAAJ+0BQCmtBMAu7QOAMu0GgDntAkA8rQCAPa0JQAdtQAAH7UHACi1FgBAtQoATbUKAFm1AABbtRgAdbUTAIq1GwCntQMArLUCALC1OQDrtXAAXbYAAF+2KACJtjUAwLYgAOK2OgAetxAAMLcWAEi3AABKt0cAk7cVAKq3KQDVtxIA6bcSAP23AAD/twcACLg0AD64RgCGuDEAubgkAN+4CQDquBQAALkKAAy5DwAduQAAH7kIACu5CAA3uQgAQbkCAEW5AABIuQkAVbkKAGG5AQBkuQcAbrkBAHO5FACJuQIAjbkAAJC5AwCVuQMAmrkAAJy5AgChuQMAqLkHALG5AAC0uQIAuLkHAMO5JQDquQIA8LkXAAm6AAALugkAFroDABu6BQAiugIAKLoIADS6AQA3ugUAProKAEq6AQBNugAAULoJAFy6AQBfugUAZroKAHK6AgB6ugAAfLoUAJK6AgCYugcAoboAAKS6FAC6ugEAvboAAL+6BQDGugEAyboBAMy6BwDVug8A5roAAOi6CADyug0AAbsBAAS7DgAUuxcALbsAADC7BAA2uwIAO7sCAEC7GwBduykAiLsqALS7gwA5vFAAi7wEAJG8JAC3vEIA+7wGAAO9EAAVvQUAHL0SADC9CgA8vTIAcL0KAHy9AQB/vRcAmL0TAK29GwDKvVoAJr4iAEq+DABYvlcCscDiAJXBeAAPwh0ALsInAFfCEABpwgQAb8IJAHrCBwCDwgUAisIRAJ3CEgCxwgQAt8IYANHCEADjwgcA7cIAAPDCKwAdwx8APsMAAEDDSACKwwMAj8MWAKfDCgCzwzMA6MP8AObEcABYxQEAW8UEAGHFDABvxQUAdsUQAIjFDQCXxQ0ApsUYAMDFAADCxSEA5cUAAOfFCgDzxQkA/sUFAAXGBQAMxgcAFcYNACTGKwBRxjIAhcYEAIvGNQDCxmwAMMfSAATIXQBjyAMAaMgVAH/IFwCYyCYAwMgGAMjICwDVyA4A5cggAAfJCAARyRcAKskSAD7JAQBByQoATckXAGbJGgCCyQAAhMkLAJLJFQCpyRMAvskEAMTJWAEeywAAIMsnAEnLAQBMyw8AXcsUAHPLLQCiyw8As8sAALXLAAC4yxoA1MsmAP3LAgABzCsAL8x8AK3MGADHzAMAzMwLANnMCwDnzBcAAM0QABLNGQAtzQoAOc0GAEHNEQBUzQkAX80TAHXNFQCMzQoAmc1JAOXNDADzzScAHM4qAEjOiwDVzgQA284NAOrOdwJj0Q0ActEBAHXRTQDE0QcAzdFQAB/SEAAx0joBbdOhABDUJwA51HMArtSSAELV6gAu1gcAN9YYAFHWIwB21gMAe9YAAH3WAwCC1hsAn9YKAKvWAQCu1gcAt9YBALrWBwDD1gEAxtYJANLWGwDv1gIA89YAAPXWFwAO1wAAENcCABTXFgAs1zoAaNdBAKvXnwBM2AEAT9gGAFfYJwCA2HkA+9idAJrZGgC22QEBudoVANDaOQAL22wAedu8ADfcLABl3CYAjdwcAKvcFgDD3EwAEd1fAHLdTQDB3R8A4t0fAAPenACh3jEA1N4SAOjehQBv3wAAcd9AALPfAQC23xsA098zAAjgEgAc4AIAIOADACXgBQAs4AAAL+ADADTgBAA64AAAPOAKAEjgCQBT4AcAXOAHAGXgHQCE4AAAhuAKAJLgFQCp4A0AuOAkAN7gAwDk4A4A9OAlABvhGQA24RcAT+ECAFPhGABt4QAAb+EbAIzhCwCZ4QUAoOEGAKjhFwDB4QAAw+EBAMbhDgDW4QsA4+EdAALiAwAH4gsAFOIDABniIgA94gEAQOIhAGPiEAB14i0ApOISALjiFADO4gwA3OITAPHiAQD04gYA/OIZABfjAQAa4wIAHuMBACHjBgAp4wQAL+MKADvjDABJ4wwAV+MOAGfjBQBu4wEAceMAAHPjBQB64woAhuMBAInjBwCT4w4Ao+MIAK3jAgCx4wcAuuMDAL/jCgDL4wAAzeMAAM/jBADW4wMA2+MEAOHjBADn4woA8+MCAPfjBQD+4wIAA+QEAArkEAAc5AAAIuQEACjkCgA05AAAOOQDAD7kDABM5AAATuQAAFDkAwBW5AMAW+QEAGHkBQBo5AoAdOQAAHjkAwB+5AwAjOQAAJDkAwCV5AQAnOQGAKTkAQCn5AQAreQDALTkCgDA5AkAy+QEANHkBQDY5AcA4eQHAOrkAgDu5AAA8OQCAPTkBgD85A4ADOUDABHlDQAg5TMAVeUdAHTlBAB65QIAfuUQAJDlFgCo5RAAuuUIAMTlCQDP5SsA/OUdABvmCQAm5hoAQuYSAFbmHAB05ikAn+YSALTmLgDl5g0A9OYxACfnAQAq5wIALuc+AG7nDgF+6BQAlOguAMToqQBv6QwAfekaAJnphwAj6rEA1uoeAPbqTwBH64YAz+sIANnrPgAZ7CYAQewnAGrsBABw7BwAjuwCAJLsCwCf7AIAo+xGAOvsNAAh7QAAI+07AGDtBQBn7S0Alu0GAJ7tLwDP7RYA5+0MAPXtJgAd7iIAQe4BAETuEQBY7g0AZ+4NAHbuLACk7jIA2O4UAO7uJgAW7xIAKu8zAF/vcQDS7yUA+e8RAA3wAgAR8BIAJfAqAFHwBABX8BAAafAfAIrwAQCN8BIAofALAK7wIwDT8A0A4vAHAOvwEgD/8AUABvETABvxCwAp8TsAZvEwAJjxHgC48QAAuvEiAN7xAwDj8Q0A8vEDAPfxBwAA8hsAHfISADHyBAA38hMATfIGAFXyEgBp8j4AqfICAK3yTgD98hsAGvNqAIbzCgCS8xMAp/MAAKrzAwCv8wkAuvMBAL7zAADA8wEAw/MBAMbzBQDN8wAA0fMAANTzAgDY8wEA2/MIAOXzAADn8wYA7/MAAPLzAwD38wMA/PMEAAT0AQAH9A0AGPQGACD0BgAo9AAALPQBAC/0AwA09AcAQPQDAEj0BgBQ9AEAU/QGAFv0CABn9AAAafQEAG/0AQB09AUAe/QEAIH0AQCF9AIAifQCAI/0AACS9AQAmPQCAJz0DgCt9AIAsfQFALj0AAC69AQAwvQEAMj0AgDM9AMA0fQCANb0AQDZ9AQA3/QBAOT0BQDr9AcA9PQHAP70AgAD9QEABvUCAAr1EQAf9Q8AMPUDADf1BAA99QUARPUGAE31AgBR9QYAWfULAGb1AQBq9RIAfvUNAI31BACV9Q8ApvUDAKz1AwCx9QYAufUGAMH1DADP9QkA2vUBAN31EADw9QMA9fUCAPn1AQD99QkACPYAAAr2BwAT9gYAG/YGACP2AAAl9gQAK/YCADD2AAAy9gcAO/YGAET2AABG9gMAS/YAAE32AwBS9gMAV/YIAGL2AABk9hEAd/YDAHz2AAB+9gMAg/YFAIr2BgCS9gEAlvYAAJj2EQCr9gEArvYHALj2AAC69gYAwvYNANH2DADg9gsA7fYBAPD2AgD09gAA9vYBAPn2AQD89goACPcDAA33CgAZ9wEAHPcBAB/3AQAi9wsAL/cCADP3AAA29wwARPcBAEf3AQBK9wwAWPcBAFv3AQBe9wkAafcAAGv3BABy9wkAffcAAH/3DACN9wEAkPcEAJb3CQCh9wAApPcDAKn3DAC39xAAyfcBAMz3EQDf9xIA9PcYAA74AAAQ+AsAHvgGACf4EQA8+AQAQvgOAFL4AABU+BYAbfgBAHD4DgCA+AsAjfgRAKD4AwCm+BEAuvgmAOL4FAD4+BQADvkJABr5CgAm+UIAavmEAPD5CwD9+TYANfoJAEH6LQBw+i0An/oLAKz6PwDt+hoACft4AIP7BgCL+xMBoPwYALr8UwAP/RIAI/0KAC/9BwA4/QkAQ/0kAGn9HgCJ/RsApv2yAFr+AQBd/gIAYf4IAGv+CgB3/gAAef5oAOP+EAD1/mIAWf+CAN3/BwDm/xkA"

func TestRoaring_OrThenAnd(t *testing.T) {
	bm1 := BitmapOf(bm1Arr...)
	require.NoError(t, bm1.Validate())

	bm2 := New()
	_, err := bm2.FromBase64(bm2Bytes)
	require.NoError(t, err)
	require.NoError(t, bm2.Validate())

	orBm := Or(bm1, bm2) // this can be FastOr, bm1.Or(bm2)
	require.NoError(t, orBm.Validate())

	mask := New()
	mask.AddRange(279_900_001, 280_000_001)
	mask.Or(orBm) // this can be roaring.And(mask, orBm)

	require.NoError(t, mask.Validate()) // <-- fails here
}

func TestRoaring_OrThenAnd_RunOptimize(t *testing.T) {
	bm1 := BitmapOf(bm1Arr...)
	require.NoError(t, bm1.Validate())

	fmt.Printf("%v\n", bm1)
	bm2 := New()
	bm2.FromBase64(bm2Bytes) // this shows that data is not corrupted
	bm2 = BitmapOf(bm2.ToArray()...)
	require.NoError(t, bm2.Validate())

	bm2.RunOptimize()
	require.NoError(t, bm2.Validate())

	orBm := Or(bm1, bm2)
	require.NoError(t, orBm.Validate())

	mask := New()
	mask.AddRange(279_900_001, 280_000_001)
	require.NoError(t, mask.Validate())

	mask.And(orBm)

	require.NoError(t, mask.Validate())
}

func TestSplitter_BrokeBm(t *testing.T) {
	bm1 := NewBitmap()
	bm2 := NewBitmap()
	bm3 := NewBitmap()

	for i := 0; i < 2000; i++ {
		bm1.Add(uint32(i))
		bm2.Add(uint32(i + 2000))
		bm3.Add(uint32(i * 4000))
	}
	res := FastOr(bm1, bm2, bm3)

	require.NoError(t, res.Validate())

	return
}

func TestRoaring_AndMask(t *testing.T) {
	r := randv2.New(randv2.NewPCG(13, 22)) // reproducible random

	rb := NewBitmap()
	for j := 0; j < 1_000_000; j++ {
		rb.Add(r.Uint32N(10_000_000) + 5_000_000)
	}

	mask := New()
	mask.AddRange(1, 10_000_001)
	mask.And(rb)

	err := mask.Validate()
	require.NoError(t, err) // this fails
}

func TestIssue440(t *testing.T) {
	a := NewBitmap()
	a.AddMany([]uint32{1, 2, 3})
	a.RunOptimize()
	b1, err := a.MarshalBinary()
	require.NoError(t, err)
	a.RunOptimize()
	b2, err := a.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, b1, b2)
}

func TestIssue440_2(t *testing.T) {
	a := NewBitmap()
	a.AddMany([]uint32{1, 2, 3, 4})
	a.RunOptimize()
	b1, err := a.MarshalBinary()
	require.NoError(t, err)
	a.RunOptimize()
	b2, err := a.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, b1, b2)
}

func TestIssue440_3(t *testing.T) {
	a := NewBitmap()
	a.AddMany([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	a.RunOptimize()
	b1, err := a.MarshalBinary()
	require.NoError(t, err)
	a.RunOptimize()
	b2, err := a.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, b1, b2)
}

func TestIssue440_4(t *testing.T) {
	a := NewBitmap()
	a.AddMany([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	a.RunOptimize()
	b1, err := a.MarshalBinary()
	require.NoError(t, err)
	a.RunOptimize()
	b2, err := a.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, b1, b2)
}

func TestIssue440_5(t *testing.T) {
	a := NewBitmap()
	a.AddMany([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})
	a.RunOptimize()
	b1, err := a.MarshalBinary()
	require.NoError(t, err)
	a.RunOptimize()
	b2, err := a.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, b1, b2)
}

func checkValidity(t *testing.T, rb *Bitmap) {
	t.Helper()

	for _, c := range rb.highlowcontainer.containers {
		switch c.(type) {
		case *arrayContainer:
			if c.getCardinality() > arrayDefaultMaxSize {
				t.Error("Array containers are limited to size ", arrayDefaultMaxSize)
			}
		case *bitmapContainer:
			if c.getCardinality() <= arrayDefaultMaxSize {
				t.Error("Bitmaps would be more concise as an array!")
			}
		case *runContainer16:
			if c.getSizeInBytes() > minOfInt(bitmapContainerSizeInBytes(), arrayContainerSizeInBytes(c.getCardinality())) {
				t.Error("Inefficient run container!")
			}
		}
	}
}

func hashTest(t *testing.T, N uint64) {
	hashes := map[uint64]struct{}{}
	count := 0

	for gap := uint64(1); gap <= 65536; gap *= 2 {
		rb1, rb2 := NewBitmap(), NewBitmap()
		for x := uint64(0); x <= N*gap; x += gap {
			rb1.AddInt(int(x))
			rb2.AddInt(int(x))
		}

		assert.EqualValues(t, rb1.Checksum(), rb2.Checksum())
		count++
		hashes[rb1.Checksum()] = struct{}{}

		rb1, rb2 = NewBitmap(), NewBitmap()
		for x := uint64(0); x <= N*gap; x += gap {
			// x+3 guarantees runs, gap/2 guarantees some variety
			if x+3+gap/2 > MaxUint32 {
				break
			}
			rb1.AddRange(uint64(x), uint64(x+3+gap/2))
			rb2.AddRange(uint64(x), uint64(x+3+gap/2))
		}

		rb1.RunOptimize()
		rb2.RunOptimize()

		assert.EqualValues(t, rb1.Checksum(), rb2.Checksum())
		count++
		hashes[rb1.Checksum()] = struct{}{}
	}

	// Make sure that at least for this reduced set we have no collisions.
	assert.Equal(t, count, len(hashes))
}

func buildRuns(includeBroken bool) *runContainer16 {
	rc := &runContainer16{}
	if includeBroken {
		for i := 0; i < 100; i++ {
			start := i * 100
			end := start + 100
			rc.iv = append(rc.iv, newInterval16Range(uint16(start), uint16(end)))
		}
	}

	for i := 0; i < 100; i++ {
		start := i*100 + i*2
		end := start + 100
		rc.iv = append(rc.iv, newInterval16Range(uint16(start), uint16(end)))
	}

	return rc
}

func TestReverseIteratorCount(t *testing.T) {
	array := []int{2, 63, 64, 65, 4095, 4096, 4097, 4159, 4160, 4161, 5000, 20000, 66666}
	for _, testSize := range array {
		b := New()
		for i := uint32(0); i < uint32(testSize); i++ {
			b.Add(i)
		}
		it := b.ReverseIterator()
		count := 0
		for it.HasNext() {
			it.Next()
			count++
		}

		assert.Equal(t, testSize, count)
	}
}

func TestRoaringIntervalCheck(t *testing.T) {
	r := BitmapOf(1, 2, 3, 1000)
	rangeb := New()
	rangeb.AddRange(10, 1000+1)

	assert.True(t, r.Intersects(rangeb))

	rangeb2 := New()
	rangeb2.AddRange(10, 1000)

	assert.False(t, r.Intersects(rangeb2))
}

func TestRoaringRangeEnd(t *testing.T) {
	r := New()
	r.Add(MaxUint32)
	assert.EqualValues(t, 1, r.GetCardinality())

	r.RemoveRange(0, MaxUint32)
	assert.EqualValues(t, 1, r.GetCardinality())

	r.RemoveRange(0, math.MaxUint64)
	assert.EqualValues(t, 0, r.GetCardinality())

	r.Add(MaxUint32)
	assert.EqualValues(t, 1, r.GetCardinality())

	r.RemoveRange(0, 0x100000001)
	assert.EqualValues(t, 0, r.GetCardinality())

	r.Add(MaxUint32)
	assert.EqualValues(t, 1, r.GetCardinality())

	r.RemoveRange(0, 0x100000000)
	assert.EqualValues(t, 0, r.GetCardinality())
}

func TestMaxPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	bm := New()
	bm.Maximum()
	t.Errorf("The code did not panic")
}

func TestMinPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	bm := New()
	bm.Minimum()
	t.Errorf("The code did not panic")
}

func TestFirstLast(t *testing.T) {
	bm := New()
	bm.AddInt(2)
	bm.AddInt(4)
	bm.AddInt(8)

	assert.EqualValues(t, 2, bm.Minimum())
	assert.EqualValues(t, 8, bm.Maximum())

	i := 1 << 5

	for ; i < (1 << 17); i++ {
		bm.AddInt(i)

		assert.EqualValues(t, 2, bm.Minimum())
		assert.EqualValues(t, i, bm.Maximum())
	}

	bm.RunOptimize()

	assert.EqualValues(t, 2, bm.Minimum())
	assert.EqualValues(t, i-1, bm.Maximum())
}

func TestRoaringBitmapBitmapOf(t *testing.T) {
	array := []uint32{5580, 33722, 44031, 57276, 83097}
	bmp := BitmapOf(array...)

	assert.EqualValues(t, len(array), bmp.GetCardinality())

	by, _ := bmp.ToBytes()

	assert.EqualValues(t, len(by), bmp.GetSerializedSizeInBytes())
}

func TestRoaringBitmapAdd(t *testing.T) {
	array := []uint32{5580, 33722, 44031, 57276, 83097}
	bmp := New()
	for _, v := range array {
		bmp.Add(v)
	}

	assert.EqualValues(t, len(array), bmp.GetCardinality())
}

func TestRoaringBitmapAddMany(t *testing.T) {
	array := []uint32{5580, 33722, 44031, 57276, 83097}
	bmp := NewBitmap()
	bmp.AddMany(array)

	assert.EqualValues(t, len(array), bmp.GetCardinality())
}

func testAddOffset(t *testing.T, arr []uint32, offset int64) {
	expected := make([]uint32, 0, len(arr))
	for _, i := range arr {
		v := int64(i) + offset
		if v >= 0 && v <= MaxUint32 {
			expected = append(expected, uint32(v))
		}
	}

	bmp := NewBitmap()
	bmp.AddMany(arr)

	cop := AddOffset64(bmp, offset)

	if !assert.EqualValues(t, len(expected), cop.GetCardinality()) {
		t.Logf("Applying offset %d", offset)
	}
	if !assert.EqualValues(t, expected, cop.ToArray()) {
		t.Logf("Applying offset %d", offset)
	}

	// Now check backing off gets us back all non-discarded numbers
	expected2 := make([]uint32, 0, len(expected))
	for _, i := range expected {
		v := int64(i) - offset
		if v >= 0 && v <= MaxUint32 {
			expected2 = append(expected2, uint32(v))
		}
	}

	cop2 := AddOffset64(cop, -offset)

	if !assert.EqualValues(t, len(expected2), cop2.GetCardinality()) {
		t.Logf("Restoring from offset %d", offset)
	}
	if !assert.EqualValues(t, expected2, cop2.ToArray()) {
		t.Logf("Restoring from offset %d", offset)
	}
}

func TestRoaringBitmapAddOffset(t *testing.T) {
	type testCase struct {
		arr    []uint32
		offset int64
	}
	cases := []testCase{
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: 25000,
		},
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: -25000,
		},
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: -83097,
		},
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: MaxUint32,
		},
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: -MaxUint32,
		},
		{
			arr:    []uint32{5580, 33722, 44031, 57276, 83097},
			offset: 0,
		},
		{
			arr:    []uint32{0},
			offset: 100,
		},
		{
			arr:    []uint32{0},
			offset: 0xffff0000,
		},
		{
			arr:    []uint32{0},
			offset: 0xffff0001,
		},
	}

	arr := []uint32{10, 0xffff, 0x010101}
	for i := uint32(100000); i < 200000; i += 4 {
		arr = append(arr, i)
	}
	arr = append(arr, 400000)
	arr = append(arr, 1400000)
	for offset := int64(3); offset < 1000000; offset *= 3 {
		c := testCase{arr, offset}
		cases = append(cases, c)
	}
	for offset := int64(1024); offset < 1000000; offset *= 2 {
		c := testCase{arr, offset}
		cases = append(cases, c)
	}

	for _, c := range cases {
		// Positive offset
		testAddOffset(t, c.arr, c.offset)
		// Negative offset
		testAddOffset(t, c.arr, -c.offset)
	}
}

func TestRoaringInPlaceAndNotBitmapContainer(t *testing.T) {
	bm := NewBitmap()
	for i := 0; i < 8192; i++ {
		bm.Add(uint32(i))
	}
	toRemove := NewBitmap()
	for i := 128; i < 8192; i++ {
		toRemove.Add(uint32(i))
	}
	bm.AndNot(toRemove)

	var b bytes.Buffer
	_, err := bm.WriteTo(&b)

	require.NoError(t, err)

	bm2 := NewBitmap()
	bm2.ReadFrom(bytes.NewBuffer(b.Bytes()))

	assert.True(t, bm2.Equals(bm))
}

// https://github.com/RoaringBitmap/roaring/issues/64
func TestFlip64(t *testing.T) {
	bm := New()
	bm.AddInt(0)
	bm.Flip(1, 2)
	i := bm.Iterator()

	assert.False(t, i.Next() != 0 || i.Next() != 1 || i.HasNext())
}

// https://github.com/RoaringBitmap/roaring/issues/64
func TestFlip64Off(t *testing.T) {
	bm := New()
	bm.AddInt(10)
	bm.Flip(11, 12)
	i := bm.Iterator()

	assert.False(t, i.Next() != 10 || i.Next() != 11 || i.HasNext())
}

func TestStringer(t *testing.T) {
	v := NewBitmap()
	for i := uint32(0); i < 10; i++ {
		v.Add(i)
	}

	assert.Equal(t, "{0,1,2,3,4,5,6,7,8,9}", v.String())

	v.RunOptimize()

	assert.Equal(t, "{0,1,2,3,4,5,6,7,8,9}", v.String())
}

func TestFastCard(t *testing.T) {
	bm := NewBitmap()
	bm.Add(1)
	bm.AddRange(21, 260000)
	bm2 := NewBitmap()
	bm2.Add(25)

	assert.EqualValues(t, 1, bm2.AndCardinality(bm))
	assert.Equal(t, bm.GetCardinality(), bm2.OrCardinality(bm))
	assert.EqualValues(t, 1, bm.AndCardinality(bm2))
	assert.Equal(t, bm.GetCardinality(), bm.OrCardinality(bm2))
	assert.EqualValues(t, 1, bm2.AndCardinality(bm))
	assert.Equal(t, bm.GetCardinality(), bm2.OrCardinality(bm))

	bm.RunOptimize()

	assert.EqualValues(t, 1, bm2.AndCardinality(bm))
	assert.Equal(t, bm.GetCardinality(), bm2.OrCardinality(bm))
	assert.EqualValues(t, 1, bm.AndCardinality(bm2))
	assert.Equal(t, bm.GetCardinality(), bm.OrCardinality(bm2))
	assert.EqualValues(t, 1, bm2.AndCardinality(bm))
	assert.Equal(t, bm.GetCardinality(), bm2.OrCardinality(bm))
}

func TestFastCardUnequalKeys(t *testing.T) {
	// These tests will excercise the interior code branches of OrCardinality

	t.Run("Merge small into large", func(t *testing.T) {
		bm := NewBitmap()
		bm.AddRange(0, 1024)
		bm2 := NewBitmap()
		start := uint64(2 << 16)
		bm2.AddRange(start, start+3)

		assert.Equal(t, uint64(1027), bm2.OrCardinality(bm))
	})
	t.Run("Merge large into small", func(t *testing.T) {
		bm := NewBitmap()
		bm.AddRange(0, 1024)
		bm2 := NewBitmap()
		start := uint64(2 << 16)
		bm2.AddRange(start, start+3)

		assert.Equal(t, uint64(1027), bm.OrCardinality(bm2))
	})

	t.Run("Merge large into small same keyrange start", func(t *testing.T) {
		bm := NewBitmap()
		start := uint64(2 << 16)
		bm.AddRange(0, 1024)
		bm.AddRange(start, start+3)

		bm2 := NewBitmap()
		bm2.AddRange(0, 512)
		bm2.AddRange(start, start+3)

		assert.Equal(t, uint64(1027), bm.OrCardinality(bm2))
	})
}

func TestIntersects1(t *testing.T) {
	bm := NewBitmap()
	bm.Add(1)
	bm.AddRange(21, 26)
	bm2 := NewBitmap()
	bm2.Add(25)

	assert.True(t, bm2.Intersects(bm))

	bm.Remove(25)
	assert.Equal(t, false, bm2.Intersects(bm))

	bm.AddRange(1, 100000)
	assert.True(t, bm2.Intersects(bm))
}

func TestIntersectsWithInterval(t *testing.T) {
	bm := NewBitmap()
	bm.AddRange(21, 26)

	// Empty interval in range
	assert.False(t, bm.IntersectsWithInterval(22, 22))
	// Empty interval out of range
	assert.False(t, bm.IntersectsWithInterval(27, 27))

	// Non-empty interval in range, fully included
	assert.True(t, bm.IntersectsWithInterval(22, 23))
	// Non-empty intervals partially overlapped
	assert.True(t, bm.IntersectsWithInterval(19, 23))
	assert.True(t, bm.IntersectsWithInterval(23, 30))
	// Non-empty interval covering the full range
	assert.True(t, bm.IntersectsWithInterval(19, 30))

	// Non-empty interval before start of bitmap
	assert.False(t, bm.IntersectsWithInterval(19, 20))
	// Non-empty interval after end of bitmap
	assert.False(t, bm.IntersectsWithInterval(28, 30))

	// Non-empty interval inside "hole" in bitmap
	bm.AddRange(30, 40)
	assert.False(t, bm.IntersectsWithInterval(28, 29))

	// Non-empty interval, non-overlapping on the open side
	assert.False(t, bm.IntersectsWithInterval(28, 30))
	// Non-empty interval, overlapping on the open side
	assert.True(t, bm.IntersectsWithInterval(28, 31))
}

func TestRangePanic(t *testing.T) {
	bm := NewBitmap()
	bm.Add(1)
	bm.AddRange(21, 26)
	bm.AddRange(9, 14)
	bm.AddRange(11, 16)
}

func TestRangeRemoval(t *testing.T) {
	bm := NewBitmap()
	bm.Add(1)
	bm.AddRange(21, 26)
	bm.AddRange(9, 14)
	bm.RemoveRange(11, 16)
	bm.RemoveRange(1, 26)
	c := bm.GetCardinality()

	assert.EqualValues(t, 0, c)

	bm.AddRange(1, 10000)
	c = bm.GetCardinality()

	assert.EqualValues(t, 10000-1, c)

	bm.RemoveRange(1, 10000)
	c = bm.GetCardinality()

	assert.EqualValues(t, 0, c)
}

func TestRangeRemovalFromContent(t *testing.T) {
	bm := NewBitmap()
	for i := 100; i < 10000; i++ {
		bm.AddInt(i * 3)
	}
	bm.AddRange(21, 26)
	bm.AddRange(9, 14)
	bm.RemoveRange(11, 16)
	bm.RemoveRange(0, 30000)
	c := bm.GetCardinality()

	assert.EqualValues(t, 0o0, c)
}

func TestFlipOnEmpty(t *testing.T) {
	t.Run("TestFlipOnEmpty in-place", func(t *testing.T) {
		bm := NewBitmap()
		bm.Flip(0, 10)
		c := bm.GetCardinality()

		assert.EqualValues(t, 10, c)
	})

	t.Run("TestFlipOnEmpty, generating new result", func(t *testing.T) {
		bm := NewBitmap()
		bm = Flip(bm, 0, 10)
		c := bm.GetCardinality()

		assert.EqualValues(t, 10, c)
	})
}

func TestBitmapRank2(t *testing.T) {
	r := NewBitmap()
	for i := uint32(1); i < 8194; i += 2 {
		r.Add(i)
	}

	rank := r.Rank(63)
	assert.EqualValues(t, 32, rank)
}

func TestBitmapRank(t *testing.T) {
	for N := uint32(1); N <= 1048576; N *= 2 {
		t.Run("rank tests"+strconv.Itoa(int(N)), func(t *testing.T) {
			for gap := uint32(1); gap <= 65536; gap *= 2 {
				rb1 := NewBitmap()
				for x := uint32(0); x <= N; x += gap {
					rb1.Add(x)
				}
				for y := uint32(0); y <= N; y++ {
					if rb1.Rank(y) != uint64((y+1+gap-1)/gap) {
						assert.Equal(t, (y+1+gap-1)/gap, rb1.Rank(y))
					}
				}
			}
		})
	}
}

func TestBitmapSelect(t *testing.T) {
	for N := uint32(1); N <= 1048576; N *= 2 {
		t.Run("rank tests"+strconv.Itoa(int(N)), func(t *testing.T) {
			for gap := uint32(1); gap <= 65536; gap *= 2 {
				rb1 := NewBitmap()
				for x := uint32(0); x <= N; x += gap {
					rb1.Add(x)
				}
				for y := uint32(0); y <= N/gap; y++ {
					expectedInt := y * gap
					i, err := rb1.Select(y)
					if err != nil {
						t.Fatal(err)
					}

					if i != expectedInt {
						assert.Equal(t, expectedInt, i)
					}
				}
			}
		})
	}
}

// some extra tests
func TestBitmapExtra(t *testing.T) {
	for N := uint32(1); N <= 65536; N *= 2 {
		t.Run("extra tests"+strconv.Itoa(int(N)), func(t *testing.T) {
			for gap := uint32(1); gap <= 65536; gap *= 2 {
				bs1 := bitset.New(0)
				rb1 := NewBitmap()
				for x := uint32(0); x <= N; x += gap {
					bs1.Set(uint(x))
					rb1.Add(x)
				}

				assert.EqualValues(t, rb1.GetCardinality(), bs1.Count())
				assert.True(t, equalsBitSet(bs1, rb1))

				for offset := uint32(1); offset <= gap; offset *= 2 {
					bs2 := bitset.New(0)
					rb2 := NewBitmap()
					for x := uint32(0); x <= N; x += gap {
						bs2.Set(uint(x + offset))
						rb2.Add(x + offset)
					}

					assert.EqualValues(t, rb2.GetCardinality(), bs2.Count())
					assert.True(t, equalsBitSet(bs2, rb2))

					clonebs1 := bs1.Clone()
					clonebs1.InPlaceIntersection(bs2)

					if !equalsBitSet(clonebs1, And(rb1, rb2)) {
						v := rb1.Clone()
						v.And(rb2)

						assert.True(t, equalsBitSet(clonebs1, v))
					}

					// testing OR
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceUnion(bs2)

					assert.True(t, equalsBitSet(clonebs1, Or(rb1, rb2)))
					// testing XOR
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceSymmetricDifference(bs2)
					assert.True(t, equalsBitSet(clonebs1, Xor(rb1, rb2)))

					// testing NOTAND
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceDifference(bs2)
					assert.True(t, equalsBitSet(clonebs1, AndNot(rb1, rb2)))
				}
			}
		})
	}
}

func FlipRange(start, end int, bs *bitset.BitSet) {
	for i := start; i < end; i++ {
		bs.Flip(uint(i))
	}
}

func TestBitmap(t *testing.T) {
	t.Run("Test Contains", func(t *testing.T) {
		rbm1 := NewBitmap()
		for k := 0; k < 1000; k++ {
			rbm1.AddInt(17 * k)
		}

		for k := 0; k < 17*1000; k++ {
			assert.Equal(t, (k/17*17 == k), rbm1.ContainsInt(k))
		}
	})

	t.Run("Test Clone", func(t *testing.T) {
		rb1 := NewBitmap()
		rb1.Add(10)

		rb2 := rb1.Clone()
		rb2.Remove(10)

		assert.True(t, rb1.Contains(10))
	})

	t.Run("Test run array not equal", func(t *testing.T) {
		rb := NewBitmap()
		rb2 := NewBitmap()
		rb.AddRange(0, 1<<16)
		for i := 0; i < 10; i++ {
			rb2.AddInt(i)
		}

		assert.EqualValues(t, 1<<16, rb.GetCardinality())
		assert.EqualValues(t, 10, rb2.GetCardinality())
		assert.False(t, rb.Equals(rb2))

		rb.RunOptimize()
		rb2.RunOptimize()

		assert.EqualValues(t, 1<<16, rb.GetCardinality())
		assert.EqualValues(t, 10, rb2.GetCardinality())
		assert.False(t, rb.Equals(rb2))
	})

	t.Run("Test ANDNOT4", func(t *testing.T) {
		rb := NewBitmap()
		rb2 := NewBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.AddInt(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.AddInt(i)
		}

		off := AndNot(rb2, rb)
		andNotresult := AndNot(rb, rb2)

		assert.True(t, rb.Equals(andNotresult))
		assert.True(t, rb2.Equals(off))

		rb2.AndNot(rb)
		assert.True(t, rb2.Equals(off))
	})

	t.Run("Test AND", func(t *testing.T) {
		rr := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
		}
		rr.Add(100000)
		rr.Add(110000)
		rr2 := NewBitmap()
		rr2.Add(13)
		rrand := And(rr, rr2)
		array := rrand.ToArray()

		assert.Equal(t, 1, len(array))
		assert.EqualValues(t, 13, array[0])

		rr.And(rr2)
		array = rr.ToArray()

		assert.Equal(t, 1, len(array))
		assert.EqualValues(t, 13, array[0])
	})

	t.Run("Test AND 2", func(t *testing.T) {
		rr := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.AddInt(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.AddInt(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.AddInt(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.AddInt(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.AddInt(k)
		}

		rr2 := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.AddInt(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.AddInt(k)
		}
		correct := And(rr, rr2)
		rr.And(rr2)

		assert.True(t, correct.Equals(rr))
	})

	t.Run("Test AND 2", func(t *testing.T) {
		rr := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
		}
		rr.AddInt(100000)
		rr.AddInt(110000)
		rr2 := NewBitmap()
		rr2.AddInt(13)

		rrand := And(rr, rr2)
		array := rrand.ToArray()

		assert.Equal(t, 1, len(array))
		assert.EqualValues(t, 13, array[0])
	})

	t.Run("Test AND 3a", func(t *testing.T) {
		rr := NewBitmap()
		rr2 := NewBitmap()
		for k := 6 * 65536; k < 6*65536+10000; k++ {
			rr.AddInt(k)
		}
		for k := 6 * 65536; k < 6*65536+1000; k++ {
			rr2.AddInt(k)
		}
		result := And(rr, rr2)

		assert.EqualValues(t, 1000, result.GetCardinality())
	})

	t.Run("Test AND 3", func(t *testing.T) {
		var arrayand [11256]uint32
		// 393,216
		pos := 0
		rr := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.AddInt(k)
		}
		for k := 3 * 65536; k < 3*65536+1000; k++ {
			rr.AddInt(k)
		}
		for k := 3*65536 + 1000; k < 3*65536+7000; k++ {
			rr.AddInt(k)
		}
		for k := 3*65536 + 7000; k < 3*65536+9000; k++ {
			rr.AddInt(k)
		}
		for k := 4 * 65536; k < 4*65536+7000; k++ {
			rr.AddInt(k)
		}
		for k := 8 * 65536; k < 8*65536+1000; k++ {
			rr.AddInt(k)
		}
		for k := 9 * 65536; k < 9*65536+30000; k++ {
			rr.AddInt(k)
		}

		rr2 := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.AddInt(k)
			arrayand[pos] = uint32(k)
			pos++
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.AddInt(k)
			arrayand[pos] = uint32(k)
			pos++
		}
		for k := 3*65536 + 1000; k < 3*65536+7000; k++ {
			rr2.AddInt(k)
			arrayand[pos] = uint32(k)
			pos++
		}
		for k := 6 * 65536; k < 6*65536+10000; k++ {
			rr.AddInt(k)
		}
		for k := 6 * 65536; k < 6*65536+1000; k++ {
			rr2.AddInt(k)
			arrayand[pos] = uint32(k)
			pos++
		}

		for k := 7 * 65536; k < 7*65536+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 10 * 65536; k < 10*65536+5000; k++ {
			rr2.AddInt(k)
		}
		rrand := And(rr, rr2)

		arrayres := rrand.ToArray()
		ok := true
		for i := range arrayres {
			if i < len(arrayand) {
				if arrayres[i] != arrayand[i] {
					t.Log(i, arrayres[i], arrayand[i])
					ok = false
				}
			} else {
				t.Log('x', arrayres[i])
				ok = false
			}
		}

		assert.Equal(t, len(arrayres), len(arrayand))
		assert.True(t, ok)
	})

	t.Run("Test AND 4", func(t *testing.T) {
		rb := NewBitmap()
		rb2 := NewBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.AddInt(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.AddInt(i)
		}
		// TODO: Bitmap.And(bm,bm2)
		andresult := And(rb, rb2)
		off := And(rb2, rb)

		assert.True(t, andresult.Equals(off))
		assert.EqualValues(t, 0, andresult.GetCardinality())

		for i := 500000; i < 600000; i += 14 {
			rb.AddInt(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.AddInt(i)
		}
		andresult2 := And(rb, rb2)

		assert.EqualValues(t, 0, andresult.GetCardinality())
		assert.EqualValues(t, 0, andresult2.GetCardinality())

		for i := 0; i < 200000; i += 4 {
			rb.AddInt(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb.AddInt(i)
		}

		assert.EqualValues(t, 0, andresult.GetCardinality())

		rc := And(rb, rb2)
		rb.And(rb2)

		assert.Equal(t, rb.GetCardinality(), rc.GetCardinality())
	})

	t.Run("ArrayContainerCardinalityTest", func(t *testing.T) {
		ac := newArrayContainer()
		for k := uint16(0); k < 100; k++ {
			ac.iadd(k)
			assert.EqualValues(t, k+1, ac.getCardinality())
		}
		for k := uint16(0); k < 100; k++ {
			ac.iadd(k)
			assert.EqualValues(t, 100, ac.getCardinality())
		}
	})

	t.Run("or test", func(t *testing.T) {
		rr := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
		}
		rr2 := NewBitmap()
		for k := 4000; k < 8000; k++ {
			rr2.AddInt(k)
		}
		result := Or(rr, rr2)

		assert.Equal(t, rr.GetCardinality()+rr2.GetCardinality(), result.GetCardinality())
	})

	t.Run("basic test", func(t *testing.T) {
		rr := NewBitmap()
		var a [4002]uint32
		pos := 0
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
			a[pos] = uint32(k)
			pos++
		}
		rr.AddInt(100000)
		a[pos] = 100000
		pos++
		rr.AddInt(110000)
		a[pos] = 110000
		pos++
		array := rr.ToArray()
		ok := true
		for i := range a {
			if array[i] != a[i] {
				t.Log("rr : ", array[i], " a : ", a[i])
				ok = false
			}
		}

		assert.Equal(t, len(a), len(array))
		assert.True(t, ok)
	})

	t.Run("BitmapContainerCardinalityTest", func(t *testing.T) {
		ac := newBitmapContainer()
		for k := uint16(0); k < 100; k++ {
			ac.iadd(k)
			assert.EqualValues(t, k+1, ac.getCardinality())
		}
		for k := uint16(0); k < 100; k++ {
			ac.iadd(k)
			assert.EqualValues(t, 100, ac.getCardinality())
		}
	})

	t.Run("BitmapContainerTest", func(t *testing.T) {
		rr := newBitmapContainer()
		rr.iadd(uint16(110))
		rr.iadd(uint16(114))
		rr.iadd(uint16(115))
		var array [3]uint16
		pos := 0
		for itr := rr.getShortIterator(); itr.hasNext(); {
			array[pos] = itr.next()
			pos++
		}

		assert.EqualValues(t, 110, array[0])
		assert.EqualValues(t, 114, array[1])
		assert.EqualValues(t, 115, array[2])
	})

	t.Run("cardinality test", func(t *testing.T) {
		N := 1024
		for gap := 7; gap < 100000; gap *= 10 {
			for offset := 2; offset <= 1024; offset *= 2 {
				rb := NewBitmap()
				for k := 0; k < N; k++ {
					rb.AddInt(k * gap)
					assert.EqualValues(t, k+1, rb.GetCardinality())
				}

				assert.EqualValues(t, N, rb.GetCardinality())

				// check the add of existing values
				for k := 0; k < N; k++ {
					rb.AddInt(k * gap)
					assert.EqualValues(t, N, rb.GetCardinality())
				}

				rb2 := NewBitmap()

				for k := 0; k < N; k++ {
					rb2.AddInt(k * gap * offset)
					assert.EqualValues(t, k+1, rb2.GetCardinality())
				}

				assert.EqualValues(t, N, rb2.GetCardinality())

				for k := 0; k < N; k++ {
					rb2.AddInt(k * gap * offset)
					assert.EqualValues(t, N, rb2.GetCardinality())
				}

				assert.EqualValues(t, N/offset, And(rb, rb2).GetCardinality())
				assert.EqualValues(t, 2*N-2*N/offset, Xor(rb, rb2).GetCardinality())
				assert.EqualValues(t, 2*N-N/offset, Or(rb, rb2).GetCardinality())
			}
		}
	})

	t.Run("clear test", func(t *testing.T) {
		rb := NewBitmap()
		for i := 0; i < 200000; i += 7 {
			// dense
			rb.AddInt(i)
		}
		for i := 200000; i < 400000; i += 177 {
			// sparse
			rb.AddInt(i)
		}

		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := 0; i < 200000; i += 4 {
			rb2.AddInt(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.AddInt(i)
		}

		rb.Clear()

		assert.EqualValues(t, 0, rb.GetCardinality())
		assert.NotEqual(t, 0, rb2.GetCardinality())

		rb.AddInt(4)
		rb3.AddInt(4)
		andresult := And(rb, rb2)
		orresult := Or(rb, rb2)

		assert.EqualValues(t, 1, andresult.GetCardinality())
		assert.Equal(t, rb2.GetCardinality(), orresult.GetCardinality())

		for i := 0; i < 200000; i += 4 {
			rb.AddInt(i)
			rb3.AddInt(i)
		}
		for i := 200000; i < 400000; i += 114 {
			rb.AddInt(i)
			rb3.AddInt(i)
		}
		checkValidity(t, rb)
		checkValidity(t, rb2)
		checkValidity(t, rb3)
		arrayrr := rb.ToArray()
		arrayrr3 := rb3.ToArray()
		ok := true
		for i := range arrayrr {
			if arrayrr[i] != arrayrr3[i] {
				ok = false
			}
		}

		assert.Equal(t, len(arrayrr3), len(arrayrr))
		assert.True(t, ok)
	})

	t.Run("container factory ", func(t *testing.T) {
		bc1 := newBitmapContainer()
		bc2 := newBitmapContainer()
		bc3 := newBitmapContainer()
		ac1 := newArrayContainer()
		ac2 := newArrayContainer()
		ac3 := newArrayContainer()

		for i := 0; i < 5000; i++ {
			bc1.iadd(uint16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc2.iadd(uint16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc3.iadd(uint16(i * 70))
		}
		for i := 0; i < 4000; i++ {
			ac1.iadd(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac2.iadd(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac3.iadd(uint16(i * 50))
		}

		rbc := ac1.clone().(*arrayContainer).toBitmapContainer()
		validate(t, rbc, ac1)

		rbc = ac2.clone().(*arrayContainer).toBitmapContainer()
		validate(t, rbc, ac2)

		rbc = ac3.clone().(*arrayContainer).toBitmapContainer()
		validate(t, rbc, ac3)
	})

	t.Run("flipTest1 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.Flip(100000, 200000) // in-place on empty bitmap
		rbcard := rb.GetCardinality()

		assert.EqualValues(t, 100000, rbcard)

		bs := bitset.New(20000 - 10000)
		for i := uint(100000); i < 200000; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest1A", func(t *testing.T) {
		rb := NewBitmap()
		rb1 := Flip(rb, 100000, 200000)
		rbcard := rb1.GetCardinality()

		assert.EqualValues(t, 100000, rbcard)
		assert.EqualValues(t, 0, rb.GetCardinality())

		bs := bitset.New(0)
		assert.True(t, equalsBitSet(bs, rb))

		for i := uint(100000); i < 200000; i++ {
			bs.Set(i)
		}
		checkValidity(t, rb1)
		checkValidity(t, rb)
		assert.True(t, equalsBitSet(bs, rb1))
	})

	t.Run("flipTest2", func(t *testing.T) {
		rb := NewBitmap()
		rb.Flip(100000, 100000)
		rbcard := rb.GetCardinality()

		assert.EqualValues(t, 0, rbcard)

		bs := bitset.New(0)
		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest2A", func(t *testing.T) {
		rb := NewBitmap()
		rb1 := Flip(rb, 100000, 100000)

		rb.AddInt(1)
		rbcard := rb1.GetCardinality()

		assert.EqualValues(t, 0, rbcard)
		assert.EqualValues(t, 1, rb.GetCardinality())

		bs := bitset.New(0)
		assert.True(t, equalsBitSet(bs, rb1))

		bs.Set(1)
		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest3A", func(t *testing.T) {
		rb := NewBitmap()
		rb.Flip(100000, 200000) // got 100k-199999
		rb.Flip(100000, 199991) // give back 100k-199990
		rbcard := rb.GetCardinality()

		assert.EqualValues(t, 9, rbcard)

		bs := bitset.New(0)
		for i := uint(199991); i < 200000; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest4A", func(t *testing.T) {
		// fits evenly on both ends
		rb := NewBitmap()
		rb.Flip(100000, 200000) // got 100k-199999
		rb.Flip(65536, 4*65536)
		rbcard := rb.GetCardinality()

		// 65536 to 99999 are 1s
		// 200000 to 262143 are 1s: total card

		assert.EqualValues(t, 96608, rbcard)

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(200000); i < 262144; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest5", func(t *testing.T) {
		// fits evenly on small end, multiple
		// containers
		rb := NewBitmap()
		rb.Flip(100000, 132000)
		rb.Flip(65536, 120000)
		rbcard := rb.GetCardinality()

		// 65536 to 99999 are 1s
		// 120000 to 131999

		assert.EqualValues(t, 46464, rbcard)

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(120000); i < 132000; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTest6", func(t *testing.T) {
		rb := NewBitmap()
		rb1 := Flip(rb, 100000, 132000)
		rb2 := Flip(rb1, 65536, 120000)
		// rbcard := rb2.GetCardinality()

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(120000); i < 132000; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb2))
	})

	t.Run("flipTest6A", func(t *testing.T) {
		rb := NewBitmap()
		rb1 := Flip(rb, 100000, 132000)
		rb2 := Flip(rb1, 99000, 2*65536)
		rbcard := rb2.GetCardinality()

		assert.EqualValues(t, rbcard, 1928)

		bs := bitset.New(0)
		for i := uint(99000); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(2 * 65536); i < 132000; i++ {
			bs.Set(i)
		}
		assert.True(t, equalsBitSet(bs, rb2))
	})

	t.Run("flipTest7", func(t *testing.T) {
		// within 1 word, first container
		rb := NewBitmap()
		rb.Flip(650, 132000)
		rb.Flip(648, 651)
		rbcard := rb.GetCardinality()

		// 648, 649, 651-131999

		assert.EqualValues(t, rbcard, 132000-651+2)

		bs := bitset.New(0)
		bs.Set(648)
		bs.Set(649)
		for i := uint(651); i < 132000; i++ {
			bs.Set(i)
		}

		assert.True(t, equalsBitSet(bs, rb))
	})

	t.Run("flipTestBig", func(t *testing.T) {
		numCases := 1000
		rb := NewBitmap()
		bs := bitset.New(0)
		// Random r = new Random(3333);
		checkTime := 2.0

		for i := 0; i < numCases; i++ {
			start := rand.Intn(65536 * 20)
			end := rand.Intn(65536 * 20)
			if rand.Float64() < float64(0.1) {
				end = start + rand.Intn(100)
			}
			rb.Flip(uint64(start), uint64(end))
			if start < end {
				FlipRange(start, end, bs) // throws exception
			}
			// otherwise
			// insert some more ANDs to keep things sparser
			if rand.Float64() < 0.2 {
				mask := NewBitmap()
				mask1 := bitset.New(0)
				startM := rand.Intn(65536 * 20)
				endM := startM + 100000
				mask.Flip(uint64(startM), uint64(endM))
				FlipRange(startM, endM, mask1)
				mask.Flip(0, 65536*20+100000)
				FlipRange(0, 65536*20+100000, mask1)
				rb.And(mask)
				bs.InPlaceIntersection(mask1)
			}
			// see if we can detect incorrectly shared containers
			if rand.Float64() < 0.1 {
				irrelevant := Flip(rb, 10, 100000)
				irrelevant.Flip(5, 200000)
				irrelevant.Flip(190000, 260000)
			}
			if float64(i) > checkTime {
				assert.True(t, equalsBitSet(bs, rb))
				checkTime *= 1.5
			}
		}
	})

	t.Run("ortest", func(t *testing.T) {
		rr := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
		}
		rr.AddInt(100000)
		rr.AddInt(110000)
		rr2 := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr2.AddInt(k)
		}

		rror := Or(rr, rr2)

		array := rror.ToArray()

		rr.Or(rr2)
		arrayirr := rr.ToArray()

		assert.True(t, IntsEquals(array, arrayirr))
	})

	t.Run("ORtest", func(t *testing.T) {
		rr := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.AddInt(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.AddInt(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.AddInt(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.AddInt(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.AddInt(k)
		}

		rr2 := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.AddInt(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.AddInt(k)
		}
		correct := Or(rr, rr2)
		rr.Or(rr2)

		assert.True(t, correct.Equals(rr))
	})

	t.Run("ortest2", func(t *testing.T) {
		arrayrr := make([]uint32, 4000+4000+2)
		pos := 0
		rr := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr.AddInt(k)
			arrayrr[pos] = uint32(k)
			pos++
		}
		rr.AddInt(100000)
		rr.AddInt(110000)
		rr2 := NewBitmap()
		for k := 4000; k < 8000; k++ {
			rr2.AddInt(k)
			arrayrr[pos] = uint32(k)
			pos++
		}

		arrayrr[pos] = 100000
		pos++
		arrayrr[pos] = 110000
		pos++

		rror := Or(rr, rr2)

		checkValidity(t, rror)
		arrayor := rror.ToArray()

		assert.True(t, IntsEquals(arrayor, arrayrr))
	})

	t.Run("ortest3", func(t *testing.T) {
		V1 := make(map[int]bool)
		V2 := make(map[int]bool)

		rr := NewBitmap()
		rr2 := NewBitmap()
		for k := 0; k < 4000; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}
		for k := 3500; k < 4500; k++ {
			rr.AddInt(k)
			V1[k] = true
		}
		for k := 4000; k < 65000; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}

		// In the second node of each roaring bitmap, we have two bitmap
		// containers.
		// So, we will check the union between two BitmapContainers
		for k := 65536; k < 65536+10000; k++ {
			rr.AddInt(k)
			V1[k] = true
		}

		for k := 65536; k < 65536+14000; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}

		// In the 3rd node of each Roaring Bitmap, we have an
		// ArrayContainer, so, we will try the union between two
		// ArrayContainers.
		for k := 4 * 65535; k < 4*65535+1000; k++ {
			rr.AddInt(k)
			V1[k] = true
		}

		for k := 4 * 65535; k < 4*65535+800; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}

		// For the rest, we will check if the union will take them in
		// the result
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr.AddInt(k)
			V1[k] = true
		}

		for k := 7 * 65535; k < 7*65535+2000; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}

		rror := Or(rr, rr2)
		valide := true
		checkValidity(t, rror)
		for _, k := range rror.ToArray() {
			_, found := V1[int(k)]
			if !found {
				valide = false
			}
			V2[int(k)] = true
		}

		for k := range V1 {
			_, found := V2[k]
			if !found {
				valide = false
			}
		}

		assert.True(t, valide)
	})

	t.Run("ortest4", func(t *testing.T) {
		rb := NewBitmap()
		rb2 := NewBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.AddInt(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.AddInt(i)
		}
		rb2card := rb2.GetCardinality()

		// check or against an empty bitmap
		orresult := Or(rb, rb2)
		off := Or(rb2, rb)

		assert.True(t, orresult.Equals(off))
		assert.Equal(t, orresult.GetCardinality(), rb2card)

		for i := 500000; i < 600000; i += 14 {
			rb.AddInt(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.AddInt(i)
		}
		// check or against an empty bitmap
		orresult2 := Or(rb, rb2)
		checkValidity(t, orresult2)
		assert.Equal(t, orresult.GetCardinality(), rb2card)
		assert.Equal(t, rb2.GetCardinality()+rb.GetCardinality(), orresult2.GetCardinality())

		rb.Or(rb2)
		assert.True(t, rb.Equals(orresult2))
	})

	t.Run("randomTest", func(t *testing.T) {
		rTest(t, 15)
		rTest(t, 1024)
		rTest(t, 4096)
		rTest(t, 65536)
		rTest(t, 65536*16)
	})

	t.Run("SimpleCardinality", func(t *testing.T) {
		N := 512
		gap := 70

		rb := NewBitmap()
		for k := 0; k < N; k++ {
			rb.AddInt(k * gap)
			assert.EqualValues(t, k+1, rb.GetCardinality())
		}

		assert.EqualValues(t, N, rb.GetCardinality())

		for k := 0; k < N; k++ {
			rb.AddInt(k * gap)
			assert.EqualValues(t, N, rb.GetCardinality())
		}
	})

	t.Run("XORtest", func(t *testing.T) {
		rr := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.AddInt(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.AddInt(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.AddInt(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.AddInt(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.AddInt(k)
		}

		rr2 := NewBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.AddInt(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.AddInt(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.AddInt(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.AddInt(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.AddInt(k)
		}

		correct := Xor(rr, rr2)
		checkValidity(t, correct)

		rr.Xor(rr2)

		assert.True(t, correct.Equals(rr))
	})

	t.Run("xortest1", func(t *testing.T) {
		V1 := make(map[int]bool)
		V2 := make(map[int]bool)

		rr := NewBitmap()
		rr2 := NewBitmap()
		// For the first 65536: rr2 has a bitmap container, and rr has
		// an array container.
		// We will check the union between a BitmapCintainer and an
		// arrayContainer
		for k := 0; k < 4000; k++ {
			rr2.AddInt(k)
			if k < 3500 {
				V1[k] = true
			}
		}
		for k := 3500; k < 4500; k++ {
			rr.AddInt(k)
		}
		for k := 4000; k < 65000; k++ {
			rr2.AddInt(k)
			if k >= 4500 {
				V1[k] = true
			}
		}

		for k := 65536; k < 65536+30000; k++ {
			rr.AddInt(k)
		}

		for k := 65536; k < 65536+50000; k++ {
			rr2.AddInt(k)
			if k >= 65536+30000 {
				V1[k] = true
			}
		}

		// In the 3rd node of each Roaring Bitmap, we have an
		// ArrayContainer. So, we will try the union between two
		// ArrayContainers.
		for k := 4 * 65535; k < 4*65535+1000; k++ {
			rr.AddInt(k)
			if k >= (4*65535 + 800) {
				V1[k] = true
			}
		}

		for k := 4 * 65535; k < 4*65535+800; k++ {
			rr2.AddInt(k)
		}

		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr.AddInt(k)
			V1[k] = true
		}

		for k := 7 * 65535; k < 7*65535+2000; k++ {
			rr2.AddInt(k)
			V1[k] = true
		}

		rrxor := Xor(rr, rr2)
		valide := true

		for _, i := range rrxor.ToArray() {
			_, found := V1[int(i)]
			if !found {
				valide = false
			}
			V2[int(i)] = true
		}
		for k := range V1 {
			_, found := V2[k]
			if !found {
				valide = false
			}
		}

		assert.True(t, valide)
	})

	t.Run("ToExistingArray-Test", func(t *testing.T) {
		values := make([]uint32, 0, 110)
		rb := NewBitmap()

		for i := 10; i < 120; i++ {
			values = append(values, uint32(i))
		}
		rb.AddMany(values)
		assert.Equal(t, values, rb.ToArray())
		existing := make([]uint32, len(values))
		buf := rb.ToExistingArray(&existing)
		assert.Equal(t, values, *buf)
	})
}

func TestXORtest4(t *testing.T) {
	t.Run("XORtest 4", func(t *testing.T) {
		rb := NewBitmap()
		rb2 := NewBitmap()
		counter := 0

		for i := 0; i < 200000; i += 4 {
			rb2.AddInt(i)
			counter++
		}

		assert.EqualValues(t, counter, rb2.GetCardinality())

		for i := 200000; i < 400000; i += 14 {
			rb2.AddInt(i)
			counter++
		}

		assert.EqualValues(t, counter, rb2.GetCardinality())

		rb2card := rb2.GetCardinality()
		assert.EqualValues(t, counter, rb2card)

		// check or against an empty bitmap
		xorresult := Xor(rb, rb2)
		assert.EqualValues(t, counter, xorresult.GetCardinality())
		off := Or(rb2, rb)

		assert.EqualValues(t, counter, off.GetCardinality())
		assert.True(t, xorresult.Equals(off))

		assert.Equal(t, xorresult.GetCardinality(), rb2card)
		for i := 500000; i < 600000; i += 14 {
			rb.AddInt(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.AddInt(i)
		}
		// check or against an empty bitmap
		xorresult2 := Xor(rb, rb2)

		assert.EqualValues(t, xorresult.GetCardinality(), rb2card)
		assert.Equal(t, xorresult2.GetCardinality(), rb2.GetCardinality()+rb.GetCardinality())

		rb.Xor(rb2)
		assert.True(t, xorresult2.Equals(rb))
	})
	// need to add the massives
}

func TestNextMany(t *testing.T) {
	count := 70000

	for _, gap := range []uint32{1, 8, 32, 128} {
		expected := make([]uint32, count)
		{
			v := uint32(0)
			for i := range expected {
				expected[i] = v
				v += gap
			}
		}
		bm := BitmapOf(expected...)
		for _, bufSize := range []int{1, 64, 4096, count} {
			buf := make([]uint32, bufSize)
			it := bm.ManyIterator()
			cur := 0
			for n := it.NextMany(buf); n != 0; n = it.NextMany(buf) {
				// much faster tests... (10s -> 5ms)
				if cur+n > count {
					assert.LessOrEqual(t, count, cur+n)
				}

				for i, v := range buf[:n] {
					// much faster tests...
					if v != expected[cur+i] {
						assert.Equal(t, expected[cur+i], v)
					}
				}

				cur += n
			}

			assert.Equal(t, count, cur)
		}
	}
}

func TestBigRandom(t *testing.T) {
	rTest(t, 15)
	rTest(t, 100)
	rTest(t, 512)
	rTest(t, 1023)
	rTest(t, 1025)
	rTest(t, 4095)
	rTest(t, 4096)
	rTest(t, 4097)
	rTest(t, 65536)
	rTest(t, 65536*16)
}

func TestHash(t *testing.T) {
	hashTest(t, 15)
	hashTest(t, 100)
	hashTest(t, 512)
	hashTest(t, 1023)
	hashTest(t, 1025)
	hashTest(t, 4095)
	hashTest(t, 4096)
	hashTest(t, 4097)
}

func rTest(t *testing.T, N int) {
	for gap := 1; gap <= 65536; gap *= 2 {
		bs1 := bitset.New(0)
		rb1 := NewBitmap()
		for x := 0; x <= N; x += gap {
			bs1.Set(uint(x))
			rb1.AddInt(x)
		}

		assert.EqualValues(t, rb1.GetCardinality(), bs1.Count())
		assert.True(t, equalsBitSet(bs1, rb1))

		for offset := 1; offset <= gap; offset *= 2 {
			bs2 := bitset.New(0)
			rb2 := NewBitmap()
			for x := 0; x <= N; x += gap {
				bs2.Set(uint(x + offset))
				rb2.AddInt(x + offset)
			}

			assert.EqualValues(t, rb2.GetCardinality(), bs2.Count())
			assert.True(t, equalsBitSet(bs2, rb2))

			clonebs1 := bs1.Clone()
			clonebs1.InPlaceIntersection(bs2)

			if !equalsBitSet(clonebs1, And(rb1, rb2)) {
				v := rb1.Clone()
				v.And(rb2)
				assert.True(t, equalsBitSet(clonebs1, v))
			}

			// testing OR
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceUnion(bs2)

			assert.True(t, equalsBitSet(clonebs1, Or(rb1, rb2)))

			// testing XOR
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceSymmetricDifference(bs2)

			assert.True(t, equalsBitSet(clonebs1, Xor(rb1, rb2)))

			// testing NOTAND
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceDifference(bs2)

			assert.True(t, equalsBitSet(clonebs1, AndNot(rb1, rb2)))
		}
	}
}

func equalsBitSet(a *bitset.BitSet, b *Bitmap) bool {
	for i, e := a.NextSet(0); e; i, e = a.NextSet(i + 1) {
		if !b.ContainsInt(int(i)) {
			return false
		}
	}
	i := b.Iterator()
	for i.HasNext() {
		if !a.Test(uint(i.Next())) {
			return false
		}
	}
	return true
}

func equalsArray(a []int, b *Bitmap) bool {
	if uint64(len(a)) != b.GetCardinality() {
		return false
	}
	for _, x := range a {
		if !b.ContainsInt(x) {
			return false
		}
	}
	return true
}

func IntsEquals(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func validate(t *testing.T, bc *bitmapContainer, ac *arrayContainer) {
	// Checking the cardinalities of each container
	t.Helper()

	if bc.getCardinality() != ac.getCardinality() {
		t.Error("cardinality differs")
	}
	// Checking that the two containers contain the same values
	counter := 0

	for i := bc.NextSetBit(0); i >= 0; i = bc.NextSetBit(uint(i) + 1) {
		counter++
		if !ac.contains(uint16(i)) {
			t.Log("content differs")
			t.Log(bc)
			t.Log(ac)
			t.Fail()
		}

	}

	// checking the cardinality of the BitmapContainer
	assert.Equal(t, counter, bc.getCardinality())
}

func TestRoaringArray(t *testing.T) {
	a := newRoaringArray()

	t.Run("Test Init", func(t *testing.T) {
		assert.Equal(t, 0, a.size())
	})

	t.Run("Test Insert", func(t *testing.T) {
		a.appendContainer(0, newArrayContainer(), false)
		assert.Equal(t, 1, a.size())
	})

	t.Run("Test Remove", func(t *testing.T) {
		a.remove(0)
		assert.Equal(t, 0, a.size())
	})

	t.Run("Test popcount Full", func(t *testing.T) {
		res := popcount(uint64(0xffffffffffffffff))
		assert.EqualValues(t, 64, res)
	})

	t.Run("Test popcount Empty", func(t *testing.T) {
		res := popcount(0)
		assert.EqualValues(t, 0, res)
	})

	t.Run("Test popcount 16", func(t *testing.T) {
		res := popcount(0xff00ff)
		assert.EqualValues(t, 16, res)
	})

	t.Run("Test ArrayContainer Add", func(t *testing.T) {
		ar := newArrayContainer()
		ar.iadd(1)

		assert.EqualValues(t, 1, ar.getCardinality())
	})

	t.Run("Test ArrayContainer Add wacky", func(t *testing.T) {
		ar := newArrayContainer()
		ar.iadd(0)
		ar.iadd(5000)

		assert.EqualValues(t, 2, ar.getCardinality())
	})

	t.Run("Test ArrayContainer Add Reverse", func(t *testing.T) {
		ar := newArrayContainer()
		ar.iadd(5000)
		ar.iadd(2048)
		ar.iadd(0)

		assert.EqualValues(t, 3, ar.getCardinality())
	})

	t.Run("Test BitmapContainer Add ", func(t *testing.T) {
		bm := newBitmapContainer()
		bm.iadd(0)

		assert.EqualValues(t, 1, bm.getCardinality())
	})
}

func TestFlipBigA(t *testing.T) {
	numCases := 1000
	bs := bitset.New(0)
	checkTime := 2.0
	rb1 := NewBitmap()
	rb2 := NewBitmap()

	for i := 0; i < numCases; i++ {
		start := rand.Intn(65536 * 20)
		end := rand.Intn(65536 * 20)
		if rand.Float64() < 0.1 {
			end = start + rand.Intn(100)
		}

		if (i & 1) == 0 {
			rb2 = FlipInt(rb1, start, end)
			// tweak the other, catch bad sharing
			rb1.FlipInt(rand.Intn(65536*20), rand.Intn(65536*20))
		} else {
			rb1 = FlipInt(rb2, start, end)
			rb2.FlipInt(rand.Intn(65536*20), rand.Intn(65536*20))
		}

		if start < end {
			FlipRange(start, end, bs) // throws exception
		}
		// otherwise
		// insert some more ANDs to keep things sparser
		if (rand.Float64() < 0.2) && (i&1) == 0 {
			mask := NewBitmap()
			mask1 := bitset.New(0)
			startM := rand.Intn(65536 * 20)
			endM := startM + 100000
			mask.FlipInt(startM, endM)
			FlipRange(startM, endM, mask1)
			mask.FlipInt(0, 65536*20+100000)
			FlipRange(0, 65536*20+100000, mask1)
			rb2.And(mask)
			bs.InPlaceIntersection(mask1)
		}

		if float64(i) > checkTime {
			var rb *Bitmap

			if (i & 1) == 0 {
				rb = rb2
			} else {
				rb = rb1
			}

			assert.True(t, equalsBitSet(bs, rb))
			checkTime *= 1.5
		}
	}
}

func TestNextManyOfAddRangeAcrossContainers(t *testing.T) {
	rb := NewBitmap()
	rb.AddRange(65530, 65540)
	expectedCard := 10
	expected := []uint32{65530, 65531, 65532, 65533, 65534, 65535, 65536, 65537, 65538, 65539, 0}

	// test where all values can be returned in a single buffer
	it := rb.ManyIterator()
	buf := make([]uint32, 11)
	n := it.NextMany(buf)

	assert.Equal(t, expectedCard, n)

	for i, e := range expected {
		assert.Equal(t, e, buf[i])
	}

	// test where buf is size 1, so many iterations
	it = rb.ManyIterator()
	n = 0
	buf = make([]uint32, 1)

	for i := 0; i < expectedCard; i++ {
		n = it.NextMany(buf)

		assert.Equal(t, 1, n)
		assert.Equal(t, expected[i], buf[0])
	}

	n = it.NextMany(buf)
	assert.Equal(t, 0, n)
}

func TestDoubleAdd(t *testing.T) {
	t.Run("doubleadd ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536)
		rb.AddRange(65530, 65536)
		rb2 := NewBitmap()
		rb2.AddRange(65530, 65536)

		assert.True(t, rb.Equals(rb2))

		rb2.RemoveRange(65530, 65536)

		assert.EqualValues(t, 0, rb2.GetCardinality())
	})

	t.Run("doubleadd2 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20)
		rb.AddRange(65530, 65536*20)
		rb2 := NewBitmap()
		rb2.AddRange(65530, 65536*20)

		assert.True(t, rb.Equals(rb2))

		rb2.RemoveRange(65530, 65536*20)

		assert.EqualValues(t, 0, rb2.GetCardinality())
	})

	t.Run("doubleadd3 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20+10)
		rb.AddRange(65530, 65536*20+10)
		rb2 := NewBitmap()
		rb2.AddRange(65530, 65536*20+10)

		assert.True(t, rb.Equals(rb2))

		rb2.RemoveRange(65530, 65536*20+1)

		assert.EqualValues(t, 9, rb2.GetCardinality())
	})

	t.Run("doubleadd4 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20)
		rb.RemoveRange(65533+5, 65536*20)

		assert.EqualValues(t, 5, rb.GetCardinality())
	})

	t.Run("doubleadd5 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20)
		rb.RemoveRange(65533+5, 65536*20-5)

		assert.EqualValues(t, 10, rb.GetCardinality())
	})

	t.Run("doubleadd6 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20-5)
		rb.RemoveRange(65533+5, 65536*20-10)

		assert.EqualValues(t, 10, rb.GetCardinality())
	})

	t.Run("doubleadd7 ", func(t *testing.T) {
		rb := NewBitmap()
		rb.AddRange(65533, 65536*20+1)
		rb.RemoveRange(65533+1, 65536*20)

		assert.EqualValues(t, 2, rb.GetCardinality())
	})

	t.Run("AndNotBug01 ", func(t *testing.T) {
		rb1 := NewBitmap()
		rb1.AddRange(0, 60000)
		rb2 := NewBitmap()
		rb2.AddRange(60000-10, 60000+10)
		rb2.AndNot(rb1)
		rb3 := NewBitmap()
		rb3.AddRange(60000, 60000+10)

		assert.True(t, rb2.Equals(rb3))
	})
}

func TestAndNot(t *testing.T) {
	rr := NewBitmap()

	for k := 4000; k < 4256; k++ {
		rr.AddInt(k)
	}
	for k := 65536; k < 65536+4000; k++ {
		rr.AddInt(k)
	}
	for k := 3 * 65536; k < 3*65536+9000; k++ {
		rr.AddInt(k)
	}
	for k := 4 * 65535; k < 4*65535+7000; k++ {
		rr.AddInt(k)
	}
	for k := 6 * 65535; k < 6*65535+10000; k++ {
		rr.AddInt(k)
	}
	for k := 8 * 65535; k < 8*65535+1000; k++ {
		rr.AddInt(k)
	}
	for k := 9 * 65535; k < 9*65535+30000; k++ {
		rr.AddInt(k)
	}

	rr2 := NewBitmap()

	for k := 4000; k < 4256; k++ {
		rr2.AddInt(k)
	}
	for k := 65536; k < 65536+4000; k++ {
		rr2.AddInt(k)
	}
	for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
		rr2.AddInt(k)
	}
	for k := 6 * 65535; k < 6*65535+1000; k++ {
		rr2.AddInt(k)
	}
	for k := 7 * 65535; k < 7*65535+1000; k++ {
		rr2.AddInt(k)
	}
	for k := 10 * 65535; k < 10*65535+5000; k++ {
		rr2.AddInt(k)
	}

	correct := AndNot(rr, rr2)
	rr.AndNot(rr2)

	assert.True(t, correct.Equals(rr))
}

func TestStats(t *testing.T) {
	t.Run("Test Stats with empty bitmap", func(t *testing.T) {
		expectedStats := Statistics{}
		rr := NewBitmap()

		assert.EqualValues(t, expectedStats, rr.Stats())
	})

	t.Run("Test Stats with bitmap Container", func(t *testing.T) {
		// Given a bitmap that should have a single bitmap container
		expectedStats := Statistics{
			Cardinality: 60000,
			Containers:  1,

			BitmapContainers:      1,
			BitmapContainerValues: 60000,
			BitmapContainerBytes:  8192,

			RunContainers:      0,
			RunContainerBytes:  0,
			RunContainerValues: 0,
		}

		rr := NewBitmap()

		for i := uint32(0); i < 60000; i++ {
			rr.Add(i)
		}

		assert.EqualValues(t, expectedStats, rr.Stats())
	})

	t.Run("Test Stats with Array Container", func(t *testing.T) {
		// Given a bitmap that should have a single array container
		expectedStats := Statistics{
			Cardinality: 2,
			Containers:  1,

			ArrayContainers:      1,
			ArrayContainerValues: 2,
			ArrayContainerBytes:  4,
		}
		rr := NewBitmap()
		rr.Add(2)
		rr.Add(4)

		assert.EqualValues(t, expectedStats, rr.Stats())
	})
}

func TestFlipVerySmall(t *testing.T) {
	rb := NewBitmap()
	rb.Flip(0, 10) // got [0,9], card is 10
	rb.Flip(0, 1)  // give back the number 0, card goes to 9
	rbcard := rb.GetCardinality()

	assert.EqualValues(t, 9, rbcard)
}

func TestReverseIterator(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		values := []uint32{0, 2, 15, 16, 31, 32, 33, 9999, MaxUint16, MaxUint32}
		bm := New()
		for n := 0; n < len(values); n++ {
			bm.Add(values[n])
		}
		i := bm.ReverseIterator()
		n := len(values) - 1

		for i.HasNext() {
			assert.EqualValues(t, i.Next(), values[n])
			n--
		}

		// HasNext() was terminating early - add test
		i = bm.ReverseIterator()
		n = len(values) - 1
		for ; n >= 0; n-- {
			assert.EqualValues(t, i.Next(), values[n])
			assert.False(t, n > 0 && !i.HasNext())
		}
	})

	t.Run("#2", func(t *testing.T) {
		bm := New()
		i := bm.ReverseIterator()

		assert.False(t, i.HasNext())
	})

	t.Run("#3", func(t *testing.T) {
		bm := New()
		bm.AddInt(0)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 0, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#4", func(t *testing.T) {
		bm := New()
		bm.AddInt(9999)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 9999, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#5", func(t *testing.T) {
		bm := New()
		bm.AddInt(MaxUint16)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, MaxUint16, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#6", func(t *testing.T) {
		bm := New()
		bm.Add(MaxUint32)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, uint32(MaxUint32), i.Next())
		assert.False(t, i.HasNext())
	})
}

func TestIteratorPeekNext(t *testing.T) {
	values := []uint32{0, 2, 15, 16, 31, 32, 33, 9999, MaxUint16, MaxUint32}
	bm := New()

	for n := 0; n < len(values); n++ {
		bm.Add(values[n])
	}

	i := bm.Iterator()
	assert.True(t, i.HasNext())

	for i.HasNext() {
		assert.Equal(t, i.PeekNext(), i.Next())
	}
}

func TestIteratorAdvance(t *testing.T) {
	values := []uint32{1, 2, 15, 16, 31, 32, 33, 9999, MaxUint16}
	bm := New()

	for n := 0; n < len(values); n++ {
		bm.Add(values[n])
	}

	cases := []struct {
		minval   uint32
		expected uint32
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 15},
		{30, 31},
		{33, 33},
		{9998, 9999},
		{MaxUint16, MaxUint16},
	}

	t.Run("advance by using a new int iterator", func(t *testing.T) {
		for _, c := range cases {
			i := bm.Iterator()
			i.AdvanceIfNeeded(c.minval)

			assert.True(t, i.HasNext())
			assert.Equal(t, c.expected, i.PeekNext())
		}
	})

	t.Run("advance by using the same int iterator", func(t *testing.T) {
		i := bm.Iterator()

		for _, c := range cases {
			i.AdvanceIfNeeded(c.minval)

			assert.True(t, i.HasNext())
			assert.Equal(t, c.expected, i.PeekNext())
		}
	})

	t.Run("advance out of a container value", func(t *testing.T) {
		i := bm.Iterator()

		i.AdvanceIfNeeded(MaxUint32)
		assert.False(t, i.HasNext())

		i.AdvanceIfNeeded(MaxUint32)
		assert.False(t, i.HasNext())
	})

	t.Run("advance on a value that is less than the pointed value", func(t *testing.T) {
		i := bm.Iterator()
		i.AdvanceIfNeeded(29)

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 31, i.PeekNext())

		i.AdvanceIfNeeded(13)

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 31, i.PeekNext())
	})
}

func TestPackageFlipMaxRangeEnd(t *testing.T) {
	var empty Bitmap
	flipped := Flip(&empty, 0, MaxRange)

	assert.EqualValues(t, MaxRange, flipped.GetCardinality())
}

func TestBitmapFlipMaxRangeEnd(t *testing.T) {
	var bm Bitmap
	bm.Flip(0, MaxRange)

	assert.EqualValues(t, MaxRange, bm.GetCardinality())
}

func TestIterate(t *testing.T) {
	rb := NewBitmap()

	for i := 0; i < 300; i++ {
		rb.Add(uint32(i))
	}

	var values []uint32
	rb.Iterate(func(x uint32) bool {
		values = append(values, x)
		return true
	})

	assert.Equal(t, rb.ToArray(), values)
}

func TestIterateCompressed(t *testing.T) {
	rb := NewBitmap()

	for i := 0; i < 300; i++ {
		rb.Add(uint32(i))
	}

	rb.RunOptimize()

	var values []uint32
	rb.Iterate(func(x uint32) bool {
		values = append(values, x)
		return true
	})

	assert.Equal(t, rb.ToArray(), values)
}

func TestIterateLargeValues(t *testing.T) {
	rb := NewBitmap()

	// This range of values ensures that all different types of containers will be used
	for i := 150000; i < 450000; i++ {
		rb.Add(uint32(i))
	}

	var values []uint32
	rb.Iterate(func(x uint32) bool {
		values = append(values, x)
		return true
	})

	assert.Equal(t, rb.ToArray(), values)
}

func TestIterateHalt(t *testing.T) {
	rb := NewBitmap()

	// This range of values ensures that all different types of containers will be used
	for i := 150000; i < 450000; i++ {
		rb.Add(uint32(i))
	}

	var values []uint32
	count := uint64(0)
	stopAt := rb.GetCardinality() - 1
	rb.Iterate(func(x uint32) bool {
		values = append(values, x)
		count++
		if count == stopAt {
			return false
		}
		return true
	})

	expected := rb.ToArray()
	expected = expected[0 : len(expected)-1]
	assert.Equal(t, expected, values)
}

func testDense(fn func(string, *Bitmap)) {
	bc := New()
	for i := 0; i <= arrayDefaultMaxSize; i++ {
		bc.Add(uint32(1 + MaxUint16 + i*2))
	}

	rc := New()
	rc.AddRange(1, 2)
	rc.AddRange(bc.GetCardinality(), bc.GetCardinality()*2)

	ac := New()
	for i := 1; i <= arrayDefaultMaxSize; i++ {
		ac.Add(uint32(MaxUint16 + i*2))
	}

	brc := New()
	for i := 150000; i < 450000; i++ {
		brc.Add(uint32(i))
	}

	for _, tc := range []struct {
		name string
		rb   *Bitmap
	}{
		{"bitmap", bc},
		{"run", rc},
		{"array", ac},
		{"bitmaps-and-runs", brc},
	} {
		fn(tc.name+"-"+strconv.FormatUint(tc.rb.GetCardinality(), 10), tc.rb)
	}
}

func TestToDense(t *testing.T) {
	testDense(func(name string, rb *Bitmap) {
		t.Run(name, func(t *testing.T) {
			bm := bitset.From(rb.ToDense())
			assert.EqualValues(t, rb.GetCardinality(), uint64(bm.Count()))
			rb.Iterate(func(x uint32) bool {
				return assert.True(t, bm.Test(uint(x)), "value %d should be set", x)
			})
		})
	})
}

func TestFromDense(t *testing.T) {
	testDense(func(name string, rb *Bitmap) {
		for _, doCopy := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s,doCopy=%t", name, doCopy), func(t *testing.T) {
				dense := rb.ToDense()
				cp := FromDense(dense, doCopy)
				if doCopy {
					// Clear the original dense slice to ensure we don't have any
					// references to it
					for i := range dense {
						dense[i] = 0
					}
				}
				assert.True(t, rb.Equals(cp))
			})
		}
	})
}

func TestFromBitSet(t *testing.T) {
	testDense(func(name string, rb *Bitmap) {
		t.Run(fmt.Sprintf("%s", name), func(t *testing.T) {
			dense := rb.ToBitSet()
			cp := FromBitSet(dense)
			assert.True(t, rb.Equals(cp))
		})
	})
}

func TestRoaringArrayValidation(t *testing.T) {
	a := newRoaringArray()

	a.keys = append(a.keys, uint16(3), uint16(1))
	assert.ErrorIs(t, a.validate(), ErrKeySortOrder)
	a.clear()

	// build up cardinality coherent arrays
	a.keys = append(a.keys, uint16(1), uint16(3), uint16(10))
	assert.ErrorIs(t, a.validate(), ErrCardinalityConstraint)
	a.containers = append(a.containers, &runContainer16{}, &runContainer16{}, &runContainer16{})
	assert.ErrorIs(t, a.validate(), ErrCardinalityConstraint)
	a.needCopyOnWrite = append(a.needCopyOnWrite, true, false, true)
	assert.ErrorIs(t, a.validate(), ErrRunIntervalsEmpty)
}

func TestBitMapValidation(t *testing.T) {
	bm := NewBitmap()
	bm.AddRange(0, 100)
	bm.AddRange(306, 406)
	bm.AddRange(102, 202)
	bm.AddRange(204, 304)
	assert.NoError(t, bm.Validate())

	randomEntries := make([]uint32, 0, 1000)
	for i := 0; i < 1000; i++ {
		randomEntries = append(randomEntries, rand.Uint32())
	}

	bm.AddMany(randomEntries)
	assert.NoError(t, bm.Validate())

	randomEntries = make([]uint32, 0, 1000)
	for i := 0; i < 1000; i++ {
		randomEntries = append(randomEntries, uint32(i))
	}
	bm.AddMany(randomEntries)
	assert.NoError(t, bm.Validate())
}

func TestBitMapValidationFromDeserialization(t *testing.T) {
	// To understand what is going on here, read https://github.com/RoaringBitmap/RoaringFormatSpec
	// Maintainers: The loader and corruptor are dependent on one another
	// The tests expect a certain size, with values at certain location.
	// The tests are geared toward single byte corruption.

	// There is no way to test Bitmap container corruption once the bitmap is deserialzied

	deserializationTests := []struct {
		name      string
		loader    func(bm *Bitmap)
		corruptor func(s []byte)
		err       error
	}{
		{
			name: "Corrupts Run Length vs Num Runs",
			loader: func(bm *Bitmap) {
				bm.AddRange(0, 2)
				bm.AddRange(4, 6)
				bm.AddRange(8, 100)
			},
			corruptor: func(s []byte) {
				// 21 is the length of the run of the last run/range
				// Shortening causes interval sum to be to short
				s[21] = 1
			},
			err: ErrRunIntervalSize,
		},
		{
			name: "Corrupts Run Length",
			loader: func(bm *Bitmap) {
				bm.AddRange(100, 110)
			},
			corruptor: func(s []byte) {
				s[13] = 0
			},
			err: ErrRunIntervalSize,
		},
		{
			name: "Creates Interval Overlap",
			loader: func(bm *Bitmap) {
				bm.AddRange(100, 110)
				bm.AddRange(115, 125)
			},
			corruptor: func(s []byte) {
				// sets the start of the second run
				// Creates overlapping intervals
				s[15] = 108
			},
			err: ErrRunIntervalOverlap,
		},
		{
			name: "Break Array Sort Order",
			loader: func(bm *Bitmap) {
				arrayEntries := make([]uint32, 0, 10)
				for i := 0; i < 10; i++ {
					arrayEntries = append(arrayEntries, uint32(i))
				}
				bm.AddMany(arrayEntries)
			},
			corruptor: func(s []byte) {
				// breaks the sort order
				s[34] = 0
			},
			err: ErrArrayIncorrectSort,
		},
	}

	for _, tt := range deserializationTests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
				}
			}()

			bm := NewBitmap()
			tt.loader(bm)
			assert.NoError(t, bm.Validate())
			serialized, err := bm.ToBytes()
			assert.NoError(t, err)
			tt.corruptor(serialized)
			corruptedDeserializedBitMap := NewBitmap()
			corruptedDeserializedBitMap.ReadFrom(bytes.NewReader(serialized))
			// Check that Validate() returns nil if and only if tt.err is nil
			validateErr := corruptedDeserializedBitMap.Validate()
			if tt.err == nil {
				assert.NoError(t, validateErr, "expected validation to succeed when tt.err is nil")
			} else {
				assert.Error(t, validateErr, "expected validation to fail when tt.err is not nil")
			}

			corruptedDeserializedBitMap = NewBitmap()
			corruptedDeserializedBitMap.MustReadFrom(bytes.NewReader(serialized))
			// We will never hit this because of the recover
			t.Errorf("did not panic")
		})
	}
}

func TestNextAndPreviousValue(t *testing.T) {
	t.Run("Java Regression1 ", func(t *testing.T) {
		// [Java1] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3645
		bmp := New()
		bmp.AddRange(64, 129)
		assert.Equal(t, int64(64), bmp.NextValue(64))
		assert.Equal(t, int64(64), bmp.NextValue(0))
		assert.Equal(t, int64(64), bmp.NextValue(64))
		assert.Equal(t, int64(65), bmp.NextValue(65))
		assert.Equal(t, int64(128), bmp.NextValue(128))
		assert.Equal(t, int64(-1), bmp.NextValue(129))

		assert.Equal(t, int64(-1), bmp.PreviousValue(0))
		assert.Equal(t, int64(-1), bmp.PreviousValue(63))
		assert.Equal(t, int64(64), bmp.PreviousValue(64))
		assert.Equal(t, int64(65), bmp.PreviousValue(65))
		assert.Equal(t, int64(128), bmp.PreviousValue(128))
		assert.Equal(t, int64(128), bmp.PreviousValue(129))

		assert.Equal(t, int64(0), bmp.NextAbsentValue(0))
		assert.Equal(t, int64(63), bmp.NextAbsentValue(63))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(64))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(65))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(128))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(129))

		assert.Equal(t, int64(0), bmp.PreviousAbsentValue(0))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(63))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(64))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(65))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(128))
	})
	t.Run("Java Regression2", func(t *testing.T) {
		// [Java2] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3655

		bmp := New()
		bmp.AddRange(64, 129)
		bmp.AddRange(256, 256+64+1)
		assert.Equal(t, int64(64), bmp.NextValue(0))
		assert.Equal(t, int64(64), bmp.NextValue(64))
		assert.Equal(t, int64(65), bmp.NextValue(65))
		assert.Equal(t, int64(128), bmp.NextValue(128))
		assert.Equal(t, int64(256), bmp.NextValue(129))
		assert.Equal(t, int64(-1), bmp.NextValue(512))

		assert.Equal(t, int64(-1), bmp.PreviousValue(0))
		assert.Equal(t, int64(-1), bmp.PreviousValue(63))
		assert.Equal(t, int64(64), bmp.PreviousValue(64))
		assert.Equal(t, int64(65), bmp.PreviousValue(65))
		assert.Equal(t, int64(128), bmp.PreviousValue(128))
		assert.Equal(t, int64(128), bmp.PreviousValue(129))
		assert.Equal(t, int64(128), bmp.PreviousValue(199))
		assert.Equal(t, int64(128), bmp.PreviousValue(200))
		assert.Equal(t, int64(128), bmp.PreviousValue(250))
		assert.Equal(t, int64(256), bmp.PreviousValue(256))
		assert.Equal(t, int64(320), bmp.PreviousValue(2500))

		assert.Equal(t, int64(0), bmp.NextAbsentValue(0))
		assert.Equal(t, int64(63), bmp.NextAbsentValue(63))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(64))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(65))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(128))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(129))
		assert.Equal(t, int64(199), bmp.NextAbsentValue(199))
		assert.Equal(t, int64(200), bmp.NextAbsentValue(200))
		assert.Equal(t, int64(250), bmp.NextAbsentValue(250))
		assert.Equal(t, int64(321), bmp.NextAbsentValue(256))
		assert.Equal(t, int64(321), bmp.NextAbsentValue(320))

		assert.Equal(t, int64(0), bmp.PreviousAbsentValue(0))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(63))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(64))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(65))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(128))
		assert.Equal(t, int64(129), bmp.PreviousAbsentValue(129))
		assert.Equal(t, int64(199), bmp.PreviousAbsentValue(199))
		assert.Equal(t, int64(200), bmp.PreviousAbsentValue(200))
		assert.Equal(t, int64(250), bmp.PreviousAbsentValue(250))
		assert.Equal(t, int64(255), bmp.PreviousAbsentValue(256))
		assert.Equal(t, int64(255), bmp.PreviousAbsentValue(300))
		assert.Equal(t, int64(500), bmp.PreviousAbsentValue(500))
		assert.Equal(t, int64(501), bmp.PreviousAbsentValue(501))
	})

	t.Run("Java Regression3", func(t *testing.T) {
		// [Java3] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3666

		bmp := New()
		bmp.AddRange(64, 129)
		bmp.AddRange(200, 200+300+1)
		bmp.AddRange(5000, 5000+200+1)
		assert.Equal(t, int64(64), bmp.NextValue(0))
		assert.Equal(t, int64(64), bmp.NextValue(63))
		assert.Equal(t, int64(64), bmp.NextValue(64))
		assert.Equal(t, int64(65), bmp.NextValue(65))
		assert.Equal(t, int64(128), bmp.NextValue(128))
		assert.Equal(t, int64(200), bmp.NextValue(129))
		assert.Equal(t, int64(200), bmp.NextValue(199))
		assert.Equal(t, int64(200), bmp.NextValue(200))
		assert.Equal(t, int64(250), bmp.NextValue(250))
		assert.Equal(t, int64(5000), bmp.NextValue(2500))
		assert.Equal(t, int64(5000), bmp.NextValue(5000))
		assert.Equal(t, int64(5200), bmp.NextValue(5200))
		assert.Equal(t, int64(-1), bmp.NextValue(5201))

		assert.Equal(t, int64(-1), bmp.PreviousValue(0))
		assert.Equal(t, int64(-1), bmp.PreviousValue(63))
		assert.Equal(t, int64(64), bmp.PreviousValue(64))
		assert.Equal(t, int64(65), bmp.PreviousValue(65))
		assert.Equal(t, int64(128), bmp.PreviousValue(128))
		assert.Equal(t, int64(128), bmp.PreviousValue(129))
		assert.Equal(t, int64(128), bmp.PreviousValue(199))
		assert.Equal(t, int64(200), bmp.PreviousValue(200))
		assert.Equal(t, int64(250), bmp.PreviousValue(250))
		assert.Equal(t, int64(500), bmp.PreviousValue(2500))
		assert.Equal(t, int64(5000), bmp.PreviousValue(5000))
		assert.Equal(t, int64(5200), bmp.PreviousValue(5200))
		assert.Equal(t, int64(5200), bmp.PreviousValue(5201))

		assert.Equal(t, int64(0), bmp.NextAbsentValue(0))
		assert.Equal(t, int64(63), bmp.NextAbsentValue(63))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(64))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(65))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(128))
		assert.Equal(t, int64(129), bmp.NextAbsentValue(129))
		assert.Equal(t, int64(199), bmp.NextAbsentValue(199))
		assert.Equal(t, int64(501), bmp.NextAbsentValue(200))
		assert.Equal(t, int64(501), bmp.NextAbsentValue(250))
		assert.Equal(t, int64(2500), bmp.NextAbsentValue(2500))
		assert.Equal(t, int64(5201), bmp.NextAbsentValue(5000))
		assert.Equal(t, int64(5201), bmp.NextAbsentValue(5200))
		assert.Equal(t, int64(5201), bmp.NextAbsentValue(5201))

		assert.Equal(t, int64(0), bmp.PreviousAbsentValue(0))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(63))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(64))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(65))
		assert.Equal(t, int64(63), bmp.PreviousAbsentValue(128))
		assert.Equal(t, int64(129), bmp.PreviousAbsentValue(129))
		assert.Equal(t, int64(199), bmp.PreviousAbsentValue(199))
		assert.Equal(t, int64(199), bmp.PreviousAbsentValue(200))
		assert.Equal(t, int64(199), bmp.PreviousAbsentValue(250))
		assert.Equal(t, int64(2500), bmp.PreviousAbsentValue(2500))
		assert.Equal(t, int64(4999), bmp.PreviousAbsentValue(5000))
		assert.Equal(t, int64(4999), bmp.PreviousAbsentValue(5200))
		assert.Equal(t, int64(5201), bmp.PreviousAbsentValue(5201))
	})

	t.Run("skip odd ", func(t *testing.T) {
		bmp := New()
		for i := 0; i < 2000; i++ {
			bmp.Add(uint32(i * 2))
		}
		for i := 0; i < 2000; i++ {
			assert.Equal(t, int64(i*2), bmp.NextValue(uint32(i*2)))
			assert.Equal(t, int64(i*2), bmp.PreviousValue(uint32(i*2)))
			assert.Equal(t, int64(i*2+1), bmp.NextAbsentValue(uint32(i*2+1)))

			if i != 0 {
				assert.Equal(t, int64(i*2-1), bmp.PreviousAbsentValue(uint32(i*2)))
			}
		}
	})

	t.Run("Absent target container", func(t *testing.T) {
		bmp := BitmapOf(2, 3, 131072, MaxUint32)

		assert.Equal(t, int64(3), bmp.PreviousValue(65536))
		assert.Equal(t, int64(131072), bmp.PreviousValue(MaxUint32>>1))
		assert.Equal(t, int64(131072), bmp.PreviousValue(MaxUint32-131071))

		bmp = BitmapOf(131072)
		assert.Equal(t, int64(-1), bmp.PreviousValue(65536))
	})

	t.Run("skipping with ranges", func(t *testing.T) {
		bmp := New()
		intervalEnd := 512
		rangeStart := intervalEnd * 2
		rangeEnd := 2048
		for i := 0; i < intervalEnd; i++ {
			bmp.Add(uint32(i * 2))
		}
		bmp.AddRange(uint64(rangeStart), uint64(rangeEnd))

		for i := 0; i < intervalEnd; i++ {
			assert.Equal(t, int64(i*2), bmp.NextValue(uint32(i*2)))
			assert.Equal(t, int64(i*2), bmp.PreviousValue(uint32(i*2)))
			assert.Equal(t, int64(i*2+1), bmp.NextAbsentValue(uint32(i*2)))
			if i != 0 {
				assert.Equal(t, int64(i*2-1), bmp.PreviousAbsentValue(uint32(i*2)))
			}
		}
		for i := rangeStart; i < rangeEnd; i++ {
			assert.Equal(t, int64(i), bmp.NextValue(uint32(i)))
			assert.Equal(t, int64(i), bmp.PreviousValue(uint32(i)))
			assert.Equal(t, int64(rangeEnd), bmp.NextAbsentValue(uint32(i)))
			assert.Equal(t, int64(rangeStart-1), bmp.PreviousAbsentValue(uint32((i))))
		}
	})

	t.Run("randomized", func(t *testing.T) {
		bmp := New()

		intervalEnd := 4096
		entries := make([]uint32, 0, intervalEnd)

		for i := 0; i < intervalEnd; i++ {
			entry := rand.Uint32()
			bmp.Add(entry)
			entries = append(entries, entry)
		}

		for i := 0; i < intervalEnd; i++ {
			entry := entries[i]
			assert.Equal(t, int64(entry), bmp.NextValue(entry))
			assert.Equal(t, int64(entry), bmp.PreviousValue(entry))
			assert.NotEqual(t, int64(entry), bmp.NextAbsentValue(entry))
			assert.NotEqual(t, int64(entry), bmp.PreviousAbsentValue(entry))

		}
	})
}

func BenchmarkFromDense(b *testing.B) {
	testDense(func(name string, rb *Bitmap) {
		dense := make([]uint64, rb.DenseSize())
		rb.WriteDenseTo(dense)
		cp := FromDense(dense, false)

		for _, doCopy := range []bool{false, true} {
			b.Run(fmt.Sprintf("%s,doCopy=%t", name, doCopy), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(dense) * 8))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					cp.FromDense(dense, doCopy)
					cp.Clear()
				}
			})
		}
	})
}

func BenchmarkWriteDenseTo(b *testing.B) {
	testDense(func(name string, rb *Bitmap) {
		b.Run(name, func(b *testing.B) {
			dense := make([]uint64, rb.DenseSize())
			b.ReportAllocs()
			b.SetBytes(int64(len(dense) * 8))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.WriteDenseTo(dense)
			}
		})
	})
}

func BenchmarkEvenIntervalArrayUnions(b *testing.B) {
	inputBitmaps := make([]*Bitmap, 40)
	for i := 0; i < 40; i++ {
		bitmap := NewBitmap()
		for j := 0; j < 100; j++ {
			bitmap.Add(uint32(2 * (j + 10*i)))
		}
		inputBitmaps[i] = bitmap
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bitmap := NewBitmap()
		for _, input := range inputBitmaps {
			bitmap.Or(input)
		}
	}
}

func BenchmarkInPlaceArrayUnions(b *testing.B) {
	rand.Seed(100)
	b.ReportAllocs()
	componentBitmaps := make([]*Bitmap, 100)
	for i := 0; i < 100; i++ {
		bitmap := NewBitmap()
		for j := 0; j < 100; j++ {
			// keep all entries in [0,4096), so they stay arrays.
			bitmap.Add(uint32(rand.Intn(arrayDefaultMaxSize)))
		}
		componentBitmaps[i] = bitmap
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitmap := NewBitmap()
		for j := 0; j < 100; j++ {
			bitmap.Or(componentBitmaps[rand.Intn(100)])
		}
	}
}

func BenchmarkAntagonisticArrayUnionsGrowth(b *testing.B) {
	left := NewBitmap()
	right := NewBitmap()
	for i := 0; i < 4096; i++ {
		left.Add(uint32(2 * i))
		right.Add(uint32(2*i + 1))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		left.Clone().Or(right)
	}
}

func BenchmarkRepeatedGrowthArrayUnion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink := NewBitmap()
		source := NewBitmap()
		for i := 0; i < 2048; i++ {
			source.Add(uint32(2 * i))
			sink.Or(source)
		}
	}
}

func BenchmarkRepeatedSelfArrayUnion(b *testing.B) {
	bitmap := NewBitmap()
	for i := 0; i < 4096; i++ {
		bitmap.Add(uint32(2 * i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		receiver := NewBitmap()
		for j := 0; j < 1000; j++ {
			receiver.Or(bitmap)
		}
	}
}

// BenchmarkArrayIorMergeThreshold tests performance
// when unioning two array containers when the cardinality sum is over 4096
func BenchmarkArrayUnionThreshold(b *testing.B) {
	testOddPoint := map[string]int{
		"mostly-overlap": 4900,
		"little-overlap": 2000,
		"no-overlap":     0,
	}
	for name, oddPoint := range testOddPoint {
		b.Run(name, func(b *testing.B) {
			left := NewBitmap()
			right := NewBitmap()
			for i := 0; i < 5000; i++ {
				if i%2 == 0 {
					left.Add(uint32(i))
				}
				if i%2 == 0 && i < oddPoint {
					right.Add(uint32(i))
				} else if i%2 == 1 && i >= oddPoint {
					right.Add(uint32(i))
				}
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				right.Clone().Or(left)
			}
		})
	}
}

func TestIssue467CaseSmall(t *testing.T) {
	b := New()
	b.AddRange(0, 16385)
	b.AddRange(16385, 20482)
	b.AddRange(20482, 27862)
	b.AddRange(27862, 44247)
	b.AddRange(45576, 61961)
	b.AddRange(61961, 66058)
	b.AddRange(66058, 67247)
	b.AddRange(68819, 73028)
	b.AddRange(73028, 89413)
	b.AddRange(92266, 108651)
	b.AddRange(108651, 113772)
	b.AddRange(113772, 118757)
	b.AddRange(118757, 132098)
	b.RunOptimize()
	require.NoError(t, b.Validate())
}

func TestIssue467CaseLarge(t *testing.T) {
	b := New()
	b.RemoveRange(0, 16385)
	b.RemoveRange(16385, 20482)
	b.RemoveRange(20482, 27862)
	b.AddRange(0, 16385)
	b.AddRange(16385, 20482)
	b.AddRange(20482, 27862)
	b.RemoveRange(27862, 44247)
	b.RemoveRange(44247, 45576)
	b.RemoveRange(45576, 61961)
	b.RemoveRange(61961, 66058)
	b.RemoveRange(66058, 67247)
	b.AddRange(27862, 44247)
	b.AddRange(45576, 61961)
	b.AddRange(44247, 45576)
	b.AddRange(61961, 66058)
	b.AddRange(66058, 67247)
	b.RemoveRange(67247, 68819)
	b.RemoveRange(68819, 73028)
	b.RemoveRange(73028, 89413)
	b.RemoveRange(89413, 92266)
	b.RemoveRange(92266, 108651)
	b.RemoveRange(108651, 113772)
	b.RemoveRange(113772, 118757)
	b.AddRange(68819, 73028)
	b.AddRange(73028, 89413)
	b.AddRange(92266, 108651)
	b.AddRange(89413, 92266)
	b.AddRange(108651, 113772)
	b.AddRange(113772, 118757)
	b.RemoveRange(118757, 132098)
	b.AddRange(118757, 132098)
	b.RemoveRange(132098, 137544)
	b.AddRange(132098, 137544)
	b.RemoveRange(137544, 153929)
	b.RemoveRange(153929, 155151)
	b.RemoveRange(155151, 162078)
	b.RemoveRange(162078, 167119)
	b.RemoveRange(167119, 181012)
	b.RemoveRange(181012, 197397)
	b.RemoveRange(197397, 201244)
	b.RemoveRange(201244, 217629)
	b.RemoveRange(217629, 222750)
	b.RemoveRange(222750, 227708)
	b.RemoveRange(227708, 235777)
	b.RemoveRange(235777, 252162)
	b.RemoveRange(252162, 256259)
	b.AddRange(252162, 256259)
	b.AddRange(227708, 235777)
	b.AddRange(235777, 252162)
	b.RunOptimize()
	require.NoError(t, b.Validate())
}

func TestValidateEmpty(t *testing.T) {
	require.NoError(t, New().Validate())
}

func TestValidate469(t *testing.T) {
	b := New()
	b.RemoveRange(0, 180)
	b.AddRange(0, 180)
	require.NoError(t, b.Validate())
	b.RemoveRange(180, 217)
	b.AddRange(180, 217)
	require.NoError(t, b.Validate())
	b.RemoveRange(217, 2394)
	b.RemoveRange(2394, 2427)
	b.AddRange(2394, 2427)
	require.NoError(t, b.Validate())
	b.RemoveRange(2427, 2428)
	b.AddRange(2427, 2428)
	require.NoError(t, b.Validate())
	b.RemoveRange(2428, 3345)
	require.NoError(t, b.Validate())
	b.RemoveRange(3345, 3346)
	require.NoError(t, b.Validate())
	b.RemoveRange(3346, 3597)
	require.NoError(t, b.Validate())
	b.RemoveRange(3597, 3815)
	require.NoError(t, b.Validate())
	b.RemoveRange(3815, 3816)
	require.NoError(t, b.Validate())
	b.AddRange(3815, 3816)
	require.NoError(t, b.Validate())
	b.RemoveRange(3816, 3856)
	b.RemoveRange(3856, 4067)
	b.RemoveRange(4067, 4069)
	b.RemoveRange(4069, 4071)
	b.RemoveRange(4071, 4095)
	b.RemoveRange(4095, 4096)
	require.NoError(t, b.Validate())
	b.RunOptimize()
	require.False(t, b.IsEmpty())
	require.NoError(t, b.Validate())
}

func TestValidateFromV1(t *testing.T) {
	v1 := New()
	for i := 0; i <= 2; i++ {
		v1.Add(uint32(i))
	}
	v1.RunOptimize()
	b, err := v1.MarshalBinary()
	require.NoError(t, err)
	v2 := New()
	require.NoError(t, v2.UnmarshalBinary(b))
	require.NoError(t, v2.Validate())
}
