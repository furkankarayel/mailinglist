package main

import (
	"context"
	"log"
	pb "mailinglist/proto"
	"time"

	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func logResponse(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if res.EmailEntry == nil {
		log.Printf("No email found")
	} else {
		log.Printf("Email: %v", res.EmailEntry)
	}
}

func createEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("Creating email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

func getEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("Getting email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponse(res, err)
	return res.EmailEntry
}

func getEmailBatch(client pb.MailingListServiceClient, page int32, count int32) *pb.GetEmailBatchResponse {
	log.Println("Getting email batch")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{Page: int32(page), Count: int32(count)})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Println("response:")
	for i := 0; i < len(res.EmailEntries); i++ {
		log.Printf("item [%v of %v]: %s", i+1, len(res.EmailEntries), res.EmailEntries[i])
	}
	return res
}

func updateEmail(client pb.MailingListServiceClient, entry pb.EmailEntry) *pb.EmailEntry {
	log.Println("Updating email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: &entry})
	logResponse(res, err)
	return res.EmailEntry
}

func deleteEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("Deleting email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{EmailAddr: addr})
	logResponse(res, err)
	return res.EmailEntry
}

var args struct {
	GrpcAddr string `arg:"env:MAILINGLIST_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8081"
	}
	conn, err := grpc.Dial(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewMailingListServiceClient(conn)

	newEmail := createEmail(client, "999@999.99")
	newEmail.ConfirmedAt = 10000
	updateEmail(client, *newEmail)

	deleteEmail(client, "999@999.99")

	getEmailBatch(client, 3, 3)

}
