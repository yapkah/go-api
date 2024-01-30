package sales_service

import (
	"net/http"
	"os"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
)

type BZZNodeContractNewUniqueSerialNumberStruct struct {
	EntMemberID  int
	SerialNumber string
}

func GetBZZNodeContractNewUniqueSerialNumber(arrData BZZNodeContractNewUniqueSerialNumberStruct) string {
	codeCharSet := "1234567890"
	serialNumber := arrData.SerialNumber
	entMemberID := arrData.EntMemberID

	for {
		if serialNumber == "0" {
			serialNumber = base.GenerateRandomString(10, codeCharSet)
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "serial_number = ?", CondValue: serialNumber},
		)
		arrExistingCode, _ := models.GetSlsMasterMiningFn(arrCond, "", false)
		if len(arrExistingCode) > 0 {
			if arrExistingCode[0].MemberID == entMemberID {
				return serialNumber
			}
		} else if len(arrExistingCode) < 1 {
			return serialNumber
		}
		serialNumber = base.GenerateRandomString(10, codeCharSet)
	}
}

type BZZContractPDFStruct struct {
	NickName     string
	SerialNumber string
	LangCode     string
	DocNo        string
	TotalAmount  string
	TotalNodes   string
}

type ProcessGenerateBZZContractPDFStruct struct {
	NickName     string
	DocNo        string
	SerialNumber string
	SlsMasterID  int
	MemberID     int
	LangCode     string
	TotalAmount  string
	TotalNode    string
}

func ProcessGenerateBZZContractPDF(arrData ProcessGenerateBZZContractPDFStruct) error {
	docPath := "./docs/member/sales/node/" + arrData.DocNo + "_node_contract_" + strings.ToLower(arrData.LangCode) + ".pdf"
	if _, err := os.Stat(docPath); os.IsNotExist(err) {

		arrBZZNodeContractNewUniqueSerialNumber := BZZNodeContractNewUniqueSerialNumberStruct{
			EntMemberID:  arrData.MemberID,
			SerialNumber: arrData.SerialNumber,
		}
		serialNumber := GetBZZNodeContractNewUniqueSerialNumber(arrBZZNodeContractNewUniqueSerialNumber)
		arrBZZContractPDF := BZZContractPDFStruct{
			SerialNumber: serialNumber,
			NickName:     arrData.NickName,
			DocNo:        arrData.DocNo,
			LangCode:     arrData.LangCode,
			TotalNodes:   arrData.TotalNode,
			TotalAmount:  arrData.TotalAmount,
		}

		err = GenerateBZZContractPDF(arrBZZContractPDF)
		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " sls_master_mining.sls_master_id = ? ", CondValue: arrData.SlsMasterID},
		)
		updateColumn := map[string]interface{}{"serial_number": serialNumber}
		_ = models.UpdatesFn("sls_master_mining", arrCond, updateColumn, false)
	}
	return nil
}

func GenerateBZZContractPDF(arrData BZZContractPDFStruct) error {
	var err error

	if strings.ToLower(arrData.LangCode) == "zh" {
		err = GenerateBZZContractZhPDF(arrData)
	} else {
		err = GenerateBZZContractENPDF(arrData)
	}
	return err
}

func GenerateBZZContractENPDF(arrData BZZContractPDFStruct) error {

	pdfPath := "./docs/templates/sales/node/sec_swarm_server_leasing_contract_node_en.pdf"
	// pdf, _ := goPDF.GetDefaultPDFConfiguration()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A
	pdf.AddPage()

	pdf.AddTTFFont("Times", "./pkg/gopdf/ttf/times.ttf")            // std alphabet
	pdf.AddTTFFont("namum", "./pkg/gopdf/ttf/NanumBarunGothic.ttf") // kr
	pdf.AddTTFFont("loma", "./pkg/gopdf/ttf/Loma.ttf")              // thai

	// Import page 1
	tpf := pdf.ImportPage(pdfPath, 1, "/MediaBox")
	// Draw pdf onto page 1
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 16)
	pdf.SetX(190)
	pdf.SetY(713)
	pdf.Text(arrData.SerialNumber)

	// Import page 2
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 2, "/MediaBox")
	// Draw pdf onto page 2
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 14)
	pdf.SetX(188)
	pdf.SetY(182)
	pdf.Text(arrData.SerialNumber)

	nickNameFont := "Times"
	if strings.ToLower(arrData.LangCode) == "kr" {
		nickNameFont = "namum"
	} else if strings.ToLower(arrData.LangCode) == "th" {
		nickNameFont = "loma"
	}
	pdf.SetFont(nickNameFont, "", 14)
	pdf.SetX(166)
	pdf.SetY(279)
	pdf.Text(arrData.NickName)

	// Import page 3
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 3, "/MediaBox")

	pdf.SetFont("Times", "", 12)
	pdf.SetX(190)
	pdf.SetY(650)
	pdf.Text(arrData.TotalNodes)

	pdf.SetFont("Times", "", 12)
	pdf.SetX(202)
	pdf.SetY(668)
	pdf.Text(arrData.TotalAmount)

	// Draw pdf onto page 3
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 4
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 4, "/MediaBox")
	// Draw pdf onto page 4
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 5
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 5, "/MediaBox")
	// Draw pdf onto page 5
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 6
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 6, "/MediaBox")
	// Draw pdf onto page 6
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	err := pdf.WritePdf("./docs/member/sales/node/" + arrData.DocNo + "_node_contract_en.pdf")

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

