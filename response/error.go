package response

import (
	"log"
	"student-service/utils/validator"

	"github.com/labstack/echo/v4"
)

func ProcessErrorResponse(err error) interface{} {
	resp := StudentConsolidatedResposne{}
	switch v := err.(type) {
	case *validator.ValidationError:
		log.Println("error type", v)
		resp.Message = "Invalid Input"
		resp.Data = []string{}
		resp.Error = err.(*validator.ValidationError).Fields
	case *echo.HTTPError:
		log.Println("error type", v)
		if err.(*echo.HTTPError).Code == 415 {
			resp.Message = "Invalid Request"
			resp.Data = []string{}
			resp.Error = map[string]string{
				"error": "Unsupported Media Type. Please use application/json in request header Content-Type",
			}
		}
	default:
		log.Println("error type", v)
		resp.Message = "unable to process request"
		resp.Data = []string{}
		resp.Error = err.Error()
	}
	return resp
}
