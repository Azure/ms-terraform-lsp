package tf

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Terraform struct {
	exec             *tfexec.Terraform
	LogEnabled       bool
	workingDirectory string
}

// FindTerraform finds the path to the terraform executable.
func FindTerraform(ctx context.Context) (string, error) {
	i := install.NewInstaller()
	return i.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: version.MustConstraints(version.NewConstraint(">=0.12")),
		},
	})
}

func NewTerraform(workingDirectory string, logEnabled bool) (*Terraform, error) {
	execPath, err := FindTerraform(context.Background())
	if err != nil {
		return nil, err
	}
	tf, err := tfexec.NewTerraform(workingDirectory, execPath)
	if err != nil {
		return nil, err
	}

	t := &Terraform{
		exec:             tf,
		workingDirectory: workingDirectory,
		LogEnabled:       logEnabled,
	}
	t.SetLogEnabled(true)
	return t, nil
}

func (t *Terraform) SetLogEnabled(enabled bool) {
	if enabled && t.LogEnabled {
		t.exec.SetStdout(os.Stdout)
		t.exec.SetStderr(os.Stderr)
		t.exec.SetLogger(log.New(os.Stdout, "", 0))
	} else {
		t.exec.SetStdout(io.Discard)
		t.exec.SetStderr(io.Discard)
		t.exec.SetLogger(log.New(io.Discard, "", 0))
	}
}

func (t *Terraform) GetExec() *tfexec.Terraform {
	return t.exec
}

func (t *Terraform) GetWorkingDirectory() string {
	return t.workingDirectory
}
