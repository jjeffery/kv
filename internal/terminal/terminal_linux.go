// This code was adapted from the following package:
//
//  golang.org/x/crypt/ssh/terminal.
//
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package terminal

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TCGETS
