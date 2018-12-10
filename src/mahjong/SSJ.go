package mahjong

// CalTai cals tai
func CalTai(hand uint64, door uint64) int {
	idx := (((gData[hand & 134217727] | 4) &
		(gData[hand >> 27] | 4) &
		(gData[door & 134217727] | 64) &
		(gData[door >> 27] | 64) &
		(484 | ((gData[door & 134217727] & gData[door >> 27] & 16) >> 1))) |
		(((gData[hand & 134217727] & (gData[door & 134217727] | 3)) | (gData[hand >> 27] & (gData[door >> 27] | 3))) & 19) |
		((gData[(hand & 134217727) + (door & 134217727)] & 3584) + (gData[(hand >> 27) + (door >> 27)] & 3584)))
	return gData[size + idx];
}

// InitHuTable intis the hu table
func InitHuTable() bool {
	var i uint
	gGroup = append(gGroup, 0)
	for i = 0; i < 9; i++ {
		gGroup = append(gGroup, 3 << (i * 3))
	}
	for i = 0; i < 7; i++ {
		gGroup = append(gGroup, 73 << (i * 3))
	}
	for i = 0; i < 9; i++ {
		gEye = append(gEye, 2 << (i * 3))
	}
	b01(4, 0, size)
	b2(4, 0, 1)
	b3(7, 0, 1)
	b4()
	b5(4, 0, 1)
	b6()
	b7()
	b8(4, 0, size)
	b9UP()
	t()
	println("Initialization Completed!")
	return true
}

const size = 76695844
const max  = 76699939

var gData  [max]int
var gGroup []int
var gEye   []int

func have(m int, s int) bool {
	for i := uint(0); i < 9; i++ {
		if ((m >> (i * 3)) & 7) < ((s >> (i * 3)) & 7) {
			return false
		}
	}
	return true
}

func b01(n int, d int, p int) {
	var i int
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, gGroup[i]) {
				b01(n - 1, d + gGroup[i], p - gGroup[i])
			}
		}
	} else {
		gData[d] |= 1
		for i = 0; i < 9; i++ {
			if have(p, gEye[i]) {
				gData[d + gEye[i]] |= 2
			}
		}
	}
}

func b2(n int, d int, c int) {
	gData[d] |= 4;
	gData[d] |= 32;
	if (d & 16777208) == 0 {
		gData[d] |= 256;
	}
	if n != 0 {
		for i := c; i <= 9; i++ {
			b2(n - 1, d + gGroup[i], i + 1);
			b2(n - 1, d + gGroup[i] / 3 * 4, i + 1);
		}
	}
}

func b3(n int, d int, c int) {
	gData[d] |= 8;
	if n != 0 {
		for i := c; i <= 9; i++ {
			b3(n - 1, d + gGroup[i] / 3 * 2, i + 1);
			b3(n - 2, d + gGroup[i] / 3 * 4, i + 1);
		}
	}
}

func b4() {
	gData[0] |= 16
}

func b5(n int, d int, c int) {
	var i int;
	gData[d] |= 32;
	for i = 0; i < 9; i++ {
		if have(size - d, gEye[i]) {
			gData[d + gEye[i]] |= 32;
		}
	}
	if n != 0 {
		for i = c; i <= 9; i++ {
			b5(n - 1, d + gGroup[i], i + 1);
		}
	}
}

func b6() {
	gData[0] |= 64;
	for i := 0; i < 9; i++ {
		gData[gEye[i]] |= 64;
	}
}

func b7() {
	for i := 0; i < size; i++ {
		if (i & 119508935) == 0 {
			gData[i] |= 128;
		}
	}
}

func b8(n int, d int, p int) {
	var i int;
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, gGroup[i]) && (i == 0 || i == 1 || i == 9 || i == 10 || i == 16) {
				b8(n - 1, d + gGroup[i], p - gGroup[i]);
			}
		}
	} else {
		gData[d] |= 256;
		for i = 0; i < 9; i++ {
			if have(p, gEye[i]) && (i == 0 || i == 8) {
				gData[d + gEye[i]] |= 256;
			}
		}
	}
}

func b9UP() {
	for i := 0; i < size; i++ {
		k := 0;
		for j := uint(0); j < 9; j++ {
			if (i & (4 << (j * 3))) != 0 {
				k++;
			}
		}
		if k > 7 {
			k = 7;
		}
		gData[i] |= (k << 9);
	}
}

func t() {
	for i := uint(0); i < 4095; i++ {
		k := uint(0);
		if (i & 7) == 7 {
			k = 1;
			if (i & 32) == 32 {
				k = 2;
			}
			if (i & 16) == 16 {
				k = 3;
			}
			if (i & 64) == 64 {
				k = 3;
			}
			if (i & 256) == 256 {
				k = 3;
			}
			if (i & 48) == 48 {
				k = 4;
			}
			if (i & 160) == 160 {
				k = 4;
			}
			if (i & 272) == 272 {
				k = 5;
			}
			if (i & 80) == 80 {
				k = 5;
			}
			if (i & 192) == 192 {
				k = 5;
			}
			k += (i >> 9);
		} else if (i & 8) != 0 {
			k = 3;
			if (i & 16) == 16 {
				k = 5;
			}
			if (i >> 9) != 0 {
				k = 4;
			}
			if (i & 16) == 16 && (i >> 9) != 0 {
				k = 5;
			}
			k += (i >> 9);
		}
		gData[size + i] = int(k);
	}
}