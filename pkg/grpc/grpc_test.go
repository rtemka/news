package grpc

import (
	"context"
	"io"
	"log"
	"net"
	"news/pkg/storage/memdb"
	"os"
	"reflect"
	"testing"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const port = ":50051"

func TestMain(m *testing.M) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	api := New(memdb.New(), log.New(io.Discard, "", 0))

	RegisterNewsServer(server, api)

	go server.Serve(listener)

	exitCode := m.Run()
	server.GracefulStop()

	os.Exit(exitCode)
}

func TestAPI_List(t *testing.T) {

	api := New(memdb.New(), log.New(io.Discard, "", 0))

	wantLen := 10

	items, err := api.List(context.Background(), &wrapperspb.Int64Value{Value: int64(wantLen)})
	if err != nil {
		t.Fatalf("API_List() error = %v", err)
	}

	if len(items.Items) != wantLen {
		t.Errorf("API_List() got items = %d, want = %d", len(items.Items), wantLen)
	}

	want := ofStorageItem(&memdb.SampleItem)

	if !reflect.DeepEqual(items.Items[0], want) {
		t.Errorf("API_List() got = %v, want = %v", items.Items[0], &want)
	}
}

func TestAPI(t *testing.T) {

	conn, err := grpc.Dial("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("API error = %v", err)
	}
	defer conn.Close()

	client := NewNewsClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wantLen := 10

	items, err := client.List(ctx, &wrapperspb.Int64Value{Value: int64(wantLen)})
	if err != nil {
		t.Fatalf("API error = %v", err)
	}

	if len(items.Items) != wantLen {
		t.Errorf("API got items = %d, want = %d", len(items.Items), wantLen)
	}
}
