package popconSCTypes

type ContestType int

const (
	ContestTypeJOI ContestType = 0
	// Score: Score(>)
	// Value1: Time(<)
	// Value2: None

	ContestTypeICPC ContestType = 1
	// Score: Score(>)
	// Value1: WA(<)
	// Value2: Time(<)

	ContestTypeAtCoder ContestType = 2
	// Score: Score(>)
	// Value1: Time+Penalty(<)
	// Value2: None

	ContestTypePCK ContestType = 3
	// Score: Score(>)
	// Value1: Time(Sum All)+Penalty(<)
	// Value2: None
)

var ContestTypeToString = map[ContestType]string{
	ContestTypeJOI:     "JOI",
	ContestTypeICPC:    "ICPC",
	ContestTypeAtCoder: "AtCoder",
	ContestTypePCK:     "PCK",
}

var ContestTypeFromString = map[string]ContestType{
	"JOI":     ContestTypeJOI,
	"ICPC":    ContestTypeICPC,
	"AtCoder": ContestTypeAtCoder,
	"PCK":     ContestTypePCK,
}
