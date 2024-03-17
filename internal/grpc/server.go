package grpcserver

import (
	"context"
	"errors"
	"strconv"
	"time"

	errText "github.com/Dorrrke/GophKeeper-server/internal/domain/errors"
	"github.com/Dorrrke/GophKeeper-server/internal/domain/models"
	"github.com/Dorrrke/GophKeeper-server/internal/service"
	"github.com/Dorrrke/GophKeeper-server/internal/storage"
	gophkeeperv1 "github.com/Dorrrke/goph-keeper-proto/gen/go/gophkeeper"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)

var ErrInvalidToken = errors.New(errText.InvalidTokenError)

const SecretKey = "Secret123Key345Super"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type KeepServer struct {
	gophkeeperv1.UnimplementedGophKeeperServer
	keepService *service.KeepService
	zlog        *zerolog.Logger
}

func RegisterGrpcServer(gRPC *grpc.Server, service *service.KeepService, log *zerolog.Logger) {
	gophkeeperv1.RegisterGophKeeperServer(gRPC, &KeepServer{keepService: service, zlog: log})
}

func (k *KeepServer) SignIn(ctx context.Context, req *gophkeeperv1.SingInRequest) (*gophkeeperv1.SignInResponse, error) {
	user, err := k.keepService.LoginUser(req.GetLogin(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotExist) {
			k.zlog.Error().Err(err).Msg("invalid login/password pair. This user does not exist.")
			return nil, status.Error(codes.Unauthenticated, errText.InvalidLoginUserError)
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			k.zlog.Error().Err(err).Msg("invalid password")
			return nil, status.Error(codes.Unauthenticated, errText.InvalidPasswordError)
		}
		k.zlog.Error().Err(err).Msg("error during user authentication attempt")
		return nil, status.Error(codes.Internal, "internal error")
	}
	jwtToken, err := createJWTToken(user.UserID)
	if err != nil {
		k.zlog.Error().Err(err).Msg("error during JWT token creation")
		return nil, status.Error(codes.Internal, "internal error")
	}
	header := metadata.Pairs("Authorization", jwtToken)
	grpc.SendHeader(ctx, header)
	return &gophkeeperv1.SignInResponse{}, nil
}

func (k *KeepServer) SignUp(ctx context.Context, req *gophkeeperv1.SignUpRequest) (*gophkeeperv1.SignUpResponse, error) {
	uid, err := k.keepService.RegisterUser(req.GetLogin(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserAlredyExist) {
			k.zlog.Error().Err(err).Msg("user alredy exist")
			return nil, status.Error(codes.Canceled, errText.UserExistsError)
		}
		k.zlog.Error().Err(err).Msg("error during user registration attempt")
		return nil, status.Error(codes.Internal, "internal error")
	}
	jwtToken, err := createJWTToken(uid)
	if err != nil {
		k.zlog.Error().Err(err).Msg("error during JWT token creation")
		return nil, status.Error(codes.Internal, "internal error")
	}
	header := metadata.Pairs("Authorization", jwtToken)
	grpc.SendHeader(ctx, header)
	k.zlog.Debug().Str("Token", jwtToken)
	return &gophkeeperv1.SignUpResponse{}, nil
}

func (k *KeepServer) SyncDB(ctx context.Context, req *gophkeeperv1.SyncDBRequest) (*gophkeeperv1.SyncDBResponse, error) {
	mData, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		k.zlog.Error().Msg(errText.MetadataError)
		return nil, status.Error(codes.PermissionDenied, errText.MetadataError)
	}
	values := mData.Get("Authorization")
	if len(values) != 1 {
		k.zlog.Error().Msg(errText.MissingAuthorizationKeyError)
		return nil, status.Error(codes.PermissionDenied, errText.MissingAuthorizationKeyError)
	}
	authToken := values[0]
	userID, err := GetUID(authToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			k.zlog.Error().Msg(err.Error())
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	k.zlog.Debug().Str("userId", userID).Msg("User id from token")
	uID, err := strconv.Atoi(userID)
	if err != nil {
		k.zlog.Error().Err(err).Msg("str to int error")
		return nil, err
	}
	model, err := k.keepService.SyncDB(models.ProtoSyncModel{
		Cards: req.Cards,
		Texts: req.Texts,
		Bins:  req.Bins,
		Auth:  req.Auth,
	}, uID)
	if err != nil {
		return nil, err
	}
	return &gophkeeperv1.SyncDBResponse{
		Auth:  model.Auth,
		Cards: model.Cards,
		Texts: model.Texts,
		Bins:  model.Bins,
	}, nil
}

func createJWTToken(uid int64) (string, error) {
	uuid := strconv.FormatInt(uid, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		},
		UserID: uuid,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUID - функция получения id пользвателя из jwt токена.
func GetUID(tokenString string) (string, error) {
	claim := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	return claim.UserID, nil
}
