package version

import (
	"fmt"

	pkgversion "github.com/iVampireSP/go-template/pkg/version"
	"github.com/spf13/cobra"
)

type Version struct{}

func NewVersion() *Version {
	return &Version{}
}

func (v *Version) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
	}
}

func (v *Version) Handle(_ *cobra.Command) error {
	fmt.Println(pkgversion.String())
	return nil
}
