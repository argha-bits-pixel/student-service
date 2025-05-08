package requests

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"student-service/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BulkUploadRequestHandler interface {
	Bind(c echo.Context, request interface{}, model *models.BulkUploadModel) error
}
type BulkUploadRequest struct {
	FilePath    string
	RequestType string `json:"request_type"`
}

type GetBulkUploadRequest struct {
	RequestId  string `param:"request_id"`
	Pagination Pagination
}

func (b BulkUploadRequest) Bind(c echo.Context, request interface{}, model *models.BulkUploadModel) error {
	switch request.(type) {
	case *BulkUploadRequest:
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return errors.New("file not received")
		}
		src, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("invalid file %s", err.Error())
		}
		defer src.Close()
		tmpPath := filepath.Join(os.TempDir(), fileHeader.Filename)
		dst, err := os.Create(tmpPath)
		if err != nil {
			return fmt.Errorf("unable to create temp file %s", err.Error())
		}
		defer dst.Close()
		io.Copy(dst, src)
		request.(*BulkUploadRequest).FilePath = dst.Name()
		model.RequestId = uuid.NewString()
		model.FileName = fileHeader.Filename
		model.FilePath = dst.Name()
		model.Status = "PENDING"
	case *GetBulkUploadRequest:
		err := c.Bind(request)
		if err != nil {
			log.Printf("error in binding Get Bulk Upload Request")
			return err
		}
		request.(*GetBulkUploadRequest).Pagination = GetPagination(request.(*GetBulkUploadRequest).Pagination)
	}

	return nil
}
func NewBulkUploadRequestHandler() BulkUploadRequestHandler {
	return BulkUploadRequest{}
}
