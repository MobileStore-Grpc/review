package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	mobilePb "github.com/MobileStore-Grpc/product/pb"
	mobileSample "github.com/MobileStore-Grpc/product/sample"
	"github.com/MobileStore-Grpc/review/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func rateLaptop(reviewClient pb.ReviewServiceClient, mobileIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	stream, err := reviewClient.ReviewMobile(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate mobile: %v", err)
	}
	waitResponse := make(chan error)

	//go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("No more responses")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}

			log.Print("receive response: ", res)
			waitResponse <- nil
		}
	}()

	//send requests
	for i, mobileID := range mobileIDs {
		req := &pb.RateMobileRequest{
			MobileId: mobileID,
			Score:    scores[i],
		}

		// time.Sleep(1 * time.Second)
		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}
		log.Print("sent request:", req)
		err = <-waitResponse
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitResponse
	return err
}

func testReviewMobile(reviewClient pb.ReviewServiceClient, mobileClient mobilePb.MobileServiceClient) {
	n := 3
	mobileIDs := make([]string, n)

	for i := 0; i < n; i++ {
		mobile := mobileSample.NewMobile()
		mobileIDs[i] = mobile.GetId()

		req := &mobilePb.CreateMobileRequest{
			Mobile: mobile,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, err := mobileClient.CreateMobile(ctx, req)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.AlreadyExists {
				log.Print("laptop already exists")
			} else {
				log.Fatal("cannot create laptop: ", err)
			}
		}
		log.Printf("create laptop with id: %s", res.Id)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = mobileSample.RandomLaptopScore()
		}

		err := rateLaptop(reviewClient, mobileIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dail server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial review service: ", err)
	}

	reviewClient := pb.NewReviewServiceClient(conn)
	mobileClient, err := Dial()
	if err != nil {
		log.Fatal("cannot dial mobile search service")
	}

	testReviewMobile(reviewClient, mobileClient)
}

func Dial() (mobilePb.MobileServiceClient, error) {
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	mobileClient := mobilePb.NewMobileServiceClient(conn)
	return mobileClient, nil
}
