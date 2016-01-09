package download

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

type Download struct {
	EnableLogging bool
	LogAllHeaders bool
	LogToFilePath string
}

func (d *Download) DownloadFile(url string, downloadToDirectoryPath string) (err error) {

	//transport := &RedirectHandlingTransport{}
	//client := &http.Client{Transport: transport}

	resp, err := http.Get(url)
	if err != nil {
		d.logError(err)
		return err
	}
	defer resp.Body.Close()

	filePath, err := d.getTargetFilePath(downloadToDirectoryPath, url, resp)
	if err != nil {
		d.logError(err)
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		d.logError(err)
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		d.logError(err)
		return err
	}

	d.log("Downloaded " + filePath)

	return nil
}

func (d *Download) TrimExtraPartsFromFileName(fileName string) (adjustedFileName string) {

	if strings.Contains(fileName, "?") {
		adjustedFileName = strings.Split(fileName, "?")[0]
	} else {
		adjustedFileName = fileName
	}

	return adjustedFileName
}

func (d *Download) getTargetFilePath(downloadToDirectoryPath string, url string, resp *http.Response) (filePath string, err error) {

	// If redirected, be sure to use the "final" url from the response
	if url != resp.Request.URL.String() {
		url = resp.Request.URL.String()
	}

	fileNamePart := path.Base(url)

	d.logAllResponseHeaders(resp)

	contentDisposition := resp.Header.Get("Content-Disposition")

	if contentDisposition != "" {

		_, params, err := mime.ParseMediaType(contentDisposition)

		if err == nil && params["filename"] != "" {
			fileNamePart = params["filename"]
		}
		d.log(contentDisposition)
	}

	fileNamePart = d.TrimExtraPartsFromFileName(fileNamePart)

	filePath = path.Join(downloadToDirectoryPath, fileNamePart)

	return filePath, err
}

func (d *Download) logAllResponseHeaders(resp *http.Response) {

	if d.LogAllHeaders {
		for k, v := range resp.Header {
			log.Println("Header:", k, "value:", v)
		}
	}
}

func (d *Download) logError(err error) {
	d.log(err.Error())
}

func (d *Download) log(text string) {

	if d.EnableLogging {

		if d.LogToFilePath != "" {

			f, err := os.OpenFile(d.LogToFilePath, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if _, err = f.WriteString(text + "\n"); err != nil {
				panic(err)
			}
		}

		fmt.Println(text)
	}
}
