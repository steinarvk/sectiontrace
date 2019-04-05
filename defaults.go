package sectiontrace

import "os"

var DebugMode bool = false

var DefaultCategory string = "Section"
var DefaultScope string = ""

var ProcessID int32 = int32(os.Getpid())

var DefaultDisplayTimeUnit string = "ms"
var DefaultOtherData = map[string]interface{}{}
