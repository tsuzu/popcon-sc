package popconSCTypes

type ContestType int

const (
	JOI ContestType = 0
	// Score: Score(>)
	// Value1: Time(<)
	// Value2: None

	PCK ContestType = 1
	// Score: Score(>)
	// Value1: WA(<)
	// Value2: Time(<)

	AtCoder ContestType = 2
	// Score: Score(>)
	// Value1: Time+Penalty(<)
	// Value2: None

	ICPC ContestType = 3
	// Score: Score(>)
	// Value1: Time(Sum All)+Penalty(<)
	// Value2: None
)

var StringToContestType = map[string]ContestType{
	"JOI":     JOI,
	"PCK":     PCK,
	"AtCoder": AtCoder,
	"ICPC":    ICPC,
}

var ContestTypeToString = map[ContestType]string{
	JOI:     "JOI",
	PCK:     "PCK",
	AtCoder: "AtCoder",
	ICPC:    "ICPC",
}
