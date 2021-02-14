package provider

import (
	"database/sql"
	"errors"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/version"
)

type (
	Provider struct {
		local         localpeer.LocalPeer
		network       network.Network
		resolver      resolver.Resolver
		objectstore   *sqlobjectstore.Store
		objectmanager objectmanager.ObjectManager
		logger        log.Logger
	}
	Config struct{}
)

func New() *Provider {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	currentUser, _ := user.Current()
	cConfig := &Config{}
	nConfig, err := config.New(
		config.WithDefaultPath(
			filepath.Join(currentUser.HomeDir, ".mochi"),
		),
		config.WithExtraConfig("CHAT", cConfig),
		config.WithDefaultListenOnLocalIPs(),
		config.WithDefaultListenOnPrivateIPs(),
		config.WithDefaultListenOnExternalPort(),
	)
	if err != nil {
		logger.Fatal("error loading config", log.Error(err))
	}

	log.DefaultLogger.SetLogLevel(nConfig.LogLevel)

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(nConfig.Peer.PrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	if nConfig.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
			ctx,
			nConfig.Peer.BindAddress,
			network.ListenOnLocalIPs,
			// network.ListenOnExternalPort,
		)
		if err != nil {
			logger.Fatal("error while listening", log.Error(err))
		}
		defer lis.Close() // nolint: errcheck
	}

	// convert shorthands into connection infos
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range nConfig.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	local.PutRelays(bootstrapPeers...)

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(nConfig.Path, "nimona.db"))
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

	// TODO application specific
	// register all stream roots
	r, err := str.Filter(
		sqlobjectstore.FilterByObjectType("stream:poc.nimona.io/conversation"),
	)
	if err == nil {
		hs := []object.Hash{}
		for {
			o, err := r.Read()
			if err != nil || o == nil {
				break
			}
			hs = append(hs, o.Hash())
		}
		if len(hs) > 0 {
			local.PutContentHashes(hs...)
		}
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	logger = logger.With(
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	logger.Error(
		"ready",
		log.Any("addresses", local.GetAddresses()),
	)

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)

	// TODO: application specifc
	// register types so object manager persists them
	local.PutContentTypes(
		"stream:poc.nimona.io/conversation",
		"poc.nimona.io/conversation.NicknameUpdated",
		"poc.nimona.io/conversation.MessageAdded",
		"poc.nimona.io/conversation.TopicUpdated",
		"nimona.io/stream.Subscription",
	)

	return &Provider{
		local:         local,
		network:       net,
		resolver:      res,
		objectstore:   str,
		objectmanager: man,
		logger:        logger,
	}
}

func (p *Provider) GetConnectionInfo() *peer.ConnectionInfo {
	return p.local.ConnectionInfo()
}

type GetRequest struct {
	Lookups  []string `json:"lookups"`
	OrderBy  string   `json:"orderBy"`
	OrderDir string   `json:"orderDir"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}

type GetResponse struct {
	ObjectBodies []string `json:"objectBodies"`
}

func (p *Provider) Get(
	ctx context.Context,
	req GetRequest,
) (object.ReadCloser, error) {
	opts := []sqlobjectstore.FilterOption{}
	filterByType := []string{}
	filterByHash := []object.Hash{}
	filterByOwner := []crypto.PublicKey{}
	filterByStream := []object.Hash{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		value := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				value,
			)
		case "hash":
			filterByHash = append(
				filterByHash,
				object.Hash(value),
			)
		case "owner":
			filterByOwner = append(
				filterByOwner,
				crypto.PublicKey(value),
			)
		case "stream":
			filterByStream = append(
				filterByStream,
				object.Hash(value),
			)
		}
		if req.OrderBy != "" {
			opts = append(
				opts,
				sqlobjectstore.FilterOrderBy(req.OrderBy),
			)
		}
		if req.OrderDir != "" {
			opts = append(
				opts,
				sqlobjectstore.FilterOrderDir(req.OrderDir),
			)
		}
		if req.Limit > 0 && req.Offset > 0 {
			opts = append(
				opts,
				sqlobjectstore.FilterLimit(req.Limit, req.Offset),
			)
		}
	}
	if len(filterByType) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByObjectType(filterByType...),
		)
	}
	if len(filterByHash) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByHash(filterByHash...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStream) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByStreamHash(filterByStream...),
		)
	}
	return p.objectstore.Filter(opts...)
}

type SubscribeRequest struct {
	Lookups []string `json:"lookups"`
}

// payload should start with one of the following:
// - type:<type>
// - hash:<hash>
// - stream:<rootHash>
// - owner:<publicKey>
func (p *Provider) Subscribe(
	ctx context.Context,
	req SubscribeRequest,
) (object.ReadCloser, error) {
	opts := []objectmanager.LookupOption{}
	filterByType := []string{}
	filterByHash := []object.Hash{}
	filterByOwner := []crypto.PublicKey{}
	filterByStream := []object.Hash{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		value := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				value,
			)
		case "hash":
			filterByHash = append(
				filterByHash,
				object.Hash(value),
			)
		case "owner":
			filterByOwner = append(
				filterByOwner,
				crypto.PublicKey(value),
			)
		case "stream":
			filterByStream = append(
				filterByStream,
				object.Hash(value),
			)
		}
	}
	if len(filterByType) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByObjectType(filterByType...),
		)
	}
	if len(filterByHash) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByHash(filterByHash...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStream) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByStreamHash(filterByStream...),
		)
	}
	reader := p.objectmanager.Subscribe(opts...)
	return reader, nil
}

func (p *Provider) RequestStream(
	ctx context.Context,
	rootHash object.Hash,
) error {
	recipients, err := p.resolver.Lookup(
		ctx,
		resolver.LookupByContentHash(rootHash),
	)
	if err != nil {
		return err
	}
	for _, recipient := range recipients {
		go func(recipient *peer.ConnectionInfo) {
			ctx := context.New(
				context.WithTimeout(10 * time.Second),
			)
			_, err := p.objectmanager.Request(ctx, rootHash, recipient)
			if err != nil {
				return
			}
			r, err := p.objectmanager.RequestStream(ctx, rootHash, recipient)
			if err != nil {
				return
			}
			object.ReadAll(r)
			r.Close()
		}(recipient)
	}
	return nil
}

func (p *Provider) Put(
	ctx context.Context,
	obj *object.Object,
) (*object.Object, error) {
	obj = object.Copy(obj)
	switch obj.Metadata.Owner {
	case "@peer":
		obj.Metadata.Owner = p.local.GetPrimaryPeerKey().PublicKey()
	case "@identity":
		obj.Metadata.Owner = p.local.GetPrimaryIdentityKey().PublicKey()
	}
	return p.objectmanager.Put(ctx, obj)
}

func (p *Provider) GetFeedRootHash(
	streamRootObjectType string,
) object.Hash {
	v := &feed.FeedStreamRoot{
		ObjectType: streamRootObjectType,
		Metadata: object.Metadata{
			Owner: p.local.GetPrimaryPeerKey().PublicKey(),
		},
	}
	return v.ToObject().Hash()
}
