//go:build !windows

package address

import "os"

func isFileExecutable(path string) (ok bool, bool error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// 0111 matches any executable bit (user, group, other)
	ok = info.Mode().Perm()&0111 != 0
	return ok, nil
}
