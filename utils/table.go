package utils

import "sync"

type Table struct {
	lock sync.Mutex

	records Records
}

type Records struct {
	Inflow  []Record `json:"inflow,omitempty"`
	Outflow []Record `json:"outflow,omitempty"`
}

type Record struct {
	PrivKey string `json:"private_key,omitempty"`

	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`

	Amount string `json:"amount"`
	Type   string `json:"type"`
	TxId   string `json:"txID,omitempty"`

	Timestamp int64 `json:"timestamp"`

	Owner string `json:"owner"`
}

func (t *Table) Inflow(record *Record) {
	defer t.lock.Unlock()
	t.lock.Lock()

	t.records.Inflow = append(t.records.Inflow, *record)
}

func (t *Table) Outflow(record *Record) {
	defer t.lock.Unlock()
	t.lock.Lock()

	t.records.Outflow = append(t.records.Outflow, *record)
}

func (t *Table) Swap(records Records) Records {
	defer t.lock.Unlock()
	t.lock.Lock()

	tmp := t.records
	t.records = records

	return tmp
}

func (r *Records) IsEmpty() bool {
	return len(r.Inflow) == 0 && len(r.Outflow) == 0
}
