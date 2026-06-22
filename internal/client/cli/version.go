package cli

import (
	"fmt"

	"goph-keeper/internal/buildinfo"
)

func runVersion() int {
	fmt.Printf("gophkeeper %s (built %s)\n", buildinfo.Version, buildinfo.BuildDate)
	return 0
}
