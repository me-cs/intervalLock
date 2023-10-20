# lazyLock

*lazyLock go 实现*

[![Go](https://github.com/me-cs/lazyLock/workflows/Go/badge.svg?branch=main)](https://github.com/me-cs/lazyLock/actions)
[![codecov](https://codecov.io/gh/me-cs/lazyLock/branch/main/graph/badge.svg)](https://codecov.io/gh/me-cs/lazyLock)
[![Release](https://img.shields.io/github/v/release/me-cs/lazyLock.svg?style=flat-square)](https://github.com/me-cs/lazyLock)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/lazyLock)](https://goreportcard.com/report/github.com/me-cs/lazyLock)
[![Go Reference](https://pkg.go.dev/badge/github.com/me-cs/lazyLock.svg)](https://pkg.go.dev/github.com/me-cs/lazyLock)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 描述
lazyLock 非常适合业务场景。在编写特定业务代码时
资源/场景上的并发操作可能不会太多、
因此您不想为它们持有一个互斥器，但又担心对其进行并发操作。
使用 lazyLock 可以在发生冲突时保证互斥操作、
并在冲突消失时自动释放资源。

举个例子：网站访问数据的大部分特点是 "二八定律"：
80% 的业务访问集中在 20% 的数据中。
我们可以使用 lazyLock 来锁定发生热访问的数据、
而不用理会其他 80% 的冷数据。


简体中文 | [English](README.md)

### 示例用法:
```go
package main

import "github.com/me-cs/lazyLock"

func main() {

	unlock := lazyLock.Lock("key")

	//需要被保护的代码

	unlock()

}
```