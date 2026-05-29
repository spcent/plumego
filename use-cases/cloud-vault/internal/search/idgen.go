package search

import "cloud-vault/internal/idgen"

func newID() string { return idgen.New() }
