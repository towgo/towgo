/*
 * @Author       : lvyitao lvyitao@fanhaninfo.com
 * @Date         : 2024-05-17 16:26:42
 * @LastEditTime : 2024-05-23 09:18:17
 */
package accountcenter

import (
	"errors"
	"log"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
)

func CreateOrUpdateThirdDepartmentService(departmentList []Department) error {
	var getExistDepartmentList []Department
	basedboperat.SqlQueryScan(&getExistDepartmentList, "select * from departments")
	thirdExistDepartmentMap := make(map[string]Department)
	for _, v := range getExistDepartmentList {
		if v.ThirdId != "" {
			thirdExistDepartmentMap[v.ThirdId] = v
		}
	}
	// var nextCreateDepartmentList []Department
	var batchCreateDepartmentList []Department
	for _, v := range departmentList {
		_, ok := thirdExistDepartmentMap[v.ThirdId]
		if !ok {
			batchCreateDepartmentList = append(batchCreateDepartmentList, v)
		}
	}
	// 批量创建
	if len(batchCreateDepartmentList) > 0 {
		_, err := basedboperat.Create(&batchCreateDepartmentList)
		if err != nil {
			log.Println("第三方部门创建报错:" + err.Error())
			return err
		}
	}
	// 这次是用来更新父级信息(因为他们数据有问题,可能没有父级信息)
	var secondGetExistDepartmentList []Department
	basedboperat.SqlQueryScan(&secondGetExistDepartmentList, "select * from departments")
	secondThirdExistDepartmentMap := make(map[string]Department)
	for _, v := range getExistDepartmentList {
		if v.ThirdId != "" {
			secondThirdExistDepartmentMap[v.ThirdId] = v
		}
	}
	for _, v := range departmentList {
		itSelfDepart, ok := secondThirdExistDepartmentMap[v.ThirdId]
		if !ok {
			log.Println("有条数据之前没有创建成功", v.ThirdId)
			continue
		}
		if v.ThirdParentId != itSelfDepart.ThirdParentId || itSelfDepart.Fid == 0 {
			parentDepart, ok := thirdExistDepartmentMap[v.ThirdParentId]
			if !ok {
				log.Println("有条数据之前没有创建成功", v.ThirdId)
				continue
			} else {
				v.Fid = parentDepart.ID
				err := basedboperat.Update(&v, nil, "id = ?", itSelfDepart.ID)
				if err != nil {
					log.Println("第三方部门更新报错:" + err.Error())
					return err
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
	return nil
}

func (d *Department) GetOneLevelDepartmentListVerify() error {
	if d.ID != 0 {
		var getDepartment Department
		basedboperat.Get(&getDepartment, nil, "id = ?", d.ID)
		if getDepartment.ID == 0 {
			return errors.New("部门id对应父级数据不存在")
		}
	}
	return nil
}
