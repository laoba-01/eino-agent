package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"eino/memory-service/proto"

	"github.com/go-redis/redis/v8"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const (
	defaultDim    = 128
	defaultTopK   = 10
	vectorField   = "vector"
	metadataField = "metadata"
	idField       = "id"
)

type MemoryServiceServer struct {
	proto.UnimplementedMemoryServiceServer
	rdb    *redis.Client
	milvus client.Client
}

func newMilvusClient() client.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.NewClient(ctx, client.Config{
		Address: getEnv("MILVUS_ADDR", "localhost:19530"),
	})
	if err != nil {
		log.Printf("Warning: Failed to connect to Milvus: %v", err)
		log.Println("Continuing without Milvus connection...")
		return nil
	}

	ver, err := c.GetVersion(ctx)
	if err != nil {
		log.Printf("Warning: Milvus version check failed: %v", err)
	} else {
		log.Printf("Connected to Milvus %s", ver)
	}
	return c
}

func (s *MemoryServiceServer) ensureCollection(ctx context.Context, collectionName string, dim int) error {
	if s.milvus == nil {
		return fmt.Errorf("milvus not connected")
	}

	has, err := s.milvus.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection failed: %w", err)
	}
	if has {
		return nil
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       idField,
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
			},
			{
				Name:     vectorField,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
			{
				Name:     metadataField,
				DataType: entity.FieldTypeJSON,
			},
		},
	}

	if err := s.milvus.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("create collection failed: %w", err)
	}

	idx, err := entity.NewIndexAUTOINDEX(entity.L2)
	if err != nil {
		return fmt.Errorf("create index config failed: %w", err)
	}
	if err := s.milvus.CreateIndex(ctx, collectionName, vectorField, idx, false); err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	log.Printf("Created Milvus collection %q (dim=%d)", collectionName, dim)
	return nil
}

func (s *MemoryServiceServer) ensureCollectionLoaded(ctx context.Context, collectionName string) error {
	if s.milvus == nil {
		return fmt.Errorf("milvus not connected")
	}

	if err := s.milvus.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("load collection failed: %w", err)
	}
	return nil
}

// --- Key-value context (Redis) ---

