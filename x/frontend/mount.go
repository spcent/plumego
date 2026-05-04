package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/router"
)

// Mount describes an explicitly constructed frontend mount.
// It separates bundle construction from router registration.
type Mount struct {
	prefix  string
	handler http.Handler
}

type routeSpec struct {
	method string
	path   string
}

// Registrar is the minimal router contract required to mount a frontend
// handler. `router.Router` satisfies this interface.
type Registrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

type routeSnapshotter interface {
	Routes() []router.RouteInfo
}

// RegisterFromDir mounts a built frontend directory (e.g. Next.js `out/`)
// at the given prefix. Returns an error if the directory is missing or unreadable.
func RegisterFromDir(r Registrar, dir string, opts ...Option) error {
	mount, err := NewMountFromDir(dir, opts...)
	if err != nil {
		return err
	}
	return mount.Register(r)
}

// NewMountFromDir constructs a frontend mount from a filesystem directory.
func NewMountFromDir(dir string, opts ...Option) (*Mount, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("frontend directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("frontend path %q is not a directory", dir)
	}
	if _, err := os.ReadDir(dir); err != nil {
		return nil, fmt.Errorf("frontend directory %q not readable: %w", dir, err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("frontend directory %q absolute path: %w", dir, err)
	}
	root, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		return nil, fmt.Errorf("frontend directory %q real path: %w", dir, err)
	}
	fsys := localDirFS(root)
	if err := validateDirectoryIndex(fsys, cfg.IndexFile); err != nil {
		return nil, fmt.Errorf("frontend directory %q index %q: %w", dir, cfg.IndexFile, err)
	}
	h, err := newHandlerFS(fsys, cfg)
	if err != nil {
		return nil, err
	}
	return &Mount{
		prefix:  cfg.Prefix,
		handler: h,
	}, nil
}

// RegisterFS mounts a frontend bundle served from the provided http.FileSystem.
// This is suitable for go:embed bundles using http.FS.
func RegisterFS(r Registrar, fsys http.FileSystem, opts ...Option) error {
	mount, err := NewMountFS(fsys, opts...)
	if err != nil {
		return err
	}
	return mount.Register(r)
}

// NewMountFS constructs a frontend mount from an http.FileSystem.
func NewMountFS(fsys http.FileSystem, opts ...Option) (*Mount, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	h, err := newHandlerFS(fsys, cfg)
	if err != nil {
		return nil, err
	}
	return &Mount{
		prefix:  cfg.Prefix,
		handler: h,
	}, nil
}

// NewHandlerFS constructs a frontend handler without registering routes.
func NewHandlerFS(fsys http.FileSystem, opts ...Option) (http.Handler, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	return newHandlerFS(fsys, cfg)
}

func newHandlerFS(fsys http.FileSystem, cfg *config) (http.Handler, error) {
	if fsys == nil {
		return nil, errors.New("filesystem cannot be nil")
	}
	return &handler{
		cfg:      *cfg,
		fs:       fsys,
		variants: newPrecompressedVariantPlan(fsys, cfg.EnablePrecompressed),
	}, nil
}

// Prefix returns the normalized mount prefix.
func (m *Mount) Prefix() string {
	if m == nil {
		return ""
	}
	return m.prefix
}

// Handler returns the mounted frontend handler.
func (m *Mount) Handler() http.Handler {
	if m == nil {
		return nil
	}
	return m.handler
}

// Register attaches the mount to the provided router.
func (m *Mount) Register(r Registrar) error {
	if m == nil {
		return errors.New("mount cannot be nil")
	}
	if r == nil {
		return errors.New("router cannot be nil")
	}
	if m.handler == nil {
		return errors.New("mount handler cannot be nil")
	}

	plan := m.routePlan()
	if err := preflightMountRoutes(r, plan); err != nil {
		return err
	}
	for _, route := range plan {
		if err := r.AddRoute(route.method, route.path, m.handler); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mount) routePlan() []routeSpec {
	if m.prefix == "/" {
		return []routeSpec{
			{method: methodAny, path: "/"},
			{method: methodAny, path: "/*filepath"},
		}
	}

	return []routeSpec{
		{method: methodAny, path: m.prefix + "/*filepath"},
		{method: methodAny, path: m.prefix},
	}
}

func preflightMountRoutes(r Registrar, plan []routeSpec) error {
	snapshotter, ok := r.(routeSnapshotter)
	if !ok {
		return nil
	}
	existing := make(map[routeSpec]struct{})
	for _, route := range snapshotter.Routes() {
		existing[routeSpec{method: route.Method, path: route.Path}] = struct{}{}
	}
	for _, route := range plan {
		if _, ok := existing[route]; ok {
			return fmt.Errorf("frontend mount route %s %s already registered", route.method, route.path)
		}
	}
	return nil
}
