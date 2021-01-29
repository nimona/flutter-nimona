package provider

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"

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

	cConfig := &Config{}
	nConfig, err := config.New(
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

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(nConfig.Path, "nimona.db"))
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

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
		"stream:poc.nimona.io/conversation",          // new(ConversationStreamRoot).Type(),
		"poc.nimona.io/conversation.NicknameUpdated", // new(ConversationMessageAdded).Type(),
		"poc.nimona.io/conversation.MessageAdded",    // new(ConversationNicknameUpdated).Type(),
		"nimona.io/stream.Subscription",              // new(stream.Subscription).Type(),
	)

	// conversationRootObject := conversationRoot.ToObject()
	// conversationRootHash := conversationRootObject.Hash()

	// // register conversation in object manager
	// if _, err := man.Put(ctx, conversationRootObject); err != nil {
	// 	logger.Fatal("could not persist conversation root", log.Error(err))
	// }

	// // add conversation to the list of content we provide
	// local.PutContentHashes(conversationRootHash)

	r, err := str.Filter(
		sqlobjectstore.FilterByObjectType("stream:poc.nimona.io/conversation"),
	)
	if err == nil {
		for {
			o, err := r.Read()
			if err != nil || o == nil {
				break
			}
			local.PutContentHashes(o.Hash())
		}
	}

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
	Lookup   string `json:"lookup"`
	OrderBy  string `json:"orderBy"`
	OrderDir string `json:"orderDir"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type GetResponse struct {
	ObjectBodies []string `json:"objectBodies"`
}

func (p *Provider) Get(
	ctx context.Context,
	req GetRequest,
) (object.ReadCloser, error) {
	opts := []sqlobjectstore.FilterOption{}
	parts := strings.Split(req.Lookup, ":")
	if len(parts) < 2 {
		return nil, errors.New("invalid lookup query")
	}
	prefix := parts[0]
	value := strings.Join(parts[1:], ":")
	switch prefix {
	case "type":
		opts = append(
			opts,
			sqlobjectstore.FilterByObjectType(value),
		)
	case "hash":
		opts = append(
			opts,
			sqlobjectstore.FilterByHash(object.Hash(value)),
		)
	case "owner":
		opts = append(
			opts,
			sqlobjectstore.FilterByOwner(crypto.PublicKey(value)),
		)
	case "stream":
		opts = append(
			opts,
			sqlobjectstore.FilterByStreamHash(object.Hash(value)),
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
	return p.objectstore.Filter(opts...)
}

// payload should start with one of the following:
// - type:<type>
// - hash:<hash>
// - stream:<rootHash>
// - owner:<publicKey>
func (p *Provider) Subscribe(
	ctx context.Context,
	lookup string,
) (object.ReadCloser, error) {
	opts := []objectmanager.LookupOption{}
	parts := strings.Split(lookup, ":")
	if len(parts) < 2 {
		return nil, errors.New("invalid lookup query")
	}
	prefix := parts[0]
	value := strings.Join(parts[1:], ":")
	switch prefix {
	case "type":
		opts = append(
			opts,
			objectmanager.FilterByObjectType(value),
		)
	case "hash":
		opts = append(
			opts,
			objectmanager.FilterByHash(object.Hash(value)),
		)
	case "owner":
		opts = append(
			opts,
			objectmanager.FilterByOwner(crypto.PublicKey(value)),
		)
	case "stream":
		opts = append(
			opts,
			objectmanager.FilterByStreamHash(object.Hash(value)),
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
			r, err := p.objectmanager.RequestStream(ctx, rootHash, recipient)
			if err != nil {
				return
			}
			r.Close()
		}(recipient)
	}
	return nil
}

func (p *Provider) Put(
	ctx context.Context,
	obj *object.Object,
) (*object.Object, error) {
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
