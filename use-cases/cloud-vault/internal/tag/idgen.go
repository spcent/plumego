package tag

import "cloud-vault/internal/idgen"

func newID() string {
	return idgen.New()
}
