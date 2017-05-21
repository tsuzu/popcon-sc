package ppjctypes

import (
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
)

type JudgeInformation struct {
	Submission database.Submission
	Problem    database.ContestProblem
	Cases      []database.ContestProblemTestCase
	Scores     []database.ContestProblemScoreSet
}
