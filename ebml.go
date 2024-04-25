package main

func variableWidthUintSize(v uint64) int {
	switch {
	case v < 0x80-1:
		return 1
	case v < 0x4000-1:
		return 2
	case v < 0x200000-1:
		return 3
	case v < 0x10000000-1:
		return 4
	case v < 0x800000000-1:
		return 5
	case v < 0x40000000000-1:
		return 6
	case v < 0x2000000000000-1:
		return 7
	default:
		return 8
	}
}
