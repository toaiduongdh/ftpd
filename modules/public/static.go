// +build bindata

// Copyright 2017 The XORM Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package public

import (
	"gitea.com/lunny/tango"
)

// Static implements the macaron static handler for serving assets.
func Static(static string) tango.Handler {
	return tango.Static(tango.StaticOptions{
		Prefix:     "public",
		FileSystem: Assets,
	})
}
