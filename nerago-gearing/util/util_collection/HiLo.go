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

func (hilo HiLoUInt32) Between(check uint32) bool {
	return hilo.Lo <= check && check <= hilo.Hi
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

func (hilo HiLoInt) Between(check int) bool {
	return hilo.Lo <= check && check <= hilo.Hi
}
