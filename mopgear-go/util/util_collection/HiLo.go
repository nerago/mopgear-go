package util_collection

type HiLoUInt32 struct {
	Lo uint32
	Hi uint32
}

func (hilo HiLoUInt32) Mid() uint32 {
	return hilo.Lo + (hilo.Hi-hilo.Lo)/2
}

func (hilo HiLoUInt32) Overlap(other HiLoUInt32) bool {
	if hilo.Lo > other.Hi {
		return false
	} else if other.Lo > hilo.Hi {
		return false
	} else {
		return true
	}
}

func (hilo HiLoUInt32) Gap(other HiLoUInt32) uint32 {
	if hilo.Lo > other.Hi {
		return hilo.Lo - other.Hi
	} else if other.Lo > hilo.Hi {
		return other.Lo - hilo.Hi
	} else {
		return 0
	}
}

func (hilo HiLoUInt32) Between(check uint32) bool {
	return hilo.Lo <= check && check <= hilo.Hi
}

func (hilo HiLoUInt32) Size() uint32 {
	return hilo.Hi - hilo.Lo + 1
}

type HiLoInt struct {
	Lo int
	Hi int
}

func (hilo HiLoInt) Mid() int {
	return hilo.Lo + (hilo.Hi-hilo.Lo)/2
}

func (hilo HiLoInt) Overlap(other HiLoInt) bool {
	if hilo.Lo > other.Hi {
		return false
	} else if other.Lo > hilo.Hi {
		return false
	} else {
		return true
	}
}

func (hilo HiLoInt) Gap(other HiLoInt) int {
	if hilo.Lo > other.Hi {
		return hilo.Lo - other.Hi
	} else if other.Lo > hilo.Hi {
		return other.Lo - hilo.Hi
	} else {
		return 0
	}
}

func (hilo HiLoInt) Between(check int) bool {
	return hilo.Lo <= check && check <= hilo.Hi
}

type HiLoFloat struct {
	Lo float64
	Hi float64
}

func (hilo HiLoFloat) Mid() float64 {
	return (hilo.Lo + hilo.Hi) / 2
}

func (hilo HiLoFloat) Overlap(other HiLoFloat) bool {
	if hilo.Lo > other.Hi {
		return false
	} else if other.Lo > hilo.Hi {
		return false
	} else {
		return true
	}
}

func (hilo HiLoFloat) Gap(other HiLoFloat) float64 {
	if hilo.Lo > other.Hi {
		return hilo.Lo - other.Hi
	} else if other.Lo > hilo.Hi {
		return other.Lo - hilo.Hi
	} else {
		return 0
	}
}

// GapSigned:
//
//	 Similar to cmp:
//		<0 if receiver is less than other,
//		 0 if receiver overlaps other,
//		>0 if receiver is greater than other.
func (hilo HiLoFloat) GapSigned(other HiLoFloat) float64 {
	if hilo.Hi < other.Lo {
		// <0 if receiver is less than other
		return hilo.Hi - other.Lo
	} else if hilo.Lo > other.Hi {
		// >0 if receiver is greater than other
		return hilo.Lo - other.Hi
	} else {
		// 0 if receiver overlaps other
		return 0
	}
}

func (hilo HiLoFloat) Between(check float64) bool {
	return hilo.Lo <= check && check <= hilo.Hi
}
