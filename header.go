package vulpo

import "C"
import "time"

type Header struct {
	recordcount uint
	lastUpdated time.Time
	hasIndex    bool
	hasFpt      bool
	codepage    Codepage
}

func (h *Header) RecordCount() uint {
	return h.recordcount
}

func (h *Header) LastUpdated() time.Time {
	return h.lastUpdated
}

func (h *Header) HasIndex() bool {
	return h.hasIndex
}

func (h *Header) HasFpt() bool {
	return h.hasFpt
}

func (h *Header) Codepage() Codepage {
	return h.codepage
}