func (s *MemoryServiceServer) SaveContext(ctx context.Context, req *proto.SaveContextRequest) (*proto.SaveContextResponse, error) {
	if s.rdb == nil {
		return &proto.SaveContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", req.UserId)
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, key, mapToInterface(req.Context)...)
	if req.Ttl > 0 {
		pipe.Expire(ctx, key, time.Duration(req.Ttl)*time.Second)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return &proto.SaveContextResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SaveContextResponse{Success: true}, nil
}

func (s *MemoryServiceServer) GetContext(ctx context.Context, req *proto.GetContextRequest) (*proto.GetContextResponse, error) {
	if s.rdb == nil {
		return &proto.GetContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", req.UserId)

	if len(req.Keys) > 0 {
		vals, err := s.rdb.HMGet(ctx, key, req.Keys...).Result()
		if err != nil {
			return &proto.GetContextResponse{Success: false, Error: err.Error()}, nil
		}
		result := make(map[string]string)
		for i, k := range req.Keys {
			if vals[i] != nil {
				result[k] = vals[i].(string)
			}
		}
		return &proto.GetContextResponse{Context: result, Success: true}, nil
	}

	result, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return &proto.GetContextResponse{Success: false, Error: err.Error()}, nil
	}
	return &proto.GetContextResponse{Context: result, Success: true}, nil
}

func (s *MemoryServiceServer) DeleteContext(ctx context.Context, req *proto.DeleteContextRequest) (*proto.DeleteContextResponse, error) {
	if s.rdb == nil {
		return &proto.DeleteContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", req.UserId)

	if len(req.Keys) > 0 {
		if err := s.rdb.HDel(ctx, key, req.Keys...).Err(); err != nil {
			return &proto.DeleteContextResponse{Success: false, Error: err.Error()}, nil
		}
	} else {
		if err := s.rdb.Del(ctx, key).Err(); err != nil {
			return &proto.DeleteContextResponse{Success: false, Error: err.Error()}, nil
		}
	}

	return &proto.DeleteContextResponse{Success: true}, nil
}

func (s *MemoryServiceServer) UpdateContext(ctx context.Context, req *proto.UpdateContextRequest) (*proto.UpdateContextResponse, error) {
	if s.rdb == nil {
		return &proto.UpdateContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", req.UserId)
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, key, mapToInterface(req.Context)...)
	if req.Ttl > 0 {
		pipe.Expire(ctx, key, time.Duration(req.Ttl)*time.Second)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return &proto.UpdateContextResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdateContextResponse{Success: true}, nil
}

// --- Vector operations (Milvus) ---

func (s *MemoryServiceServer) SaveVector(ctx context.Context, req *proto.SaveVectorRequest) (*proto.SaveVectorResponse, error) {
	if s.milvus == nil {
		return &proto.SaveVectorResponse{Success: false, Error: "milvus not connected"}, nil
	}

	dim := defaultDim
	if len(req.Vectors) > 0 && len(req.Vectors[0].Vector) > 0 {
		dim = len(req.Vectors[0].Vector)
	}

	if err := s.ensureCollection(ctx, req.Collection, dim); err != nil {
		return &proto.SaveVectorResponse{Success: false, Error: err.Error()}, nil
	}

	ids := make([]int64, 0, len(req.Vectors))
	idCol := make([]int64, 0, len(req.Vectors))
	vecCol := make([][]float32, 0, len(req.Vectors))
	metaCol := make([][]byte, 0, len(req.Vectors))

	for _, v := range req.Vectors {
		idCol = append(idCol, v.Id)
		ids = append(ids, v.Id)
		vec := make([]float32, len(v.Vector))
		for i, f := range v.Vector {
			vec[i] = f
		}
		vecCol = append(vecCol, vec)

		jsonBytes, err := json.Marshal(v.Metadata)
		if err != nil {
			return &proto.SaveVectorResponse{Success: false, Error: fmt.Sprintf("marshal metadata for id %d: %v", v.Id, err)}, nil
		}
		metaCol = append(metaCol, jsonBytes)
	}

	columns := []entity.Column{
		entity.NewColumnInt64(idField, idCol),
		entity.NewColumnFloatVector(vectorField, dim, vecCol),
		entity.NewColumnJSONBytes(metadataField, metaCol),
	}

	if _, err := s.milvus.Insert(ctx, req.Collection, "", columns...); err != nil {
		return &proto.SaveVectorResponse{Success: false, Error: err.Error()}, nil
	}

	if err := s.milvus.Flush(ctx, req.Collection, false); err != nil {
		log.Printf("Warning: flush failed: %v", err)
	}

	return &proto.SaveVectorResponse{Success: true, InsertedIds: ids}, nil
}

func (s *MemoryServiceServer) SearchSimilar(ctx context.Context, req *proto.SearchSimilarRequest) (*proto.SearchSimilarResponse, error) {
	if s.milvus == nil {
		return &proto.SearchSimilarResponse{Success: false, Error: "milvus not connected"}, nil
	}

	topK := int(req.TopK)
	if topK <= 0 {
		topK = defaultTopK
	}

	if err := s.ensureCollectionLoaded(ctx, req.Collection); err != nil {
		return &proto.SearchSimilarResponse{Success: false, Error: err.Error()}, nil
	}

	queryVec := make([]float32, len(req.QueryVector))
	for i, f := range req.QueryVector {
		queryVec[i] = f
	}

	sp, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		return &proto.SearchSimilarResponse{Success: false, Error: fmt.Sprintf("create search param: %v", err)}, nil
	}

	searchRes, err := s.milvus.Search(
		ctx,
		req.Collection,
		[]string{},
		buildFilterExpr(req.Filter),
		[]string{idField, metadataField},
		[]entity.Vector{entity.FloatVector(queryVec)},
		vectorField,
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return &proto.SearchSimilarResponse{Success: false, Error: err.Error()}, nil
	}

	results := make([]*proto.SearchResult, 0)
	if len(searchRes) > 0 {
		// Find metadata column from Fields
		var metaCol *entity.ColumnJSONBytes
		for _, col := range searchRes[0].Fields {
			if col.Name() == metadataField {
				if jc, ok := col.(*entity.ColumnJSONBytes); ok {
					metaCol = jc
					break
				}
			}
		}

		for i, hit := range searchRes[0].IDs.(*entity.ColumnInt64).Data() {
			score := float32(0)
			if i < len(searchRes[0].Scores) {
				score = searchRes[0].Scores[i]
			}
			sr := &proto.SearchResult{
				Id:       hit,
				Score:    score,
				Metadata: make(map[string]string),
			}
			if metaCol != nil {
				if bs, err := metaCol.ValueByIdx(i); err == nil {
					var m map[string]string
					if json.Unmarshal(bs, &m) == nil {
						sr.Metadata = m
					}
				}
			}
			results = append(results, sr)
		}
	}

	return &proto.SearchSimilarResponse{Success: true, Results: results}, nil
}

func (s *MemoryServiceServer) DeleteVector(ctx context.Context, req *proto.DeleteVectorRequest) (*proto.DeleteVectorResponse, error) {
	if s.milvus == nil {
		return &proto.DeleteVectorResponse{Success: false, Error: "milvus not connected"}, nil
	}

	ids := make([]string, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = fmt.Sprintf("%d", id)
	}
	expr := fmt.Sprintf("id in [%s]", joinInt64s(req.Ids))

	if err := s.milvus.Delete(ctx, req.Collection, "", expr); err != nil {
		return &proto.DeleteVectorResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeleteVectorResponse{Success: true}, nil
}

// --- Helpers ---

func mapToInterface(m map[string]string) []interface{} {
	result := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		result = append(result, k, v)
	}
	return result
}

func buildFilterExpr(filter map[string]string) string {
	if len(filter) == 0 {
		return ""
	}
	exprs := make([]string, 0, len(filter))
	for k, v := range filter {
		exprs = append(exprs, fmt.Sprintf(`metadata["%s"] == "%s"`, k, v))
	}
	result := ""
	for i, e := range exprs {
		if i > 0 {
			result += " and "
		}
		result += e
	}
	return result
}

func joinInt64s(ids []int64) string {
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%d", id)
	}
	return s
}

func main() {
	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis connection...")
	}

	// Milvus
	milvusClient := newMilvusClient()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterMemoryServiceServer(s, &MemoryServiceServer{
		rdb:    rdb,
		milvus: milvusClient,
	})

	fmt.Println("Memory Service listening on port 50053...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
