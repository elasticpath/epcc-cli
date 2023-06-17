package encoding

import (
	"bytes"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"mime/multipart"
	"os"
)

func ToMultiPartEncoding(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute) (*bytes.Buffer, string, error) {
	if len(args)%2 == 1 {
		return nil, "", fmt.Errorf("the number of arguments %d supplied isn't even, json should be passed in key value pairs. Do you have an extra/missing id?", len(args))
	}

	values := map[string]string{}
	formFiles := make([]formFile, 0)

	for i := 0; i < len(args); i += 2 {

		k := args[i]
		v := args[i+1]

		if attribute, ok := attributes[k]; ok {
			if attribute.Type == "FILE" {

				fileContents, err := os.ReadFile(v)

				if err != nil {
					return nil, "", err
				}

				formFiles = append(formFiles, formFile{
					parameterName: k,
					filename:      v,
					fileContents:  fileContents,
				})

			} else {
				values[k] = v
			}

		} else {
			values[k] = v
		}

	}

	return encodeForm(values, formFiles)

}

type formFile struct {
	filename      string
	parameterName string
	fileContents  []byte
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func encodeForm(values map[string]string, formFiles []formFile) (byteBuf *bytes.Buffer, contentType string, err error) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range values {
		_ = writer.WriteField(key, val)
	}

	for _, formFile := range formFiles {

		part, err := writer.CreateFormFile(formFile.parameterName, formFile.filename)

		if err != nil {
			return nil, "", err
		}

		part.Write(formFile.fileContents)
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
