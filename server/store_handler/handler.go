package store_handler

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	pb "baiyecha/ipvs-manager/grpc/proto"
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
	db *badger.DB
}
 func NewGrpcStoreServer(db *badger.DB)*GrpcStoreServer{
	 return &GrpcStoreServer{
		 db: db,
	 }
 }
