package skills

import (
	"clawrt/internal/sys"
)

func ExecuteTypedUCISet(pkg, path, val string) (string, error) {
	return sys.ExecuteTypedUCISet(pkg, path, val)
}

func ExecuteTypedServiceRestart(service string) (string, error) {
	return sys.ExecuteTypedServiceRestart(service)
}
