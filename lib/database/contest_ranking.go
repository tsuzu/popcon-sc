package database

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/jinzhu/gorm"
)

// For ppjc node

func ParseRankingCell(str string) *sctypes.RankingCell {
	rc := &sctypes.RankingCell{}

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
	pidsStr = append(pidsStr, "general VARCHAR(256) DEFAULT '"+sctypes.RankingCell{}.String()+"'")
	pidsStr = append(pidsStr, "iid BIGINT UNIQUE")
	pidsStr = append(pidsStr, "score BIGINT NOT NULL DEFAULT 0")
	pidsStr = append(pidsStr, "value1 BIGINT NOT NULL DEFAULT 0")
	pidsStr = append(pidsStr, "value2 BIGINT NOT NULL  DEFAULT 0")
	pidsStr = append(pidsStr, "rid BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT")
	pidsStr = append(pidsStr, "INDEX(score)")
	pidsStr = append(pidsStr, "INDEX(value1)")
	pidsStr = append(pidsStr, "INDEX(value2)")

	query = query + "(" + strings.Join(pidsStr, ",") + ")"

	_, err := dm.db.CommonDB().Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) RankingDelete(cid int64) error {
	_, err := dm.DB().CommonDB().Exec("DROP TABLE " + dm.RankingTableName(cid))

	return err
}

func (dm *DatabaseManager) RankingProblemAdd(cid, pid int64) error {
	// for i := range pids {
	// 	pidsStr = append(pidsStr, dm.RankingCellName(pids[i])+" VARCHAR(256)")
	// }
	return dm.Begin(func(db *gorm.DB) error {
		_, err := db.CommonDB().Exec("ALTER TABLE " + dm.RankingTableName(cid) + " ADD COLUMN " + dm.RankingCellName(pid) + " VARCHAR(256)")

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
		_, err := db.CommonDB().Exec("ALTER TABLE " + dm.RankingTableName(cid) + " DROP COLUMN " + dm.RankingCellName(pid))

		if err != nil {
			return err
		}

		return nil
	})
}

