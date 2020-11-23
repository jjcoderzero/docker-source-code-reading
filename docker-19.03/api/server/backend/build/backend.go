package build // import "docker-19.03/api/server/backend/build"

import (
	"context"
	"fmt"

	"docker-19.03/api/types"
	"docker-19.03/api/types/backend"
	"docker-19.03/builder"
	buildkit "docker-19.03/builder/builder-next"
	"docker-19.03/builder/fscache"
	"docker-19.03/image"
	"docker-19.03/pkg/stringid"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// ImageComponent提供了一个处理镜像的接口
type ImageComponent interface {
	SquashImage(from string, to string) (string, error)
	TagImageWithReference(image.ID, reference.Named) error
}

// Builder defines interface for running a build
type Builder interface {
	Build(context.Context, backend.BuildConfig) (*builder.Result, error)
}

// Backend provides build functionality to the API router
type Backend struct {
	builder        Builder
	fsCache        *fscache.FSCache
	imageComponent ImageComponent
	buildkit       *buildkit.Builder
}

// NewBackend creates a new build backend from components
func NewBackend(components ImageComponent, builder Builder, fsCache *fscache.FSCache, buildkit *buildkit.Builder) (*Backend, error) {
	return &Backend{imageComponent: components, builder: builder, fsCache: fsCache, buildkit: buildkit}, nil
}

// RegisterGRPC registers buildkit controller to the grpc server.
func (b *Backend) RegisterGRPC(s *grpc.Server) {
	if b.buildkit != nil {
		b.buildkit.RegisterGRPC(s)
	}
}

// Build builds an image from a Source
func (b *Backend) Build(ctx context.Context, config backend.BuildConfig) (string, error) {
	options := config.Options
	useBuildKit := options.Version == types.BuilderBuildKit

	tagger, err := NewTagger(b.imageComponent, config.ProgressWriter.StdoutFormatter, options.Tags)
	if err != nil {
		return "", err
	}

	var build *builder.Result
	if useBuildKit {
		build, err = b.buildkit.Build(ctx, config)
		if err != nil {
			return "", err
		}
	} else {
		build, err = b.builder.Build(ctx, config)
		if err != nil {
			return "", err
		}
	}

	if build == nil {
		return "", nil
	}

	var imageID = build.ImageID
	if options.Squash {
		if imageID, err = squashBuild(build, b.imageComponent); err != nil {
			return "", err
		}
		if config.ProgressWriter.AuxFormatter != nil {
			if err = config.ProgressWriter.AuxFormatter.Emit("moby.image.id", types.BuildResult{ID: imageID}); err != nil {
				return "", err
			}
		}
	}

	if !useBuildKit {
		stdout := config.ProgressWriter.StdoutFormatter
		fmt.Fprintf(stdout, "Successfully built %s\n", stringid.TruncateID(imageID))
	}
	if imageID != "" {
		err = tagger.TagImages(image.ID(imageID))
	}
	return imageID, err
}

// PruneCache removes all cached build sources
func (b *Backend) PruneCache(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error) {
	eg, ctx := errgroup.WithContext(ctx)

	var fsCacheSize uint64
	eg.Go(func() error {
		var err error
		fsCacheSize, err = b.fsCache.Prune(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to prune fscache")
		}
		return nil
	})

	var buildCacheSize int64
	var cacheIDs []string
	eg.Go(func() error {
		var err error
		buildCacheSize, cacheIDs, err = b.buildkit.Prune(ctx, opts)
		if err != nil {
			return errors.Wrap(err, "failed to prune build cache")
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &types.BuildCachePruneReport{SpaceReclaimed: fsCacheSize + uint64(buildCacheSize), CachesDeleted: cacheIDs}, nil
}

// Cancel cancels the build by ID
func (b *Backend) Cancel(ctx context.Context, id string) error {
	return b.buildkit.Cancel(ctx, id)
}

func squashBuild(build *builder.Result, imageComponent ImageComponent) (string, error) {
	var fromID string
	if build.FromImage != nil {
		fromID = build.FromImage.ImageID()
	}
	imageID, err := imageComponent.SquashImage(build.ImageID, fromID)
	if err != nil {
		return "", errors.Wrap(err, "error squashing image")
	}
	return imageID, nil
}