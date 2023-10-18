# intervalLock

*intervalLock go implementation*

[![Go](https://github.com/me-cs/intervalLock/workflows/Go/badge.svg?branch=main)](https://github.com/me-cs/intervalLock/actions)
[![codecov](https://codecov.io/gh/me-cs/intervalLock/branch/main/graph/badge.svg)](https://codecov.io/gh/me-cs/intervalLock)
[![Release](https://img.shields.io/github/v/release/me-cs/intervalLock.svg?style=flat-square)](https://github.com/me-cs/intervalLock)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/intervalLock)](https://goreportcard.com/report/github.com/me-cs/intervalLock)
[![Go Reference](https://pkg.go.dev/badge/github.com/me-cs/intervalLock.svg)](https://pkg.go.dev/github.com/me-cs/intervalLock)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description
intervalLock is great for business scenarios. When writing business-specific code,
there may not be too many concurrent operations on a resource/scenario,
so you don't want to hold a mutex for them, but you're worried about concurrent operations on it. 
Using intervalLock, it provides the ability to guarantee mutually exclusive operations when a conflict occurs,
and to automatically release resources when the conflict disappears.


English 

### Example use: