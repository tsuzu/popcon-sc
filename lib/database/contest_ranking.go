package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/jinzhu/gorm"
)

// For ppjc node

type RankingCell struct {
	Valid          bool
	Sid, Jid       int64
	Time           time.Duration
	Score, Penalty int64
}

var InvalidRankingCell = RankingCell{Valid: false}

func (rc RankingCell) IsValid() bool {
	return rc.Valid
}

func (rc RankingCell) String() string {
	b, _ := json.Marshal(rc)

	return string(b)
}

func (rc *RankingCell) Parse(str string) {
	if len(str) == 0 {
		rc.Valid = false

		return
	}

	if err := json.Unmarshal([]byte(str), rc); err != nil {
		rc.Valid = false
	} else {
		rc.Valid = true
	}
}

func (rc *RankingCell) Scan(v interface{}) error {
	var str sql.NullString

	if err := str.Scan(v); err != nil {
		return err
	}

	rc.Parse(str.String)

	return nil
}

func ParseRankingCell(str string) *RankingCell {
	rc := &RankingCell{}

	rc.Parse(str)

	return rc
}

func (dm *DatabaseManager) CreateRankingTable() error {
	// Do nothing

	return nil
}

func (dm *DatabaseManager) RankingTableName(cid int64) string {
	return "contest_ranking_" + strconv.FormatInt(cid, 10)
}

func (dm *DatabaseManager) RankingCellName(pid int64) string {
	return "p" + strconv.FormatInt(pid, 10)
}

const RankingNumberOfNotProblem = 6

func (dm *DatabaseManager) RankingAutoMigrate(cid int64) error {
	query := "CREATE TABLE IF NOT EXISTS " + dm.RankingTableName(cid) + " "

	pidsStr := make([]string, 0, 10)
	pidsStr = append(pidsStr, "general VARCHAR(256) DEFAULT '"+RankingCell{}.String()+"'")
	pidsStr = append(pidsStr, "iid BIGINT UNIQUE")
	pidsStr = append(pidsStr, "score BIGINT DEFAULT 0")
	pidsStr = append(pidsStr, "value1 BIGINT")
	pidsStr = append(pidsStr, "value2 BIGINT")
	pidsStr = append(pidsStr, "rid BIGINT NOT NULL PRIMARY KEY")
	pidsStr = append(pidsStr, "INDEX(score)")
	pidsStr = append(pidsStr, "INDEX(value1)")
	pidsStr = append(pidsStr, "INDEX(value2)")

	query = query + "(" + strings.Join(pidsStr, ",") + ")"

	_, err := dm.db.DB().Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) RankingProblemAdd(cid, pid int64) error {
	// for i := range pids {
	// 	pidsStr = append(pidsStr, dm.RankingCellName(pids[i])+" VARCHAR(256)")
	// }
	return dm.Begin(func(db *gorm.DB) error {
		_, err := db.DB().Exec("ALTER TABLE " + dm.RankingTableName(cid) + " ADD COLUMN " + dm.RankingCellName(pid) + " VARCHAR(256)")

		if err != nil {
			return err
		}

		return nil
	})
}

func (dm *DatabaseManager) RankingProblemDelete(cid, pid int64) error {
	// for i := range pids {
	// 	pidsStr = append(pidsStr, dm.RankingCellName(pids[i])+" VARCHAR(256)")
	// }
	return dm.Begin(func(db *gorm.DB) error {
		_, err := db.DB().Exec("ALTER TABLE " + dm.RankingTableName(cid) + " DROP COLUMN " + dm.RankingCellName(pid))

		if err != nil {
			return err
		}

		return nil
	})
}

func (dm *DatabaseManager) RankingUserAdd(cid, iid int64) error {
	_, err := dm.db.DB().Exec("INSERT INTO "+dm.RankingTableName(cid)+"(iid) VALUES (?)", iid)

	if err != nil {
		if strings.Index(err.Error(), "Duplicate") != -1 {
			return nil
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) RankingGetCell(cid, iid, pid int64) (*RankingCell, error) {
	rows, err := dm.db.DB().Query("SELECT "+dm.RankingCellName(pid)+" FROM "+dm.RankingTableName(cid)+" WHERE iid=?", iid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, ErrUnknownRankingCell
	}

	var c *RankingCell
	if err := rows.Scan(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func (dm *DatabaseManager) RankingGetCellAndGeneral(cid, iid, pid int64) (*RankingCell, *RankingCell, error) {
	rows, err := dm.db.DB().Query("SELECT "+dm.RankingCellName(pid)+", general FROM "+dm.RankingTableName(cid)+" WHERE iid=?", iid)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, nil, ErrUnknownRankingCell
	}

	var c, g RankingCell
	if err := rows.Scan(&c, &g); err != nil {
		return nil, nil, err
	}

	return &c, &g, nil
}

func (dm *DatabaseManager) RankingCellUpdate(cid, iid, pid int64, rc RankingCell) error {
	_, err := dm.db.DB().Exec("UPDATE "+dm.RankingTableName(cid)+" SET "+dm.RankingCellName(pid)+"=? WHERE iid=?", rc.String(), iid)

	return err
}

func (dm *DatabaseManager) RankingGeneralUpdate(cid, iid int64, value1, value2 int64, rc RankingCell) error {
	_, err := dm.db.DB().Exec("UPDATE "+dm.RankingTableName(cid)+" SET general=?, score=?, value1=?, value2=? WHERE iid=?", rc.String(), rc.Score, value1, value2, iid)

	return err
}

func (dm *DatabaseManager) RankingUpdate(cid, iid, pid int64, rc RankingCell) error {
	cont, err := dm.ContestFind(cid)

	if err != nil {
		return err
	}

	return dm.BeginDM(func(dm *DatabaseManager) error {
		entry := dm.Logger().WithField("cid", cid).WithField("iid", iid).WithField("pid", pid)

		cell, general, err := dm.RankingGetCellAndGeneral(cid, iid, pid)

		if err != nil {
			return err
		}

		var timeDiff time.Duration = 0
		var scoreDiff, penaltyDiff int64 = 0, 0
		if cell.IsValid() {
			if cell.Sid == rc.Sid {
				if cell.Jid > rc.Jid {
					return nil
				}

				if cell.Jid == rc.Jid {
					entry.WithField("sid", cell.Sid).WithField("jid", cell.Jid).Error("Impossible status of Jid and Sid")

					return nil
				}

				if cell.Score <= rc.Score {
					scoreDiff = rc.Score - cell.Score
					timeDiff = rc.Time - cell.Time
				} else {
					sm, err := dm.SubmissionMaximumScore(cid, iid, pid)

					if err != nil {
						return err
					}

					scoreDiff = rc.Score - cell.Score
					timeDiff = sm.SubmitTime.Sub(cont.StartTime) - cell.Time

					cnt, err := dm.SubmissionCountForPenalty(cid, iid, pid, sm.Sid, sctypes.ContestTypeCEPenalty[cont.Type])

					if err != nil {
						return err
					}

					penaltyDiff = cnt - cell.Penalty
				}
			} else {
				if cell.Score > rc.Score {
					return nil
				}
				if cell.Score == rc.Score && cell.Sid < rc.Sid {
					return nil
				}

				penalty, err := dm.SubmissionCountForPenalty(cid, iid, pid, rc.Sid, sctypes.ContestTypeCEPenalty[cont.Type])

				if err != nil {
					return err
				}

				penaltyDiff = penalty - cell.Penalty
				timeDiff = rc.Time - cell.Time
				scoreDiff = rc.Score - cell.Score
			}
		} else {
			timeDiff = rc.Time
			scoreDiff = rc.Score
		}

		general.Score += scoreDiff
		general.Time += timeDiff
		general.Penalty += penaltyDiff

		value1 := sctypes.ContestTypeToEvaluationFunction1[cont.Type](general.Score, general.Penalty, general.Time)
		value2 := sctypes.ContestTypeToEvaluationFunction2[cont.Type](general.Score, general.Penalty, general.Time)

		if err := dm.RankingCellUpdate(cid, iid, pid, rc); err != nil {
			return err
		}
		if err := dm.RankingGeneralUpdate(cid, iid, value1, value2, *general); err != nil {
			return err
		}

		return nil
	})
}

type RankingRow struct {
	Iid      int64
	Problems map[int64]RankingCell
	General  RankingCell
}

func (dm *DatabaseManager) RankingGetAll(cid, offset, limit int64) ([]RankingRow, error) {
	var str string
	if offset != -1 {
		str = strconv.FormatInt(offset, 10)
	}
	if limit != -1 {
		if len(str) != 0 {
			str = str + ","
		}
		str = str + strconv.FormatInt(limit, 10)
	}

	query := "SELECT * FROM " + dm.RankingTableName(cid) + " ORDER BY score DESC, value1 DESC, value2 DESC "

	if len(str) != 0 {
		query = query + "LIMIT " + str
	}
	rows, err := dm.db.DB().Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	if len(columnNames)-RankingNumberOfNotProblem < 0 {
		return nil, errors.New("Too few columns")
	}

	res := make([]RankingRow, 0, 50)
	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		for i := 0; i < len(columnNames)-RankingNumberOfNotProblem+1; i++ {
			columns[i] = &RankingCell{}
		}
		for i := len(columnNames) - RankingNumberOfNotProblem + 1; i < len(columnNames); i++ {
			var val int64
			columns[i] = &val
		}

		err := rows.Scan(columns...)

		if err != nil {
			return nil, err
		}

		var rr RankingRow
		cells := make(map[int64]RankingCell)
		for i := range columns {
			if columnNames[i][0] == 'p' {
				id, _ := strconv.ParseInt(columnNames[i][1:], 10, 64)
				cells[id] = *(columns[i].(*RankingCell))
			} else if columnNames[i] == "general" {
				rr.General = *(columns[i].(*RankingCell))
			} else if columnNames[i] == "iid" {
				rr.Iid = *(columns[i].(*int64))
			}
		}
		rr.Problems = cells
		res = append(res, rr)
	}

	return res, nil
}
