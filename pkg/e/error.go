package e

// CustomError struct
type CustomError struct {
	HTTPCode     int
	Code         int
	Msg          string
	Data         interface{}
	TemplateData map[string]interface{}
}

// Error func
func (e *CustomError) Error() string {
	if e.Msg == "" {
		return GetMsg(e.Code)
	}
	return e.Msg
}
