package store_handler

import (
	pb "baiyecha/ipvs-manager/grpc/proto"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

// handler struct handler
type handler struct {
	raft           *raft.Raft
	db             *badger.DB
	clusterAddress []string
}

func New(raft *raft.Raft, db *badger.DB, clusterAddress []string) *handler {
	return &handler{
		raft:           raft,
		db:             db,
		clusterAddress: clusterAddress,
	}
}

type GrpcStoreServer struct {
	pb.UnimplementedIpvsListServiceServer
	db             *badger.DB
	clusterAddress []string
}

func NewGrpcStoreServer(db *badger.DB, clusterAddress []string) *GrpcStoreServer {
	return &GrpcStoreServer{
		db:             db,
		clusterAddress: clusterAddress,
	}
}
