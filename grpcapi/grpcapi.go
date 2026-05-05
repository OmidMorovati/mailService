package grpcapi

import (
	"context"
	"database/sql"
	"google.golang.org/grpc"
	"log"
	"mailingService"
	"mailingService/mdb"
	"net"
	"time"
)

type MailServer struct {
	mailingService.UnimplementedMailingServiceServer
	db *sql.DB
}

func pbEntryToMdbEntry(pbEntry *mailingService.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)
	return mdb.EmailEntry{EmailAddress: pbEntry.Email, OptOut: pbEntry.OptOut, ConfirmedAt: &t, Id: pbEntry.Id}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) mailingService.EmailEntry {
	return mailingService.EmailEntry{
		Id:          mdbEntry.Id,
		Email:       mdbEntry.EmailAddress,
		OptOut:      mdbEntry.OptOut,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(),
	}
}

func emailResponse(db *sql.DB, email string) (*mailingService.EmailResponse, error) {
	entry, err := mdb.GetEmail(db, email)
	if err != nil {
		return &mailingService.EmailResponse{}, err
	}

	if entry == nil {
		return &mailingService.EmailResponse{}, nil
	}

	res := mdbEntryToPbEntry(entry)
	return &mailingService.EmailResponse{EmailEntry: &res}, nil
}

func (s *MailServer) GetEmail(ctx context.Context, req *mailingService.GetEmailRequest) (*mailingService.EmailResponse, error) {
	log.Printf("gRPC GetEmail:%v\n", req)
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) GetEmailPaginated(ctx context.Context, req *mailingService.GetEmailPaginatedRequest) (*mailingService.GetEmailPaginatedResponse, error) {
	log.Printf("gRPC GetEmailPaginated:%v\n", req)

	params := mdb.GetEmailPaginatedQueryParams{Page: int(req.Page), Count: int(req.Count)}

	mdbEntries, err := mdb.GetEmailPaginated(s.db, params)
	if err != nil {
		return &mailingService.GetEmailPaginatedResponse{}, err
	}

	pbEntries := make([]*mailingService.EmailEntry, len(mdbEntries))

	for _, entry := range mdbEntries {
		pbEntry := mdbEntryToPbEntry(&entry)
		pbEntries = append(pbEntries, &pbEntry)
	}
	return &mailingService.GetEmailPaginatedResponse{EmailEntries: pbEntries}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, req *mailingService.CreateEmailRequest) (*mailingService.EmailResponse, error) {
	log.Printf("gRPC CreateEmail:%v\n", req)
	err := mdb.CreateEmail(s.db, req.EmailAddr)
	if err != nil {
		return &mailingService.EmailResponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) UpdateEmail(ctx context.Context, req *mailingService.UpdateEmailRequest) (*mailingService.EmailResponse, error) {
	log.Printf("gRPC UpdateEmail:%v\n", req)

	entry := pbEntryToMdbEntry(req.EmailEntry)
	err := mdb.UpdateEmail(s.db, entry)
	if err != nil {
		return &mailingService.EmailResponse{}, err
	}
	return emailResponse(s.db, entry.EmailAddress)
}

func (s *MailServer) DeleteEmail(ctx context.Context, req *mailingService.DeleteEmailRequest) (*mailingService.EmailResponse, error) {
	log.Printf("gRPC DeleteEmail:%v\n", req)

	err := mdb.DeleteEmail(s.db, req.EmailAddr)
	if err != nil {
		return &mailingService.EmailResponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

func Serve(db *sql.DB, bind string) {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("gRPC server 	failed to listen: %v", bind)
	}

	grpcServer := grpc.NewServer()
	mailServer := &MailServer{db: db}
	mailingService.RegisterMailingServiceServer(grpcServer, mailServer)

	log.Printf("gRPC server listening on: %v\n", bind)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server failed to serve: %v", err)
	}
}
