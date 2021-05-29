package service

import (
	"context"
	"io"
	"log"

	mobilePb "github.com/MobileStore-Grpc/product/pb"
	"github.com/MobileStore-Grpc/review/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//ReviewService provides mobile review services
type ReviewService struct {
	reviewStore ReviewStore
	pb.UnimplementedReviewServiceServer
}

//NewReviewService returns a new review service
func NewReviewService(reviewStore ReviewStore) *ReviewService {
	return &ReviewService{
		reviewStore: reviewStore,
	}
}

// ReviewMobile is a bidirectional-streaming RPC that allows client to rate a stream of mobiles with score,
// and returns a stream of average score for each of them.
func (server *ReviewService) ReviewMobile(stream pb.ReviewService_ReviewMobileServer) error {
	mobileClient, err := Dial()
	log.Print("dail connection to mobile service successfully")
	if err != nil {
		log.Println("cannot dial mobile search service")
		return err
	}
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive stream request: %v", err))
		}

		mobileID := req.GetMobileId()
		score := req.GetScore()
		log.Printf("receive a review-mobile request: id = %s, score = %.2f", mobileID, score)

		//checking whether mobile exist or not before rating
		err = find(mobileID, mobileClient, stream.Context())
		if err != nil {
			return err
		}

		//Review mobile
		review, err := server.reviewStore.Add(mobileID, score)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot add review to the store: %v", err))
		}

		res := &pb.RateMobileResponse{
			MobileId:     mobileID,
			RatedCount:   review.Count,
			AverageScore: review.Sum / float64(review.Count),
		}
		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot send stream response: %v", err))
		}
	}
	return nil
}

func find(mobileID string, mobileClient mobilePb.MobileServiceClient, ctx context.Context) error {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	req := mobilePb.SearchMobileRequest{
		MobileId: mobileID,
	}
	_, err := mobileClient.SearchMobile(ctx, &req)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	log.Printf("mobile with mobileID %s found", mobileID)
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
