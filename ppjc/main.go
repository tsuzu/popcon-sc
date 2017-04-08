package main

import (
	"os"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	GeneralSetting.RankingRunningTerm = 10
	GeneralSetting.SavingTerm = 5

	//token := os.Getenv("POPCON_SC_RANKING_TOKEN")
	addr := os.Getenv("POPCON_SC_RANKING_ADDR")
	//	db := os.Getenv("POPCON_SC_RANKING_DB")

	
}
