package app

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/pkg/translation"
)

// BindAndValid binds and validates data
func BindAndValid(c *gin.Context, form interface{}) (bool, []string) {
	err := c.ShouldBind(form)

	if err != nil {
		return false, strings.Split(err.Error(), "\n")
	}

	//added by kahhou- for validate dynamic routes
	err = c.BindUri(form)
	if err != nil {
		return false, strings.Split(err.Error(), "\n")
	}
	//end add

	valid := validation.Validation{}

	check, err := valid.Valid(form)
	if err != nil {
		return false, []string{err.Error()}
	}
	if !check {
		return false, []string{vErrorToString(c, valid.Errors)}
	}

	return true, []string{}
}

// vErrorToString change validate error to string
func vErrorToString(c *gin.Context, err []*validation.Error) string {
	var template = make(map[string]interface{})

	if err[0].LimitValue != nil {
		switch t := err[0].LimitValue.(type) {
		case int:
			template["limit"] = t
		case string:
			template["limit"] = t
		case []int:
			template["min"] = t[0]
			template["max"] = t[1]
		}
	}

	eName := "beego_" + err[0].Name

	msg := trans(c, eName, template)

	if msg == eName {
		return fmt.Sprintf("%v %v", err[0].Field, err[0].Message)
	}

	name := trans(c, err[0].Field, nil)

	return fmt.Sprintf("%v %v", name, msg)
}

func trans(c *gin.Context, text string, template map[string]interface{}) string {
	l, ok := c.Get("localizer")

	if !ok {
		return text
	}

	localizer, ok := l.(*translation.Localizer)

	if !ok {
		return text
	}

	return localizer.Trans(text, template)
}

func FormValidation(c *gin.Context, form interface{}) (bool, []string) {
	valid := validation.Validation{}

	check, err := valid.Valid(form)
	if err != nil {
		return false, []string{err.Error()}
	}
	if !check {
		return false, []string{vErrorToString(c, valid.Errors)}
	}
	return true, []string{}
}
