package server

import (
	"context"
	"net"
	"net/http"
	"sync"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/authn"
	"authcore.io/authcore/internal/authn/idp"
	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	httpServer "authcore.io/authcore/internal/http"
	"authcore.io/authcore/internal/oauth"
	"authcore.io/authcore/internal/rbac"
	"authcore.io/authcore/internal/secretdgateway"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/settings"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/user/registration"
	"authcore.io/authcore/internal/v1/authapi"
	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/v1/managementapi"
	authapipb "authcore.io/authcore/pkg/api/authapi"
	managementapipb "authcore.io/authcore/pkg/api/managementapi"
	secretdgatewaypb "authcore.io/authcore/pkg/api/secretdgateway"
	"authcore.io/authcore/pkg/grpcinterceptor/xrequestid"
	"authcore.io/authcore/pkg/messageencryptor"
	"authcore.io/authcore/pkg/secret"

	"github.com/go-redis/redis"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	buildVersion string = "develop"
)

// Server represents the global context of Authcore server.
type Server struct {
	rpc              *grpc.Server
	httpHandler      http.Handler
	rpcLock          sync.Mutex
	db               *db.DB
	redis            *redis.Client
	keyGenerator     *messageencryptor.KeyGenerator
	messageEncryptor *messageencryptor.MessageEncryptor

	enforcer              *rbac.Enforcer
	templateStore         *template.Store
	emailService          *email.Service
	smsService            *sms.Service
	userStore             *user.Store
	sessionStore          *session.Store
	authenticationService *authentication.Service
	auditStore            *audit.Store
	authService           *authapi.Service
	rbacService           *rbac.Service
	managementService     *managementapi.Service
	secretdGatewayService *secretdgateway.Service
	authnStore            *authn.Store
	authnTC               *authn.TransactionController

	http *httpServer.Server
}

// NewServer create new instance of the API server.
// Server requires initialization and manuel start.
func NewServer() *Server {
	s := &Server{}

	s.initConfig()
	s.initLogging()
	s.initDB()
	s.initRedis()
	s.initService()

	return s
}

// BuildVersion returns the version set at build time.
func (s *Server) BuildVersion() string {
	return buildVersion
}

// Start the gRPC server
func (s *Server) Start() {
	log.Infof("Authcore version: %v", s.BuildVersion())
	if viper.GetBool("apiv1_enabled") {
		log.Info("API v1 enabled")
		go s.startGRPCServer()
	}
	s.http.Start()
}

// Stop stops the gRPC server. It immediately closes all open connections and listeners.
func (s *Server) Stop() {
	s.rpcLock.Lock()
	if s.rpc != nil {
		s.rpc.Stop()
	}
	s.rpcLock.Unlock()
}

// GracefulStop stops the gRPC server gracefully.
func (s *Server) GracefulStop() {
	s.rpcLock.Lock()
	if s.rpc != nil {
		s.rpc.GracefulStop()
	}
	s.rpcLock.Unlock()
}

// CreateFirstAdminUser creates the first admin user for a deployment. If a user is already registered, this method
// returns an error. This method is intented to be called by CLI .
func (s *Server) CreateFirstAdminUser(ctx context.Context, user *user.User, password string) (*user.User, error) {
	return s.managementService.CreateFirstAdminUser(ctx, user, password)
}

func (s *Server) initConfig() {
	config.InitDefaults()
	config.InitConfig()
	config.PrintConfig()
}

func (s *Server) initLogging() {
	// Logrus setting
	// Default setting, Set to JSONFormatter if necessary
	log.SetFormatter(&log.TextFormatter{})
}

func (s *Server) initDB() {
	s.db = db.NewDBFromConfig()
}

func (s *Server) initRedis() {
	s.redis = NewRedisClientFromConfig()
}

func (s *Server) initService() {
	s.initEncryptor()
	s.templateStore = template.NewStore(s.db)
	s.emailService = email.NewService(s.templateStore)
	s.smsService = sms.NewService(s.templateStore)
	s.userStore = user.NewStore(s.db, s.redis, s.messageEncryptor)
	s.sessionStore = session.NewStore(s.db, s.redis, s.userStore)
	s.authenticationService = authentication.NewService(s.redis, s.userStore)
	s.auditStore = audit.NewStore(s.db)
	s.authService = authapi.NewService(s.db, s.redis, s.userStore, s.sessionStore, s.authenticationService, s.auditStore, s.emailService, s.smsService)
	roleResolver := rbac.NewUserRoleResolver(s.userStore)
	s.rbacService = rbac.NewService(roleResolver, managementapi.PermissionAssignments)
	s.managementService = managementapi.NewService(s.db, s.userStore, s.sessionStore, s.authenticationService, s.auditStore, s.rbacService, s.templateStore, s.emailService, s.smsService)
	s.secretdGatewayService, _ = secretdgateway.NewService(s.auditStore)
	s.authnStore = authn.NewStore(s.redis, s.messageEncryptor)
	s.enforcer = rbac.NewEnforcer(s.userStore, s.sessionStore)
	s.initAuthnTC()
	s.initHTTPServer()
	s.initGRPCServer()
}

