// Package config 存放默认设定档。
//
// 这两个文件会在编译时打包进 exe；程序第一次运行时，
// 若运行时文件夹里的 config/ 还没有设定档，就用它们建立。
//
//	settings.ini  程序（网页服务、日志）设置
//	app.json      网页显示设置（学校名称、GitHub 网址…）
package config

import _ "embed"

//go:embed settings.ini
var SettingsINI []byte

//go:embed app.json
var AppJSON []byte
