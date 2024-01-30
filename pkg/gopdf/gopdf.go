package goPDF

import (
	"net/http"

	"github.com/signintech/gopdf"
	"github.com/yapkah/go-api/pkg/e"
)

var DefPdfFontSize = 14

type FontFamilyInfoStruct struct {
	FontFamilyName string
	FilePath       string
}

func GetDefaultFontFamilyInfo() []FontFamilyInfoStruct {
	arrFontFamilyInfo := make([]FontFamilyInfoStruct, 0)

	arrFontFamilyInfo = append(arrFontFamilyInfo,
		FontFamilyInfoStruct{FontFamilyName: " HDZB_5", FilePath: "./pkg/gopdf/ttf/wts11.ttf"},              // china
		FontFamilyInfoStruct{FontFamilyName: " TakaoPGothic", FilePath: "./pkg/gopdf/ttf/TakaoPGothic.ttf"}, // japan
		FontFamilyInfoStruct{FontFamilyName: " loma", FilePath: "./pkg/gopdf/ttf/Loma.ttf"},                 // thai
		FontFamilyInfoStruct{FontFamilyName: " namum", FilePath: "./pkg/gopdf/ttf/NanumBarunGothic.ttf"},    // korean
		FontFamilyInfoStruct{FontFamilyName: " Roboto", FilePath: "./pkg/gopdf/ttf/Roboto-Regular.ttf"},     // test composite glyph
		FontFamilyInfoStruct{FontFamilyName: " Times", FilePath: "./pkg/gopdf/ttf/times.ttf"},               // std alphabet
	)
	return arrFontFamilyInfo
}

func GetDefaultPDFConfiguration() (gopdf.GoPdf, error) {

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A4

	arrFontFamilyInfo := GetDefaultFontFamilyInfo()

	for _, arrFontFamilyInfoV := range arrFontFamilyInfo {

		// start add diff language font
		err := pdf.AddTTFFont(arrFontFamilyInfoV.FontFamilyName, arrFontFamilyInfoV.FilePath)
		if err != nil {
			return pdf, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Data: map[string]interface{}{"action": "AddTTFFont", "data": arrFontFamilyInfoV}, Msg: err.Error()}
		}
		// end add diff language font

		// start set diff language font
		// err = pdf.SetFont(arrFontFamilyInfoV.FontFamilyName, "", DefPdfFontSize)
		// if err != nil {
		// 	return pdf, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Data: map[string]interface{}{"action": "SetFont", "data": arrFontFamilyInfoV}, Msg: err.Error()}
		// }
		// end add diff language font
	}

	return pdf, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// func DownloadFile(filepath string, url string) error {
// 	// Get the data
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	// Create the file
// 	out, err := os.Create(filepath)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()

// 	// Write the body to file
// 	_, err = io.Copy(out, resp.Body)
// 	return err
// }
