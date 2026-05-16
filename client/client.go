package main

import (
	"context"
	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"mailingService"
	"time"
)

func logResponse(res *mailingService.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("     Error: %v", err)
	}

	if res.EmailEntry == nil {
		log.Printf("    No email entry")
	} else {
		log.Printf("    Email entry: %v", res.EmailEntry)
	}
}

func createEmail(client mailingService.MailingServiceClient, addr string) *mailingService.EmailEntry {
	log.Printf("Creating email for %s", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailingService.CreateEmailRequest{EmailAddr: addr}
	res, err := client.CreateEmail(ctx, req)
	logResponse(res, err)
	return res.EmailEntry
}

func getEmail(client mailingService.MailingServiceClient, addr string) *mailingService.EmailEntry {
	log.Printf("Get email for %s", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailingService.GetEmailRequest{EmailAddr: addr}
	res, err := client.GetEmail(ctx, req)
	logResponse(res, err)
	return res.EmailEntry
}

func getEmailPaginated(client mailingService.MailingServiceClient, count int, page int) {
	log.Printf("Get email Paginated %d of %d", count, page)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailingService.GetEmailPaginatedRequest{Count: int32(count), Page: int32(page)}
	res, err := client.GetEmailPaginated(ctx, req)
	if err != nil {
		log.Fatalf("     Error: %v", err)
	}
	log.Println("response:")
	for _, entry := range res.EmailEntries {
		log.Println(entry)
	}
}

func updateEmail(client mailingService.MailingServiceClient, entry *mailingService.EmailEntry) *mailingService.EmailEntry {
	log.Printf("Update email for %s", entry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailingService.UpdateEmailRequest{EmailEntry: entry}
	res, err := client.UpdateEmail(ctx, req)
	logResponse(res, err)
	return res.EmailEntry
}

func deleteEmail(client mailingService.MailingServiceClient, addr string) *mailingService.EmailEntry {
	log.Printf("Delete email for %s", addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailingService.DeleteEmailRequest{EmailAddr: addr}
	res, err := client.DeleteEmail(ctx, req)
	logResponse(res, err)
	return res.EmailEntry
}

var args struct {
	GrpcAddr string `arg:"env:MAILING_SERVICE_GRPC_ADDR" help:"Mailing service address"`
}

func main() {
	arg.MustParse(&args)

	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8082"
	}

	clientConn, err := grpc.NewClient(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Note: NewClient doesn't return a *grpc.ClientConn that needs Close() in older style,
	// but it still returns a *grpc.ClientConn that should be closed for cleanup.
	defer clientConn.Close()

	client := mailingService.NewMailingServiceClient(clientConn)

	newEmail := createEmail(client, "test3@test.com")
	newEmail.OptOut = true
	newEmail.ConfirmedAt = time.Now().Unix()
	updateEmail(client, newEmail)
	deleteEmail(client, newEmail.Email)
	getEmailPaginated(client, 5, 1)
}
