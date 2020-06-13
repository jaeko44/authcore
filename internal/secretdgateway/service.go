package secretdgateway

import (
	"context"
	"encoding/base64"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	pb "authcore.io/authcore/pkg/api/secretdgateway"
	"authcore.io/authcore/pkg/secret"

	"github.com/flynn/noise"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/blocksq/secretd-client-go"
	"golang.org/x/crypto/curve25519"
)

// Service is a service that forward requests to secretd.
type Service struct {
	auditStore *audit.Store
}

// NewService initialize a new Service.
func NewService(auditStore *audit.Store) (*Service, error) {
	return &Service{
		auditStore: auditStore,
	}, nil
}

// Forward sends a request to secretd. Secretd will directly authenticates the request and the
// messages are end-to-end encrypted.
func (s *Service) Forward(ctx context.Context, in *pb.ForwardRequest) (*pb.ForwardResponse, error) {
	user, ok := user.CurrentUserFromContext(ctx)
	if !ok || user == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	address := viper.GetString("secretd_address")
	privateKey := viper.Get("secrets.secretd_client_private_key").(secret.String).SecretString()

	authProvider, err := newStaticKeyAuthProvider(privateKey)
	if err != nil {
		return nil, err
	}

	clusterIdentity, err := s.getClusterIdentity()
	if err != nil {
		return nil, err
	}

	client, err := secretd.Dial(address, clusterIdentity, authProvider)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	params := [][]byte{in.GetRequestMessage()}
	var responseMessage []byte
	err = client.Call("system_subrequest", params, &responseMessage)
	if err != nil {
		log.Errorf("error occurred when calling secretd %v", err)
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	response := pb.ForwardResponse{
		ResponseMessage: responseMessage,
	}
	return &response, nil
}

// GetInfo returns information about a secretd cluster.
func (s *Service) GetInfo(ctx context.Context, in *pb.GetInfoRequest) (*pb.Info, error) {
	clusterIdentity, err := s.getClusterIdentity()
	if err != nil {
		return nil, err
	}
	info := pb.Info{
		ClusterIdentity: base64.StdEncoding.EncodeToString(clusterIdentity),
	}
	return &info, nil
}

func (s *Service) getClusterIdentity() ([]byte, error) {
	identityString := viper.GetString("secretd_cluster_identity")
	if identityString == "" {
		log.Error("missing secretd_cluster_identity config")
		return nil, errors.New(errors.ErrorUnknown, "")
	}
	identity, err := base64.StdEncoding.DecodeString(identityString)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return identity, nil
}

// staticKeyAuthProvider implements secretd static key authentication
type staticKeyAuthProvider struct {
	staticPrivateKey [32]byte
	staticPublicKey  [32]byte
}

// newStaticKeyAuthProvider creates a new static key auth provider with the given base64-formated
// local static private key.
func newStaticKeyAuthProvider(staticPrivateKey string) (*staticKeyAuthProvider, error) {
	privateKeySlice, err := base64.StdEncoding.DecodeString(staticPrivateKey)
	if err != nil {
		log.Error("cannot decode static private key")
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if len(privateKeySlice) != 32 {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid private key length")
	}
	var privateKey, publicKey [32]byte
	copy(privateKey[:], privateKeySlice[:32])

	curve25519.ScalarBaseMult(&publicKey, &privateKey)
	return &staticKeyAuthProvider{
		staticPrivateKey: privateKey,
		staticPublicKey:  publicKey,
	}, nil
}

// Name returns the name of the AuthProvider
func (p *staticKeyAuthProvider) Name() string {
	return "static_key"
}

// AuthParams returns the authentication parameters
func (p *staticKeyAuthProvider) AuthParams() interface{} {
	return nil
}

// LocalStaticKeyPair returns local
func (p *staticKeyAuthProvider) LocalStaticKeypair() noise.DHKey {
	return noise.DHKey{
		Private: p.staticPrivateKey[:],
		Public:  p.staticPublicKey[:],
	}
}
