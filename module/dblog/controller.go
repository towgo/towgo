package dblog

import (
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/towgo"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/system"
	"github.com/xuri/excelize/v2"
)

func InitDbLogApi() {
	towgo.SetFunc("/system/log/list", logList)
	towgo.SetFunc("/system/log/create", logCreate)
	http.HandleFunc("/api/system/download/log/file", downloadLogFile)
	basedboperat.Sync(new(Log), new(OperateType))
	// clearExp()

	go BatchInsert(
		NewOperateType(UPLOAD, "/system/log/list", "account_center", "查询日志列表"),
		NewOperateType(UPDATE, "/system/log/create", "account_center", "增加日志记录"),
	)
}

func logList(rpcConn towgo.JsonRpcConnection) {
	var log Log
	var logList []Log
	var list basedboperat.List
	rpcConn.ReadParams(&list)
	basedboperat.ListScan(&list, &log, &logList)
	if list.Error != nil {
		rpcConn.GetRpcResponse().Error.Set(500, list.Error.Error())
		rpcConn.Write()
		return
	}
	result := make(map[string]any)
	result["count"] = list.Count
	result["rows"] = logList
	rpcConn.WriteResult(result)
}

func logCreate(rpcConn towgo.JsonRpcConnection) {
	var log Log
	rpcConn.ReadParams(&log)
	log.ID = 0
	_, err := basedboperat.Create(&log)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func downloadLogFile(w http.ResponseWriter, req *http.Request) {
	res := make(map[string]any)
	if req.Method != "POST" {
		res["code"] = 500
		res["message"] = "method不支持"
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	excelFile := excelize.NewFile()
	defer func() {
		if err := excelFile.Close(); err != nil {
			log.Println("文件关闭失败")
			return
		}
	}()
	// session := req.Header.Get("Session")

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res["code"] = 500
		res["message"] = err.Error()
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	defer req.Body.Close()
	var downloadLogFileReq DownloadLogFileReq
	if len(reqBody) != 0 {
		log.Println(string(reqBody))
		err = json.Unmarshal(reqBody, &downloadLogFileReq)
		if err != nil {
			res["code"] = 500
			res["message"] = "json解析body报错"
			res["data"] = ""
			byteResponse, _ := json.Marshal(res)
			w.Write(byteResponse)
			return
		}
	}

	sw, err := excelFile.NewStreamWriter("Sheet1")
	if err != nil {
		res["code"] = 500
		res["message"] = "method不支持"
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	sql := "select * from logs"
	args := []any{}
	if downloadLogFileReq.StartTime != 0 {
		args = append(args, downloadLogFileReq.StartTime)
		if !strings.Contains(sql, "where") {
			sql = fmt.Sprintf(sql + " where created_at > ?")
		} else {
			sql = fmt.Sprintf(sql + " and created_at > ?")
		}
	}
	if downloadLogFileReq.EndTime != 0 {
		args = append(args, downloadLogFileReq.EndTime)
		if !strings.Contains(sql, "where") {
			sql = fmt.Sprintf(sql + " where created_at < ?")
		} else {
			sql = fmt.Sprintf(sql + " and created_at < ?")
		}
	}
	sql = sql + " ORDER BY id desc"
	interval := 1000
	sw.SetRow("A1", []interface{}{"id", "服务名", "服务ip", "模块名", "协议", "请求参数", "返回参数", "请求ip", "方法路径", "操作人", "创建时间", "更新时间"})
	i := 0

	for {
		newSql := fmt.Sprintf(sql+" LIMIT %d,%d", i*interval, interval)
		var logList []Log
		basedboperat.SqlQueryScan(&logList, newSql, args...)
		if len(logList) == 0 {
			break
		}
		rowNum := 2
		for _, v := range logList {
			cell := fmt.Sprintf("A%d", rowNum)
			var rowData []interface{}
			rowData = append(rowData, v.ID, v.ServiceName, v.ServiceIp, v.ModuleName, v.Protocol, v.RequestParam, v.ResponseParam, v.RequestIp, v.Method, v.CreatedBy)
			rowData = append(rowData, time.Unix(v.UpdatedAt, 0).Format(time.DateTime), time.Unix(v.UpdatedAt, 0).Format(time.DateTime))
			sw.SetRow(cell, rowData)
			rowNum += 1
		}
		i += 1
	}
	err = sw.Flush()
	if err != nil {
		res["code"] = 500
		res["message"] = "数据刷入失败"
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	uuid := GetUuid()
	logFileBasePath := filepath.Join(system.GetPathOfProgram(), "log_file")
	PathIfNotExistsCreate(logFileBasePath)
	fileName := fmt.Sprintf("log_data_%s.xlsx", uuid)
	theWholeFilePathName := filepath.Join(logFileBasePath, fileName)
	if err := excelFile.SaveAs(theWholeFilePathName); err != nil {
		res["code"] = 500
		res["message"] = "数据刷入失败"
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}

	file, err := os.Stat(theWholeFilePathName)
	if err != nil {
		res["code"] = 500
		res["message"] = err.Error()
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	f, err := os.Open(theWholeFilePathName)
	if err != nil {
		res["code"] = 500
		res["message"] = err.Error()
		res["data"] = ""
		byteResponse, _ := json.Marshal(res)
		w.Write(byteResponse)
		return
	}
	//把file读取到缓冲区中
	defer f.Close()
	// var chunk []byte
	buf := make([]byte, 1024)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size()))
	for {
		//从file读取到buf中
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			res["code"] = 500
			res["message"] = err.Error()
			res["data"] = ""
			byteResponse, _ := json.Marshal(res)
			w.Write(byteResponse)
			return
		}
		//说明读取结束
		if n == 0 {
			break
		}
		//读取到最终的缓冲区中
		// chunk = append(chunk, buf[:n]...)
		w.Write(buf[:n])
	}
}