func GenerateBZZContractZhPDF(arrData BZZContractPDFStruct) error {
	pdfPath := "./docs/templates/sales/node/sec_swarm_server_leasing_contract_node_zh.pdf"

	// pdf, _ := goPDF.GetDefaultPDFConfiguration()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A
	pdf.AddPage()

	pdf.AddTTFFont("Times", "./pkg/gopdf/ttf/times.ttf")   // std alphabet
	pdf.AddTTFFont("simhei", "./pkg/gopdf/ttf/simhei.ttf") // zh

	// Import page 1
	tpf := pdf.ImportPage(pdfPath, 1, "/MediaBox")
	// Draw pdf onto page 1
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 16)
	pdf.SetX(192)
	pdf.SetY(738)
	pdf.Text(arrData.SerialNumber)

	// Import page 2
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 2, "/MediaBox")
	// Draw pdf onto page 2
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 14)
	pdf.SetX(177)
	pdf.SetY(182)
	pdf.Text(arrData.SerialNumber)

	pdf.SetFont("simhei", "", 14)
	pdf.SetX(166)
	pdf.SetY(279)
	pdf.Text(arrData.NickName)

	// Import page 3
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 3, "/MediaBox")

	pdf.SetFont("Times", "", 12)
	pdf.SetX(160)
	pdf.SetY(630)
	pdf.Text(arrData.TotalNodes)

	pdf.SetFont("Times", "", 12)
	pdf.SetX(196)
	pdf.SetY(656)
	pdf.Text(arrData.TotalAmount)

	// Draw pdf onto page 3
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 4
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 4, "/MediaBox")
	// Draw pdf onto page 4
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 5
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 5, "/MediaBox")
	// Draw pdf onto page 5
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 6
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 6, "/MediaBox")
	// Draw pdf onto page 6
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	err := pdf.WritePdf("./docs/member/sales/node/" + arrData.DocNo + "_node_contract_zh.pdf")

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

type BZZBroadbandContractNewUniqueSerialNumberStruct struct {
	EntMemberID  int
	SerialNumber string
}

func GetBZZBroadbandContractNewUniqueSerialNumber(arrData BZZBroadbandContractNewUniqueSerialNumberStruct) string {
	codeCharSet := "1234567890"
	serialNumber := arrData.SerialNumber
	entMemberID := arrData.EntMemberID

	for {
		if serialNumber == "0" {
			serialNumber = base.GenerateRandomString(10, codeCharSet)
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "serial_number = ?", CondValue: serialNumber},
		)
		arrExistingCode, _ := models.GetSlsMasterMiningNodeTopupFn(arrCond, true)
		if len(arrExistingCode) > 0 {
			if arrExistingCode[0].MemberID == entMemberID {
				return serialNumber
			}
		} else if len(arrExistingCode) < 1 {
			return serialNumber
		}
		serialNumber = base.GenerateRandomString(10, codeCharSet)
	}
}

type BZZBroadbandContractPDFStruct struct {
	NickName     string
	SerialNumber string
	LangCode     string
	DocNo        string
	Months       string
	TotalNodes   string
}

type ProcessGenerateBZZBroadbandContractPDFStruct struct {
	NickName     string
	DocNo        string
	ID           int
	MemberID     int
	LangCode     string
	Months       string
	TotalNode    string
	SerialNumber string
}

func ProcessGenerateBZZBroadbandContractPDF(arrData ProcessGenerateBZZBroadbandContractPDFStruct) error {
	docPath := "./docs/member/sales/broadband/" + arrData.DocNo + "_broadband_contract_" + strings.ToLower(arrData.LangCode) + ".pdf"
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		arrBZZBroadbandContractNewUniqueSerialNumber := BZZBroadbandContractNewUniqueSerialNumberStruct{
			EntMemberID:  arrData.MemberID,
			SerialNumber: arrData.SerialNumber,
		}
		serialNumber := GetBZZBroadbandContractNewUniqueSerialNumber(arrBZZBroadbandContractNewUniqueSerialNumber)

		arrPDF := BZZBroadbandContractPDFStruct{
			SerialNumber: serialNumber,
			NickName:     arrData.NickName,
			DocNo:        arrData.DocNo,
			LangCode:     arrData.LangCode,
			TotalNodes:   arrData.TotalNode,
			Months:       arrData.Months,
		}
		err = GenerateBZZBroadbandContractPDF(arrPDF)
		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " sls_master_mining_node_topup.id = ? ", CondValue: arrData.ID},
		)
		updateColumn := map[string]interface{}{"serial_number": serialNumber}
		_ = models.UpdatesFn("sls_master_mining_node_topup", arrCond, updateColumn, false)
	}
	return nil
}

