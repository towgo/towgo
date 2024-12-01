package oapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func GetRequest(url string, queryStrings map[string]string, destResult, destErrStruct any) error {

	request, err := http.NewRequest("GET", url, nil)
	urlvalues := request.URL.Query()

	for k, v := range queryStrings {
		urlvalues.Add(k, v)
	}

	if err != nil {
		log.Print(err.Error())
	}
	HttpClient := http.Client{}
	//request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := HttpClient.Do(request)
	if err != nil {
		log.Print(err.Error())
	}

	data, _ := io.ReadAll(response.Body)

	json.Unmarshal(data, destResult)
	json.Unmarshal(data, destErrStruct)

	return err
}

func PostRequest(url string, requestBody any, destResult, destErrStruct any) error {
	bytesData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}
	HttpClient := http.Client{}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := HttpClient.Do(request)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	//log.Print(string(data))

	if destResult != nil {
		err = json.Unmarshal(data, destResult)
		if err != nil {
			log.Print(err.Error())
		}
	}

	if destErrStruct != nil {
		err = json.Unmarshal(data, destErrStruct)
		if err != nil {
			log.Print(err.Error())
		}
	}

	return err
}

func PostRequestWithHeader(url string, requestBody any, destResult, destErrStruct any, header map[string]string) error {
	bytesData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	log.Print(string(bytesData))

	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}
	HttpClient := http.Client{}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range header {
		request.Header.Set(k, v)
	}

	response, err := HttpClient.Do(request)
	if err != nil {
		return err
	}

	data, _ := io.ReadAll(response.Body)

	log.Print(string(data))

	json.Unmarshal(data, destResult)
	json.Unmarshal(data, destErrStruct)
	return err
}

// key:file 里面放一个文件
// multipart/form-data 传一个文件
func PostFormDataWithSingleFile(url string, fullfilepath string, filename string) (string, error) {
	client := http.Client{}
	bodyBuf := &bytes.Buffer{}
	bodyWrite := multipart.NewWriter(bodyBuf)
	file, err := os.Open(fullfilepath)

	if err != nil {
		return "", err
	}
	defer file.Close()
	// file 为key
	fileWrite, err := bodyWrite.CreateFormFile("image_binary", filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fileWrite, file)
	if err != nil {
		return "", err
	}

	bodyWrite.Close() //要关闭，会将w.w.boundary刷写到w.writer中
	// 创建请求
	contentType := bodyWrite.FormDataContentType()
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		return "", err
	}

	// 设置头
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
