package licenseterminal

type License struct {
	ID                  int64  `json:"id"`                    //Licenseid
	LicenseKey          string `json:"license_key"`           //许可证密钥
	AccessCode          string `json:"access_code"`           //授权码
	AccessCodeVersion   string `json:"access_code_version"`   //授权码的版本
	OrderSerialNumber   string `json:"order_serial_number"`   //订单编号
	ProductSerialNumber string `json:"product_serial_number"` //产品序列号
	ProductNumber       string `json:"product_number"`        //产品编号
	ProductName         string `json:"product_name"`          //产品名称
	Starttime           int64  `json:"starttime"`             //起始日
	Endtime             int64  `json:"endtime"`               //截止日期
	Sale                string `json:"sale"`                  //销售人员
	CustomerName        string `json:"customer_name"`         //客户名称
	Remark              string `json:"remark"`                //备注
	Valid               bool   `json:"valid"`                 //是否有效（非存储结构）
	ValidErrMessage     string `json:"valid_err_message"`     //无效的错误信息
}

type ActiveRequestInfo struct {
	ProductNumber     string `json:"product_number"`
	AccessCode        string `json:"access_code"`
	AccessCodeVersion string `json:"access_code_version"`
	LicenseKey        string `json:"license_key"`
}
