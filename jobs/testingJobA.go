package jobs

import (
	"net/http"

	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/e"
)

//"golang.org/x/text/language"
//"golang.org/x/text/message"

func TestingJobA() error {

	models.ErrorLog("TestingJobA", "test", nil) //store error log
	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "test", Data: ""}
}
