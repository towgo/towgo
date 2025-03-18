package accountcenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/towgo/towgo/dao/basedboperat"
)

type DepartmentNode struct {
	Department
	Childs []*DepartmentNode `json:"child"`
}

func (dn *DepartmentNode) CreateChild(d *DepartmentNode) {
	dn.Childs = append(dn.Childs, d)
}

func (dn *DepartmentNode) CreatedChildWithDepartmentRows(ds []Department) {
	for _, department := range ds {
		if dn.ID == department.Fid {
			newDepartmentNode := &DepartmentNode{
				Department: department,
			}
			dn.CreateChild(newDepartmentNode)
			newDepartmentNode.CreatedChildWithDepartmentRows(ds)
		}
	}
}

type Departments struct {
	Rows []Department
}

func (Department) TableName() string {
	return "departments"
}

func (*Department) CacheExpire() int64 {
	return 5000
}

type Department struct {
	ID              int64             `json:"id"`                    //部门ID
	Fid             int64             `json:"fid" xorm:"default(0)"` //上级部门
	Parent          *ParentDepartment `json:"parent" gorm:"-" xorm:"-"`
	Name            string            `json:"name"`                            //部门名称唯一key
	ThirdId         string            `json:"third_id"`                        //第三方组织id
	ThirdParentId   string            `json:"third_parent_id"`                 //第三方父级组织id
	Nickname        string            `json:"nickname"`                        //部门名称
	Jurisdictions   []int64           `json:"jurisdictions" gorm:"-" xorm:"-"` //部门权限
	HaveChildren    bool              `json:"have_children" gorm:"-" xorm:"-"` //是否有子集
	Remark          string            `json:"remark"`                          //备注
	DepartmentAdmin string            `json:"department_admin"`                //部门管理员
	ChildDepartment []Department      `json:"child_department" xorm:"-"`       //子集部门
	Users           []Account         `json:"users" xorm:"-"`
	UpdatedAt       int64             `json:"updated_at"`
	CreatedAt       int64             `json:"created_at"`
}

func (ParentDepartment) TableName() string {
	return "departments"
}

func (*ParentDepartment) CacheExpire() int64 {
	return 5000
}

type ParentDepartment struct {
	ID            int64   `json:"id"`                              //部门ID
	Name          string  `json:"name"`                            //部门名称唯一key
	Nickname      string  `json:"nickname"`                        //部门名称
	Jurisdictions []int64 `json:"jurisdictions" gorm:"-" xorm:"-"` //部门权限
	Remark        string  `json:"remark"`                          //备注
	UpdatedAt     int64   `json:"updated_at"`
	CreatedAt     int64   `json:"created_at"`
}

func (DepartmentAccounts) TableName() string {
	return "departments_accounts"
}

func (*DepartmentAccounts) CacheExpire() int64 {
	return 5000
}

type DepartmentAccounts struct {
	ID           int64 `json:"id"`            //部门ID
	DepartmentId int64 `json:"department_id"` //上级部门
	AccountId    int64 `json:"account_id"`    //部门名称唯一key
}

func (d *Department) Create() (int64, error) {
	d.ID = 0
	rowsAffected, err := basedboperat.Create(d)
	if err != nil {
		return 0, err
	}
	if d.ID <= 0 {
		return 0, errors.New("identity id not be 0")
	}
	rowsAffected2, err := d.CreateRelation()
	rowsAffected = rowsAffected + rowsAffected2

	return rowsAffected, err
}

func (d *Department) Update() error {
	var findModel Department
	basedboperat.Get(&findModel, nil, "id = ?", d.ID)
	if findModel.ID <= 0 {
		return errors.New("记录不存在")
	}
	findModel.DeleteRelation()
	findModel.CreateRelation()
	basedboperat.Update(d, nil, "id = ?", d.ID)
	return nil
}

func (d *Department) Delete() (int64, error) {
	d.DeleteRelation()
	return basedboperat.Delete(d, d.ID, nil)
}

func (d *Department) AfterQuery() {

	if d.Fid > 0 {
		var parentDepartment ParentDepartment
		basedboperat.Get(&parentDepartment, nil, "id = ?", d.Fid)
		if parentDepartment.ID > 0 {
			d.Parent = &parentDepartment
		}
	}

	var dj DepartmentJurisdiction
	var djs []DepartmentJurisdiction
	basedboperat.SqlQueryScan(&djs, "select * from "+dj.TableName()+" where department_id = ?", d.ID)
	for _, v := range djs {
		d.Jurisdictions = append(d.Jurisdictions, v.JurisdictionId)
	}
}

