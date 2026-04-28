//go:build windows

package output

func getTermSize() (int, int, error) {
	return 80, 24, nil
}
