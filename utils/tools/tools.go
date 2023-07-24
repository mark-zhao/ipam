package tools

import (
	"fmt"
	"ipam/component"
	"ipam/utils/logging"
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// 生产随机数
func NewRandString() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logging.Error(err)
	}
	return fmt.Sprintf("%x-%x", b[0:4], b[4:6])
}

// 判断是否存在
func IsExistItem(value interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(value, s.Index(i).Interface()) {
				return true
			}
		}
	}
	return false
}

// 权限认证
func FunAuth(c *gin.Context, model, method string) (string, bool) {
	if claims, T := c.Get("claims"); T == false {
		return "", false
	} else {
		if user, ok := claims.(*component.CustomClaims); ok {
			if IsExistItem(model, user.Role) || IsExistItem(method, user.Role) || IsExistItem("admin", user.Role) {
				return user.Name, true
			}
		} else {
			logging.Debug("解析失败")
			return "", false
		}
	}
	return "", false
}

// 删除切片指定元素
func RemoveElement(slice []string, elem string) []string {
	for i, v := range slice {
		if v == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// 时间转字符串
func DateToString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// strings 切片去重
func RemoveDuplicateString(list []string) []string {
	sort.Strings(list)
	result := []string{}
	var last string
	for i, v := range list {
		if i == 0 || v != last {
			result = append(result, v)
			last = v
		}
	}
	return result
}
