package snowflake

import "strconv"

func ParseID(config IDConfig) {

}

func IntToBitString(num int64) string {
	bitCount := strconv.IntSize
	bits := make([]byte, bitCount)

	for i := 0; i < bitCount; i++ {
		if num&(1<<uint(i)) != 0 {
			bits[bitCount-i-1] = '1'
		} else {
			bits[bitCount-i-1] = '0'
		}
	}

	return string(bits)
}
