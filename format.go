package sectiontrace

type Phase string

const (
	Begin = Phase("b")
	End   = Phase("e")
)

type Record struct {
	Category        string                 `json:"cat"`
	Name            string                 `json:"name"`
	Phase           Phase                  `json:"ph"`
	Scope           string                 `json:"scope,omitempty"`
	TimestampMicros int64                  `json:"ts"`
	ID              int32                  `json:"id"`
	ProcessID       int32                  `json:"pid"`
	Args            map[string]interface{} `json:"args,omitempty"`
}

type Summary struct {
	TraceEvents     []*Record              `json:"traceEvents"`
	DisplayTimeUnit string                 `json:"displayTimeUnit"`
	OtherData       map[string]interface{} `json:"otherData,omitempty"`
}

func Export(recs []*Record) *Summary {
	rv := &Summary{
		TraceEvents:     recs,
		DisplayTimeUnit: DefaultDisplayTimeUnit,
	}
	if len(DefaultOtherData) > 0 {
		rv.OtherData = DefaultOtherData
	}
	return rv
}
