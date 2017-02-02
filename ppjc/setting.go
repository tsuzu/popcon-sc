package main

var GeneralSetting struct {
	RankingSavedFolderPath string
	RankingRunningTerm     int64 /*Unit: min*/
	SavingTerm             int64 /*Unit: min*/
	UpdateRankingTerm      int64 /*Unit: ms*/
}
