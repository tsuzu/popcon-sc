package ppjctypes

import (
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
)

type JudgeInformation struct {
	Submission database.Submission
	Problem    database.ContestProblem
	Cases      []database.ContestProblemTestCase
	Scores     []database.ContestProblemScoreSet
}

type JudgeTestcaseResult struct {
	Cid      int64
	Sid      int64
	Jid      int64
	Status   string
	Testcase database.SubmissionTestCase
}

type JudgeSubmissionResult struct {
	Cid    int64
	Sid    int64
	Jid    int64
	Status sctypes.SubmissionStatusType
	Time   int64
	Mem    int64
	Score  int64
}
