package main

import (
	"log"
	"os"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/cs3238-tsuzu/popcon-sc/ranking/restapi"
	"github.com/cs3238-tsuzu/popcon-sc/ranking/restapi/swagger"
)

func main() {
	GeneralSetting.RankingRunningTerm = 10
	GeneralSetting.SavingTerm = 5

	//token := os.Getenv("POPCON_SC_RANKING_TOKEN")
	addr := os.Getenv("POPCON_SC_RANKING_ADDR")
	//	db := os.Getenv("POPCON_SC_RANKING_DB")

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := swagger.NewPopconSCRankingAndSubmissionManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	h, p, err := swag.SplitHostPort(addr)

	if err != nil {
		log.Fatalln("The format of POPCON_SC_RANKING_ADDR is illegal.")
	}

	server.Host = h
	server.Port = p

	server.ConfigureAPI()

	api.PostContestsCreateCidHandler = swagger.PostContestsCreateCidHandlerFunc(func(params swagger.PostContestsCreateCidParams, principle interface{}) middleware.Responder {

		return middleware.NotImplemented("operation .PostContestsGetID has not been implemented yet")
	})

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