func (dm *DatabaseManager) RankingUserAdd(cid, iid int64) error {
	_, err := dm.db.CommonDB().Exec("INSERT INTO "+dm.RankingTableName(cid)+"(iid) VALUES (?)", iid)

	if err != nil {
		if strings.Index(err.Error(), "Duplicate") != -1 {
			return nil
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) RankingGetCell(cid, iid, pid int64) (*sctypes.RankingCell, error) {
	rows, err := dm.db.CommonDB().Query("SELECT "+dm.RankingCellName(pid)+" FROM "+dm.RankingTableName(cid)+" WHERE iid=?", iid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, ErrUnknownRankingCell
	}

	var c *sctypes.RankingCell
	if err := rows.Scan(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func (dm *DatabaseManager) RankingGetCellAndGeneral(cid, iid, pid int64) (*sctypes.RankingCell, *sctypes.RankingCell, error) {
	rows, err := dm.db.CommonDB().Query("SELECT "+dm.RankingCellName(pid)+", general FROM "+dm.RankingTableName(cid)+" WHERE iid=?", iid)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, nil, ErrUnknownRankingCell
	}

	var c, g sctypes.RankingCell
	if err := rows.Scan(&c, &g); err != nil {
		return nil, nil, err
	}

	return &c, &g, nil
}

func (dm *DatabaseManager) RankingCellUpdate(cid, iid, pid int64, rc sctypes.RankingCell) error {
	_, err := dm.db.CommonDB().Exec("UPDATE "+dm.RankingTableName(cid)+" SET "+dm.RankingCellName(pid)+"=? WHERE iid=?", rc.String(), iid)

	return err
}

func (dm *DatabaseManager) RankingGeneralUpdate(cid, iid int64, value1, value2 int64, rc sctypes.RankingCell) error {
	_, err := dm.db.CommonDB().Exec("UPDATE "+dm.RankingTableName(cid)+" SET general=?, score=?, value1=?, value2=? WHERE iid=?", rc.String(), rc.Score, value1, value2, iid)

	return err
}

func (dm *DatabaseManager) RankingUpdate(cid, iid, pid int64, rc sctypes.RankingCell) error {
	cont, err := dm.ContestFind(cid)

	if err != nil {
		return err
	}

	return dm.BeginDM(func(dm *DatabaseManager) error {
		entry := dm.Logger().WithField("cid", cid).WithField("iid", iid).WithField("pid", pid)

		row, err := dm.Clone(dm.DB().Set("gorm:query_options", "FOR UPDATE")).RankingGetRow(cid, iid)

		if err != nil {
			if err == ErrUnknownRankingRow {
				return nil
			}

			return err
		}

		if _, ok := row.Problems[pid]; !ok {
			return ErrUnknownProblem
		}
		cell := row.Problems[pid]

		newCell := cell
		if cell.IsValid() {
			if cell.Sid == rc.Sid {
				if cell.Jid > rc.Jid {
					return nil
				}

				if cell.Jid == rc.Jid {
					entry.WithField(
						"sid", cell.Sid).WithField("jid", cell.Jid).Error("Invalid status of Jid and Sid")

					return nil
				}

				if cell.Score <= rc.Score {
					newCell = rc
				} else {
					sm, err := dm.SubmissionMaximumScore(cid, iid, pid)

					if err != nil {
						return err
					}

					newCell.Score = rc.Score
					newCell.Time = sm.SubmitTime.Sub(cont.StartTime)
					newCell.Sid = sm.Sid
					newCell.Jid = sm.Jid

				}
			} else {
				if cell.Score < rc.Score {
					newCell = rc
				} else if cell.Score == rc.Score && cell.Sid > rc.Sid {
					newCell = rc
				}
			}
		} else {
			newCell = rc
		}

		sid := newCell.Sid
		if newCell.Score > 0 {
			sid--
		}
		cnt, err := dm.SubmissionCountForPenalty(cid, iid, pid, sid, sctypes.ContestTypeCEPenalty[cont.Type])

		if err != nil {
			return err
		}

		newCell.Penalty = cnt

		row.Problems[pid] = newCell

		var general sctypes.RankingCell
		general.Score = sctypes.ContestTypeCalculateGeneralScore[cont.Type](row.Problems)
		general.Time = sctypes.ContestTypeCalculateGeneralTime[cont.Type](row.Problems)
		general.Penalty = sctypes.ContestTypeCalculateGeneralPenalty[cont.Type](row.Problems)
		general.Valid = true

		value1 := sctypes.ContestTypeToEvaluationFunction1[cont.Type](general.Score, general.Penalty, cont.Penalty, general.Time)
		value2 := sctypes.ContestTypeToEvaluationFunction2[cont.Type](general.Score, general.Penalty, cont.Penalty, general.Time)

		if err := dm.RankingCellUpdate(cid, iid, pid, newCell); err != nil {
			return err
		}
		if err := dm.RankingGeneralUpdate(cid, iid, value1, value2, general); err != nil {
			return err
		}

		return nil
	})
}

func (dm *DatabaseManager) RankingGetRow(cid, iid int64) (*sctypes.RankingRow, error) {
	rows, err := dm.rankingGetAll(cid, -1, -1, iid)

	if err != nil {
		return nil, err
	}

	if rows == nil || len(rows) == 0 {
		return nil, ErrUnknownRankingRow
	}

	return &rows[0], nil
}

func (dm *DatabaseManager) RankingGetAll(cid, offset, limit int64) ([]sctypes.RankingRow, error) {
	return dm.rankingGetAll(cid, offset, limit, -1)
}

func (dm *DatabaseManager) RankingCount(cid int64) (int64, error) {
	rows, err := dm.DB().CommonDB().Query("SELECT COUNT(iid) FROM " + dm.RankingTableName(cid))

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, ErrUnknownRankingRow
	}

	var cnt int64
	if err := rows.Scan(&cnt); err != nil {
		return 0, err
	}

	return cnt, nil
}

func (dm *DatabaseManager) rankingGetAll(cid, offset, limit, iid int64) ([]sctypes.RankingRow, error) {
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

	query := "SELECT * FROM " + dm.RankingTableName(cid)

	if iid != -1 {
		query = query + " WHERE iid=" + strconv.FormatInt(iid, 10)
	}

	query = query + " ORDER BY score DESC, value1 DESC, value2 DESC"

	if len(str) != 0 {
		query = query + " LIMIT " + str
	}

	rows, err := dm.db.CommonDB().Query(query)

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

	res := make([]sctypes.RankingRow, 0, 50)
	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		for i := range columns {
			if columnNames[i][0] == 'p' {
				columns[i] = &sctypes.RankingCell{}
			} else if columnNames[i] == "general" {
				columns[i] = &sctypes.RankingCell{}
			} else if columnNames[i] == "iid" || columnNames[i] == "rid" {
				var val int64
				columns[i] = &val
			} else if columnNames[i] == "score" || columnNames[i] == "value1" || columnNames[i] == "value2" {
				var val int64
				columns[i] = &val
			} else {
				return nil, errors.New("Unknown column: " + columnNames[i])
			}
		}

		err := rows.Scan(columns...)

		if err != nil {
			return nil, err
		}

		var rr sctypes.RankingRow
		cells := make(map[int64]sctypes.RankingCell)
		for i := range columns {
			if columnNames[i][0] == 'p' {
				id, _ := strconv.ParseInt(columnNames[i][1:], 10, 64)
				cells[id] = *(columns[i].(*sctypes.RankingCell))
			} else if columnNames[i] == "general" {
				rr.General = *(columns[i].(*sctypes.RankingCell))
			} else if columnNames[i] == "iid" {
				rr.Iid = *(columns[i].(*int64))
			}
		}
		rr.Problems = cells
		res = append(res, rr)
	}

	return res, nil
}

func (dm *DatabaseManager) RankingGetAllWithUserData(cid, offset, limit int64) ([]sctypes.RankingRowWithUserData, error) {
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

	query := "SELECT * FROM " + dm.RankingTableName(cid)

	query = query + " INNER JOIN users ON " + dm.RankingTableName(cid) + ".iid=users.iid"
	query = query + " ORDER BY score DESC, value1 DESC, value2 DESC"

	if len(str) != 0 {
		query = query + " LIMIT " + str
	}

	rows, err := dm.db.CommonDB().Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	res := make([]sctypes.RankingRowWithUserData, 0, 50)
	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		for i := range columns {
			if columnNames[i][0] == 'p' {
				columns[i] = &sctypes.RankingCell{}
			} else if columnNames[i] == "general" {
				columns[i] = &sctypes.RankingCell{}
			} else if columnNames[i] == "iid" || columnNames[i] == "rid" {
				var val int64
				columns[i] = &val
			} else if columnNames[i] == "score" || columnNames[i] == "value1" || columnNames[i] == "value2" {
				var val int64
				columns[i] = &val
			} else if columnNames[i] == "uid" || columnNames[i] == "user_name" {
				var str string
				columns[i] = &str
			} else {
				columns[i] = &SqlScanIgnore{}
			}
		}

		err := rows.Scan(columns...)

		if err != nil {
			return nil, err
		}

		var rr sctypes.RankingRowWithUserData
		cells := make(map[int64]sctypes.RankingCell)
		for i := range columns {
			if columnNames[i][0] == 'p' {
				id, _ := strconv.ParseInt(columnNames[i][1:], 10, 64)
				cells[id] = *(columns[i].(*sctypes.RankingCell))
			} else if columnNames[i] == "general" {
				rr.General = *(columns[i].(*sctypes.RankingCell))
			} else if columnNames[i] == "iid" {
				rr.Iid = *(columns[i].(*int64))
			} else if columnNames[i] == "uid" {
				rr.Uid = *(columns[i].(*string))
			} else if columnNames[i] == "user_name" {
				rr.UserName = *(columns[i].(*string))
			}
		}
		rr.Problems = cells
		res = append(res, rr)
	}

	return res, nil
}
