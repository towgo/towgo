package sliceOperate

import "strconv"

/**
 * @description: 数字列表去重
 * @param {[]int64} silce
 * @return {*}
 */
func RemoveDuplicateIntElement(silce []int64) []int64 {
	result := make([]int64, 0, len(silce))
	temp := map[int64]struct{}{}
	for _, item := range silce {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

/**
 * @description: 用于计算数据库中和即将要插入数据的交集,两个差集
 * @param {*} slice1 数据库中已存在的数据
 * @param {[]int64} slice2 需要新插入的数据
 * @return {*} []int64 交集,[]int64 slice1中多出来的差集,[]int64 slice2中多出来的差集
 */
func GetListInterIntDiffer(slice1, slice2 []int64) ([]int64, []int64, []int64) {
	slice1 = RemoveDuplicateIntElement(slice1)
	slice2 = RemoveDuplicateIntElement(slice2)
	slice1m := make(map[int64]int)
	slice2m := make(map[int64]int)
	intersection := []int64{}
	slice2differ := []int64{}
	slice1differ := []int64{}
	for _, v := range slice1 {
		slice1m[v]++
	}

	for _, v := range slice2 {
		_, ok := slice1m[v]
		slice2m[v]++
		if ok {
			intersection = append(intersection, v)
		} else {
			slice2differ = append(slice2differ, v)
		}
	}
	for _, v := range slice1 {
		_, ok := slice2m[v]
		if !ok {
			slice1differ = append(slice1differ, v)
		}
	}
	return intersection, slice1differ, slice2differ
}

/**
 * @description: 转换列表成string的列表,用于拼接sql的in
 * @param {interface{}} strList
 * @return {*}
 */
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

/**
 * @description: 判断某个数据存在于某个切片中
 * @param {interface{}} list
 * @param {interface{}} sub
 * @return {*}
 */
func ListContainsAny(list interface{}, sub interface{}) bool {
	switch sub.(type) {
	case int:
		lists := list.([]int)
		for _, v := range lists {
			if sub == v {
				return true
			}
		}
	case int64:
		lists := list.([]int64)
		for _, v := range lists {
			if sub == v {
				return true
			}
		}
	case string:
		lists := list.([]string)
		for _, v := range lists {
			if sub == v {
				return true
			}
		}
	}
	return false
}
