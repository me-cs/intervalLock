# lazyLock

*lazyLock go implementation*

[![Go](https://github.com/me-cs/lazyLock/workflows/Go/badge.svg?branch=main)](https://github.com/me-cs/lazyLock/actions)
[![codecov](https://codecov.io/gh/me-cs/lazyLock/branch/main/graph/badge.svg)](https://codecov.io/gh/me-cs/lazyLock)
[![Release](https://img.shields.io/github/v/release/me-cs/lazyLock.svg?style=flat-square)](https://github.com/me-cs/lazyLock)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/lazyLock)](https://goreportcard.com/report/github.com/me-cs/lazyLock)
[![Go Reference](https://pkg.go.dev/badge/github.com/me-cs/lazyLock.svg)](https://pkg.go.dev/github.com/me-cs/lazyLock)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description
lazyLock is great for business scenarios. When writing business-specific code,
there may not be too many concurrent operations on a resource/scenario,
so you don't want to hold a mutex for them, but you're worried about concurrent operations on it. 
Using lazyLock, it provides the ability to guarantee mutually exclusive operations when a conflict occurs,
and to automatically release resources when the conflict disappears.

For example: most of the characteristics of the website access data in the "law of two or eight": 
80% of the business access is concentrated in 20% of the data. 
We can use lazyLock to lock the data where the hot access occurs, 
without paying attention to the other 80% of the cold data.

English 

### Example use: