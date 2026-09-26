// Package web 把网页文件（index.html、css/、js/、favicon.ico）打包进 exe。
package web

import "embed"

//go:embed index.html favicon.ico css js
var Files embed.FS
