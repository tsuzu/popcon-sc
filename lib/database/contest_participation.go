package database

// TODO: 処理をppjqに移動するため不要となる

type ContestParticipation struct {
	Cpid    int64 `gorm:"primary_key"`
	Iid     int64
	Cid     int64
	Score   int64
	Time    int64
	Details string
}

func (dm *DatabaseManager) CreateContestParticipationTable() error {
	err := dm.db.AutoMigrate(&ContestParticipation{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestParticipationAdd(iid, cid int64) error {
	return nil
}

func (dm *DatabaseManager) ContestParticipationCheck(iid, cid int64) (bool, error) {
	return true, nil
}

func (dm *DatabaseManager) ContestParticipationRemove(cid int64) error {
	return nil
}

type RankingHighScoreData struct {
	Sid   int64
	Score int64
	Time  int64
}

func (dm *DatabaseManager) ContestRankingCount(cid int64) (int64, error) {
	return 0, nil
}

type RankingRow struct {
	Uid      string
	UserName string
	Score    int64
	Time     int64
	Probs    map[int64]RankingHighScoreData
}

func (dm *DatabaseManager) ContestRankingList(cid int64, offset int64, limit int64) ([]RankingRow, error) {
	return []RankingRow{}, nil
}

func (dm *DatabaseManager) ContestRankingUpdate(sm Submission) (rete error) {
	return nil
}

func (dm *DatabaseManager) ContestRankingCheckProblem(cid int64) (rete error) {

	return nil
}
