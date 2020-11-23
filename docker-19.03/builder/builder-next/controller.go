package buildkit

import (
	"net/http"
	"os"
	"path/filepath"

	"docker-19.03/api/types"
	"docker-19.03/api/types/filters"
	"docker-19.03/builder/builder-next/adapters/containerimage"
	"docker-19.03/builder/builder-next/adapters/localinlinecache"
	"docker-19.03/builder/builder-next/adapters/snapshot"
	containerimageexp "docker-19.03/builder/builder-next/exporter"
	"docker-19.03/builder/builder-next/imagerefchecker"
	mobyworker "docker-19.03/builder/builder-next/worker"
	"docker-19.03/buildkit/cache"
	"docker-19.03/buildkit/cache/metadata"
	"docker-19.03/buildkit/cache/remotecache"
	inlineremotecache "docker-19.03/buildkit/cache/remotecache/inline"
	localremotecache "docker-19.03/buildkit/cache/remotecache/local"
	"docker-19.03/buildkit/client"
	"docker-19.03/buildkit/control"
	"docker-19.03/buildkit/frontend"
	dockerfile "docker-19.03/buildkit/frontend/dockerfile/builder"
	"docker-19.03/buildkit/frontend/gateway"
	"docker-19.03/buildkit/frontend/gateway/forwarder"
	"docker-19.03/buildkit/snapshot/blobmapping"
	"docker-19.03/buildkit/solver/bboltcachestorage"
	"docker-19.03/buildkit/util/binfmt_misc"
	"docker-19.03/buildkit/util/entitlements"
	"docker-19.03/buildkit/worker"
	"docker-19.03/daemon/config"
	"docker-19.03/daemon/graphdriver"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/platforms"
	units "github.com/docker/go-units"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func newController(rt http.RoundTripper, opt Opt) (*control.Controller, error) {
	if err := os.MkdirAll(opt.Root, 0711); err != nil {
		return nil, err
	}

	dist := opt.Dist
	root := opt.Root

	var driver graphdriver.Driver
	if ls, ok := dist.LayerStore.(interface {
		Driver() graphdriver.Driver
	}); ok {
		driver = ls.Driver()
	} else {
		return nil, errors.Errorf("could not access graphdriver")
	}

	sbase, err := snapshot.NewSnapshotter(snapshot.Opt{
		GraphDriver:     driver,
		LayerStore:      dist.LayerStore,
		Root:            root,
		IdentityMapping: opt.IdentityMapping,
	})
	if err != nil {
		return nil, err
	}

	store, err := local.NewStore(filepath.Join(root, "content"))
	if err != nil {
		return nil, err
	}
	store = &contentStoreNoLabels{store}

	md, err := metadata.NewStore(filepath.Join(root, "metadata.db"))
	if err != nil {
		return nil, err
	}

	snapshotter := blobmapping.NewSnapshotter(blobmapping.Opt{
		Content:       store,
		Snapshotter:   sbase,
		MetadataStore: md,
	})

	layerGetter, ok := sbase.(imagerefchecker.LayerGetter)
	if !ok {
		return nil, errors.Errorf("snapshotter does not implement layergetter")
	}

	refChecker := imagerefchecker.New(imagerefchecker.Opt{
		ImageStore:  dist.ImageStore,
		LayerGetter: layerGetter,
	})

	cm, err := cache.NewManager(cache.ManagerOpt{
		Snapshotter:     snapshotter,
		MetadataStore:   md,
		PruneRefChecker: refChecker,
	})
	if err != nil {
		return nil, err
	}

	src, err := containerimage.NewSource(containerimage.SourceOpt{
		CacheAccessor:   cm,
		ContentStore:    store,
		DownloadManager: dist.DownloadManager,
		MetadataStore:   dist.V2MetadataService,
		ImageStore:      dist.ImageStore,
		ReferenceStore:  dist.ReferenceStore,
		ResolverOpt:     opt.ResolverOpt,
	})
	if err != nil {
		return nil, err
	}

	dns := getDNSConfig(opt.DNSConfig)

	exec, err := newExecutor(root, opt.DefaultCgroupParent, opt.NetworkController, dns, opt.Rootless, opt.IdentityMapping)
	if err != nil {
		return nil, err
	}

	differ, ok := sbase.(containerimageexp.Differ)
	if !ok {
		return nil, errors.Errorf("snapshotter doesn't support differ")
	}

	exp, err := containerimageexp.New(containerimageexp.Opt{
		ImageStore:     dist.ImageStore,
		ReferenceStore: dist.ReferenceStore,
		Differ:         differ,
	})
	if err != nil {
		return nil, err
	}

	cacheStorage, err := bboltcachestorage.NewStore(filepath.Join(opt.Root, "cache.db"))
	if err != nil {
		return nil, err
	}

	gcPolicy, err := getGCPolicy(opt.BuilderConfig, root)
	if err != nil {
		return nil, errors.Wrap(err, "could not get builder GC policy")
	}

	layers, ok := sbase.(mobyworker.LayerAccess)
	if !ok {
		return nil, errors.Errorf("snapshotter doesn't support differ")
	}

	p, err := parsePlatforms(binfmt_misc.SupportedPlatforms())
	if err != nil {
		return nil, err
	}

	wopt := mobyworker.Opt{
		ID:                "moby",
		MetadataStore:     md,
		ContentStore:      store,
		CacheManager:      cm,
		GCPolicy:          gcPolicy,
		Snapshotter:       snapshotter,
		Executor:          exec,
		ImageSource:       src,
		DownloadManager:   dist.DownloadManager,
		V2MetadataService: dist.V2MetadataService,
		Exporter:          exp,
		Transport:         rt,
		Layers:            layers,
		Platforms:         p,
	}

	wc := &worker.Controller{}
	w, err := mobyworker.NewWorker(wopt)
	if err != nil {
		return nil, err
	}
	wc.Add(w)

	frontends := map[string]frontend.Frontend{
		"dockerfile.v0": forwarder.NewGatewayForwarder(wc, dockerfile.Build),
		"gateway.v0":    gateway.NewGatewayFrontend(wc),
	}

	return control.NewController(control.Opt{
		SessionManager:   opt.SessionManager,
		WorkerController: wc,
		Frontends:        frontends,
		CacheKeyStorage:  cacheStorage,
		ResolveCacheImporterFuncs: map[string]remotecache.ResolveCacheImporterFunc{
			"registry": localinlinecache.ResolveCacheImporterFunc(opt.SessionManager, opt.ResolverOpt, dist.ReferenceStore, dist.ImageStore),
			"local":    localremotecache.ResolveCacheImporterFunc(opt.SessionManager),
		},
		ResolveCacheExporterFuncs: map[string]remotecache.ResolveCacheExporterFunc{
			"inline": inlineremotecache.ResolveCacheExporterFunc(),
		},
		Entitlements: getEntitlements(opt.BuilderConfig),
	})
}

func getGCPolicy(conf config.BuilderConfig, root string) ([]client.PruneInfo, error) {
	var gcPolicy []client.PruneInfo
	if conf.GC.Enabled {
		var (
			defaultKeepStorage int64
			err                error
		)

		if conf.GC.DefaultKeepStorage != "" {
			defaultKeepStorage, err = units.RAMInBytes(conf.GC.DefaultKeepStorage)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse '%s' as Builder.GC.DefaultKeepStorage config", conf.GC.DefaultKeepStorage)
			}
		}

		if conf.GC.Policy == nil {
			gcPolicy = mobyworker.DefaultGCPolicy(root, defaultKeepStorage)
		} else {
			gcPolicy = make([]client.PruneInfo, len(conf.GC.Policy))
			for i, p := range conf.GC.Policy {
				b, err := units.RAMInBytes(p.KeepStorage)
				if err != nil {
					return nil, err
				}
				if b == 0 {
					b = defaultKeepStorage
				}
				gcPolicy[i], err = toBuildkitPruneInfo(types.BuildCachePruneOptions{
					All:         p.All,
					KeepStorage: b,
					Filters:     filters.Args(p.Filter),
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return gcPolicy, nil
}

func parsePlatforms(platformsStr []string) ([]specs.Platform, error) {
	out := make([]specs.Platform, 0, len(platformsStr))
	for _, s := range platformsStr {
		p, err := platforms.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, platforms.Normalize(p))
	}
	return out, nil
}

func getEntitlements(conf config.BuilderConfig) []string {
	var ents []string
	// Incase of no config settings, NetworkHost should be enabled & SecurityInsecure must be disabled.
	if conf.Entitlements.NetworkHost == nil || *conf.Entitlements.NetworkHost {
		ents = append(ents, string(entitlements.EntitlementNetworkHost))
	}
	if conf.Entitlements.SecurityInsecure != nil && *conf.Entitlements.SecurityInsecure {
		ents = append(ents, string(entitlements.EntitlementSecurityInsecure))
	}
	return ents
}
