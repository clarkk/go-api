package api

type Limit struct {
	Offset		uint32		`json:"offset"`
	Limit		uint8		`json:"limit"`
	Entries		uint32		`json:"entries"`
}

func (l *Limit) Limit_max(max uint8){
	l.Limit = min(l.Limit, max)
}

func (l *Limit) Count(count uint32) bool {
	l.Entries = count
	//	Offset out of range
	if l.Offset != 0 && l.Offset + 1 > count {
		return false
	}
	return true
}