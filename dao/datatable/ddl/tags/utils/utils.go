package utils

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
)

func GetAcTableName(tableName string) string {
	return "datatable_" + tableName
}

func IsTableExist(tableName string) bool {
	db := xormDriver.DbMaster()
	res, err := db.Engine.IsTableExist(tableName)
	if err != nil {
		log.Println(err)
		return false
	}
	return res
}

// IndexName returns index name
func IndexName(tableName, idxName string) string {
	return fmt.Sprintf("IDX_%v_%v", tableName, idxName)
}

// SeqName returns sequence name for some table
func SeqName(tableName string) string {
	return "SEQ_" + strings.ToUpper(tableName)
}

// SliceEq return true if two slice have the same elements even if different sort.
func SliceEq(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	sort.Strings(left)
	sort.Strings(right)
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// IndexSlice search c in slice s and return the index, return -1 if s don't contain c
func IndexSlice(s []string, c string) int {
	for i, ss := range s {
		if c == ss {
			return i
		}
	}
	return -1
}

// ReflectValue returns value of a bean
func ReflectValue(bean interface{}) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(bean))
}

// IsSubQuery returns true if it contains a sub query
func IsSubQuery(tbName string) bool {
	const selStr = "select"
	if len(tbName) <= len(selStr)+1 {
		return false
	}

	return strings.EqualFold(tbName[:len(selStr)], selStr) ||
		strings.EqualFold(tbName[:len(selStr)+1], "("+selStr)
}
