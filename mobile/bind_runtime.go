package mobile

// gomobile bind 在构建时生成引用 golang.org/x/mobile/bind 的胶水代码，
// 但源码里没有任何地方直接 import 它，`go mod tidy` 会把该依赖删掉，
// 导致流水线里的 bind 步骤找不到运行时。这里用空导入把它钉在依赖图里。
import _ "golang.org/x/mobile/bind"
