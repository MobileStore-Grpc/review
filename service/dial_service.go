package service

import (
	mobilePb "github.com/MobileStore-Grpc/product/pb"
	"github.com/MobileStore-Grpc/review/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// var MobileServer *string

func Dial() (mobilePb.MobileServiceClient, error) {
	// conn, err := grpc.Dial(*MobileServer, grpc.WithInsecure())
	config, err := util.LoadConfig("/etc")
	conn, err := grpc.Dial(config.MobileServerAddress, grpc.WithInsecure())

	if err != nil {
		return nil, logError(status.Errorf(codes.Internal, "cannot dial mobile search service", err))
	}
	mobileClient := mobilePb.NewMobileServiceClient(conn)
	return mobileClient, nil
}