// 创建关联数据
func (d *Department) CreateRelation() (int64, error) {
	var rowsAffected int64
	for _, v := range d.Jurisdictions {
		var dj DepartmentJurisdiction
		dj.JurisdictionId = v
		dj.DepartmentId = d.ID
		rowsAffected2, _ := basedboperat.Create(&dj)
		rowsAffected = rowsAffected + rowsAffected2
	}
	return rowsAffected, nil
}

// 删除关联数据
func (d *Department) DeleteRelation() {
	//var da DepartmentAccounts
	var dj DepartmentJurisdiction
	basedboperat.SqlExec("delete from "+dj.TableName()+" where department_id = ?", d.ID)
	//basedboperat.SqlExec("delete from "+da.TableName()+" where department_id = ?", d.ID)
}

func GetDepartmentTreeList() []interface{} {
	var tree []interface{}
	department := Department{}
	departments := []Department{}

	list := &basedboperat.List{}
	list.Limit = -1
	basedboperat.ListScan(list, department, &departments)

	//确定根部门
	for _, department := range departments {
		if department.Fid == 0 {
			departmentTreeRoot := DepartmentNode{
				Department: department,
			}
			departmentTreeRoot.CreatedChildWithDepartmentRows(departments)
			tree = append(tree, departmentTreeRoot)
		}
	}
	return tree
}

// 递归通过当前部门ID查找所有下级部门id
func GetDepartmentTreeListId(departmentsid []int64, jdepartments *[]int64) []int64 {

	if jdepartments == nil {
		jdepartments = &[]int64{}
		for _, v := range departmentsid {
			*jdepartments = append(*jdepartments, v)
		}
	}

	list := &basedboperat.List{}
	list.Limit = -1

	var where []basedboperat.Condition

	where = append(where, basedboperat.Condition{Field: "fid", Operator: "IN"})

	list.Where = where

	department := Department{}
	departments := []Department{}

	basedboperat.ListScan(list, department, &departments)

	departmentsid = []int64{}
	for _, d := range departments {
		departmentsid = append(departmentsid, d.ID)
		*jdepartments = append(*jdepartments, d.ID)
	}
	//存在子部门继续递归
	if len(departmentsid) > 0 {
		GetDepartmentTreeListId(departmentsid, jdepartments)
	}
	return *jdepartments
}

func (d *Departments) List() string {
	var jsonStr string
	m := d.Rows
	mjson, _ := json.Marshal(m)
	jsonStr = string(mjson)
	return jsonStr
}

func (d *Departments) ToString() string {
	var jsonStr string
	m := d.Rows
	mjson, _ := json.Marshal(m)
	jsonStr = string(mjson)
	return jsonStr
}

func GetDepartmentChildIdList(fidList []int64, idList *[]int64) error {
	if len(fidList) == 0 {
		return nil
	}
	var childDepartIdList []Department
	fidStrList := TranslateSqlInAny(fidList)
	sql := fmt.Sprintf("select id from departments where fid in (%s)", strings.Join(fidStrList, ","))
	err := basedboperat.SqlQueryScan(&childDepartIdList, sql)
	if err != nil {
		log.Println(err)
		return err
	}
	var childIdList []int64
	for _, v := range childDepartIdList {
		*idList = append(*idList, v.ID)
		childIdList = append(childIdList, v.ID)
	}
	GetDepartmentChildIdList(childIdList, idList)
	return nil
}

func TranslateSqlInAny(strList interface{}) []string {
	var strRoles []string
	switch v1 := strList.(type) {
	case []float64:
		for _, v := range v1 {
			strRoles = append(strRoles, strconv.FormatInt(int64(v), 10))
		}
	case []int64:
		for _, v := range v1 {
			strRoles = append(strRoles, strconv.FormatInt(v, 10))
		}

	case []string:
		for _, v := range v1 {
			strRoles = append(strRoles, strconv.Quote(v))
		}
	}
	return strRoles
}

type DepartmentIdList struct {
	IdList []int64 `json:"id_list"`
}