func (s *Server) initEncryptor() {
	secret, err := viper.Get("secret_key_base").(secret.String).SecretBytes()
	if err != nil {
		log.Fatalf("cannot get secret: %v", err)
	}
	if len(secret) < 32 {
		log.Fatalf("secret_key_base must be a hex string of at least 32 bytes")
	}
	s.keyGenerator = messageencryptor.NewKeyGenerator(secret)
	s.messageEncryptor, err = messageencryptor.NewMessageEncryptor(
		s.keyGenerator.Derive(
			"FieldEncryptor/Xsalsa20Poly1305",
			messageencryptor.CipherXsalsa20Poly1305.KeyLength(),
		),
		messageencryptor.CipherXsalsa20Poly1305,
	)
	if err != nil {
		log.Fatalf("cannot initialize message encryptor: %v", err)
	}
}

func (s *Server) initAuthnTC() {
	tc := authn.NewTransactionController(s.db, s.authnStore, s.userStore, s.sessionStore)
	tc.RegisterVerifier(verifier.SMSOTP, verifier.SMSOTPVerifierFactory(s.smsService, s.redis))
	tc.RegisterVerifier(verifier.ResetLink, verifier.ResetLinkVerifierFactory(s.smsService, s.emailService, s.redis))
	if viper.IsSet("google_app_id") {
		tc.RegisterIDP(idp.NewGoogleIDP())
	}
	if viper.IsSet("facebook_app_id") {
		tc.RegisterIDP(idp.NewFacebookIDP())
	}
	if viper.IsSet("twitter_consumer_key") {
		tc.RegisterIDP(idp.NewTwitterIDP())
	}
	if viper.IsSet("apple_app_id") {
		tc.RegisterIDP(idp.NewAppleIDP())
	}
	if viper.IsSet("matters_app_id") {
		tc.RegisterIDP(idp.NewMattersIDP())
	}
	s.authnTC = tc
}

func (s *Server) initGRPCServer() {

	logrusEntry := log.NewEntry(log.StandardLogger())
	interceptors := []grpc.UnaryServerInterceptor{
		HTTPAddressUnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		xrequestid.UnaryServerInterceptor(),
		grpc_opentracing.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrusEntry),
		ErrorLoggingUnaryServerInterceptor(), // should be placed below grpc_logrus.UnaryServerInterceptor to log stack trace upon errors.
		NewAuthorizationUnaryInterceptor(
			s.sessionStore.VerifyAccessToken,
			s.userStore.UserByPublicID,
			s.sessionStore.FindSessionByPublicID,
		),
	}
	chain := grpc_middleware.ChainUnaryServer(interceptors...)

	s.rpcLock.Lock()
	s.rpc = grpc.NewServer(grpc.UnaryInterceptor(chain))
	authapipb.RegisterAuthServiceServer(s.rpc, s.authService)
	managementapipb.RegisterManagementServiceServer(s.rpc, s.managementService)
	secretdgatewaypb.RegisterSecretdGatewayServer(s.rpc, s.secretdGatewayService)
	reflection.Register(s.rpc)
	s.rpcLock.Unlock()
}

func (s *Server) initHTTPServer() {
	s.http = httpServer.NewServer(
		session.UserAgentMiddleware(nil),
		session.AccessTokenAuthMiddleware(nil, s.sessionStore),
		rbac.EnforcerMiddleware(nil, s.enforcer),
	)

	if viper.GetBool("apiv1_enabled") {
		s.http.GRPCGateway("/api/auth", authapipb.RegisterAuthServiceHandler)
		s.http.GRPCGateway("/api/management", managementapipb.RegisterManagementServiceHandler)
	}
	if viper.GetBool("secretdgateway_enabled") {
		log.Info("secretdgateway enabled")
		s.http.GRPCGateway("/api/secretdgateway", secretdgatewaypb.RegisterSecretdGatewayHandler)
	}
	s.http.Register(oauth.API(s.userStore, s.sessionStore, s.authnTC))
	s.http.Register(authn.APIv2(s.authnTC, s.auditStore))
	s.http.Register(audit.APIv2(s.auditStore))
	s.http.Register(user.APIv2(s.userStore))
	s.http.Register(registration.APIv2(s.userStore, s.sessionStore, s.emailService, s.smsService))
	s.http.Register(template.APIv2(s.templateStore))
	s.http.Register(session.APIv2(s.sessionStore))
	s.http.Register(settings.APIv2())
}

func (s *Server) startGRPCServer() {
	grpcListen := viper.Get("grpc_listen").(string)
	listen, err := net.Listen("tcp", grpcListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("GRPC server is listening at", grpcListen)

	if err := s.rpc.Serve(listen); err != nil {
		log.Fatalf("failed to start GRPC server: %v", err)
	}
}
