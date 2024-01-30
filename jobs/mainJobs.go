package jobs

import (
	"encoding/json"

	"github.com/yapkah/go-api/models"
	//"golang.org/x/text/language"
	//"golang.org/x/text/message"
)

func RunMainJobs(arrJob *models.GolangJobs) {

	// in order to add new job. please remember to add ur job in jobList variable on line 20
	type JobListStruct struct {
		JobName string
		Func    func() error
	}

	var jobList []JobListStruct
	jobList = append(jobList,
		JobListStruct{JobName: "TestingJobA", Func: TestingJobA},
	)

	var payLoadJob models.GolangJobsFnStruct
	err := json.Unmarshal([]byte(arrJob.Payload), &payLoadJob)

	if err != nil {
		models.ErrorLog("RunMainJobs_failed_to_decode_payLoadJob", err.Error(), payLoadJob)
		return
	}
	db := models.GetDB() // no need transaction because if failed no need rollback
	for _, v1 := range jobList {
		if v1.JobName == payLoadJob.JobName {
			err := v1.Func()
			if err != nil {
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " id = ? ", CondValue: arrJob.ID},
				)
				attempts := arrJob.Attempts + 1
				updateColumn := map[string]interface{}{"attempts": attempts}
				models.UpdatesFn("golang_jobs", arrCond, updateColumn, false)

				if attempts > 3 {
					arrData := models.FailedGolangJobs{
						Connection: "golang_job",
						Queue:      arrJob.Queue,
						Payload:    arrJob.Payload,
						Exception:  err.Error(),
					}

					_ = models.AddFailedGolangJobs(db, arrData)

					arrCond := make([]models.WhereCondFn, 0)
					arrCond = append(arrCond,
						models.WhereCondFn{Condition: " id = ?", CondValue: arrJob.ID},
					)
					models.DeleteFn("golang_jobs", arrCond, false)
				}
			}
		}
	}
}
