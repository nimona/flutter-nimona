package provider

import (
	"database/sql"
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
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/version"
)

type (
	// Provider interface {
	// 	Subscribe(context.Context, ...string) (object.ReadCloser, error)
	// 	Put(context.Context, *object.Object) (*object.Object, error)
	// 	RequestStream(context.Context, object.Hash) (object.ReadCloser, error)
	// }
	Provider struct {
		local         localpeer.LocalPeer
		network       network.Network
		resolver      resolver.Resolver
		objectstore   objectstore.Store
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

	logger.Info("ready")

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(nConfig.Path, "chat.db"))
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

// payload should start with one of the following:
// - type:<type>
// - hash:<hash>
// - stream:<rootHash>
// - owner:<publicKey>
func (p *Provider) Subscribe(
	ctx context.Context,
	lookups ...string,
) (object.ReadCloser, error) {
	opts := []objectmanager.LookupOption{}
	for _, l := range lookups {
		parts := strings.Split(l, ":")
		if len(parts) < 2 {
			continue
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
	}
	return p.objectmanager.Subscribe(), nil
}

func (p *Provider) RequestStream(
	ctx context.Context,
	rootHash object.Hash,
) (object.ReadCloser, error) {
	recipients, err := p.resolver.Lookup(
		ctx,
		resolver.LookupByContentHash(rootHash),
	)
	if err != nil {
		return nil, err
	}
	return p.objectmanager.RequestStream(ctx, rootHash, recipients...)
}

func (p *Provider) Put(
	ctx context.Context,
	obj *object.Object,
) (*object.Object, error) {
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
