package debug

import (
	"fmt"
	"net/http"
)

var (
	// ServiceKey  注册到接口平台所用的Key.
	ServiceKey = ""
	// GitTime git log中记录的提交时间.
	GitTime = ""
	// GitHash git commit hash.
	GitHash = ""
	// GitMessage git log 中记录的提交信息.
	GitMessage = ""
)

// Print 输出当前程序编译信息.
func Print() {
	fmt.Printf("Service Key: %s\n", ServiceKey)
	fmt.Printf("Commit Hash: %s\n", GitHash)
	fmt.Printf("Commit Time: %s\n", GitTime)
	fmt.Printf("Commit Message: %s\n", GitMessage)
}

//Version 版本信息.
type Version struct {
}

//GET 输出当前应用版本信息.
func (v *Version) GET(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ServiceKey: %s\n", ServiceKey)
	fmt.Fprintf(w, "Commit Hash: %s\n", GitHash)
	fmt.Fprintf(w, "Commit Time: %s\n", GitTime)
	fmt.Fprintf(w, "Commit Message: %s\n", GitMessage)
}
