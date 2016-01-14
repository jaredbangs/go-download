package download

import (
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

type Download struct {
	EnableLogging                bool
	LogAllHeaders                bool
	ShouldDownloadResponseFilter func(*http.Response) bool
}

func (d *Download) DownloadFile(url string, downloadToDirectoryPath string) (downlaodedFilePath string, err error) {

	//transport := &RedirectHandlingTransport{}
	//client := &http.Client{Transport: transport}

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if d.ShouldDownloadResponseFilter == nil || d.ShouldDownloadResponseFilter(resp) {

		filePath, err := d.getTargetFilePath(downloadToDirectoryPath, url, resp)
		if err != nil {
			log.Println(err)
			return "", err
		}

		out, err := os.Create(filePath)
		if err != nil {
			log.Println(err)
			return "", err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Println(err)
			return "", err
		}

		log.Println("Downloaded " + filePath)
		return filePath, nil
	} else {

		return "", nil
	}
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
		log.Println("Content-Disposition: " + contentDisposition)
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