func GenerateBZZBroadbandContractPDF(arrData BZZBroadbandContractPDFStruct) error {
	var err error

	if strings.ToLower(arrData.LangCode) == "zh" {
		err = GenerateBZZBroadbandContractZhPDF(arrData)
	} else {
		err = GenerateBZZBroadbandContractENPDF(arrData)
	}
	return err
}

func GenerateBZZBroadbandContractENPDF(arrData BZZBroadbandContractPDFStruct) error {

	pdfPath := "./docs/templates/sales/broadband/sec_swarm_server_leasing_contract_broadband_en.pdf"
	// pdf, _ := goPDF.GetDefaultPDFConfiguration()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A
	pdf.AddPage()

	pdf.AddTTFFont("Times", "./pkg/gopdf/ttf/times.ttf")            // std alphabet
	pdf.AddTTFFont("namum", "./pkg/gopdf/ttf/NanumBarunGothic.ttf") // kr
	pdf.AddTTFFont("loma", "./pkg/gopdf/ttf/Loma.ttf")              // thai

	// Import page 1
	tpf := pdf.ImportPage(pdfPath, 1, "/MediaBox")
	// Draw pdf onto page 1
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 16)
	pdf.SetX(190)
	pdf.SetY(713)
	pdf.Text(arrData.SerialNumber)

	// Import page 2
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 2, "/MediaBox")
	// Draw pdf onto page 2
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 14)
	pdf.SetX(188)
	pdf.SetY(182)
	pdf.Text(arrData.SerialNumber)

	nickNameFont := "Times"
	if strings.ToLower(arrData.LangCode) == "kr" {
		nickNameFont = "namum"
	} else if strings.ToLower(arrData.LangCode) == "th" {
		nickNameFont = "loma"
	}
	pdf.SetFont(nickNameFont, "", 14)
	pdf.SetX(166)
	pdf.SetY(279)
	pdf.Text(arrData.NickName)

	// Import page 3
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 3, "/MediaBox")

	pdf.SetFont("Times", "", 12)
	pdf.SetX(190)
	pdf.SetY(648)
	pdf.Text(arrData.TotalNodes)

	pdf.SetFont("Times", "", 12)
	pdf.SetX(252)
	pdf.SetY(668)
	pdf.Text(arrData.Months)

	// Draw pdf onto page 3
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 4
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 4, "/MediaBox")
	// Draw pdf onto page 4
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 5
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 5, "/MediaBox")
	// Draw pdf onto page 5
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 6
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 6, "/MediaBox")
	// Draw pdf onto page 6
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	err := pdf.WritePdf("./docs/member/sales/broadband/" + arrData.DocNo + "_broadband_contract_en.pdf")

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

func GenerateBZZBroadbandContractZhPDF(arrData BZZBroadbandContractPDFStruct) error {
	pdfPath := "./docs/templates/sales/broadband/sec_swarm_server_leasing_contract_broadband_zh.pdf"

	// pdf, _ := goPDF.GetDefaultPDFConfiguration()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A
	pdf.AddPage()

	pdf.AddTTFFont("Times", "./pkg/gopdf/ttf/times.ttf")   // std alphabet
	pdf.AddTTFFont("simhei", "./pkg/gopdf/ttf/simhei.ttf") // zh

	// Import page 1
	tpf := pdf.ImportPage(pdfPath, 1, "/MediaBox")
	// Draw pdf onto page 1
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 16)
	pdf.SetX(192)
	pdf.SetY(738)
	pdf.Text(arrData.SerialNumber)

	// Import page 2
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 2, "/MediaBox")
	// Draw pdf onto page 2
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	pdf.SetFont("Times", "", 14)
	pdf.SetX(177)
	pdf.SetY(182)
	pdf.Text(arrData.SerialNumber)

	pdf.SetFont("simhei", "", 14)
	pdf.SetX(166)
	pdf.SetY(279)
	pdf.Text(arrData.NickName)

	// Import page 3
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 3, "/MediaBox")

	pdf.SetFont("Times", "", 12)
	pdf.SetX(175)
	pdf.SetY(630)
	pdf.Text(arrData.TotalNodes)

	pdf.SetFont("Times", "", 12)
	pdf.SetX(220)
	pdf.SetY(656)
	pdf.Text(arrData.Months)

	// Draw pdf onto page 3
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 4
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 4, "/MediaBox")
	// Draw pdf onto page 4
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 5
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 5, "/MediaBox")
	// Draw pdf onto page 5
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	// Import page 6
	pdf.AddPage()
	tpf = pdf.ImportPage(pdfPath, 6, "/MediaBox")
	// Draw pdf onto page 6
	pdf.UseImportedTemplate(tpf, 0, 0, 0, 0)

	err := pdf.WritePdf("./docs/member/sales/broadband/" + arrData.DocNo + "_broadband_contract_zh.pdf")

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}
