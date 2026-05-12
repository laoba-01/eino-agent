package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"eino/memory-service/proto"

	"github.com/go-redis/redis/v8"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"google.golang.org/grpc"
)

const (
	milvusAddr    = "localhost:19530"
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
		Address: milvusAddr,
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

// --- Knowledge point tracking (Redis) ---

type knowledgePointJSON struct {
	Topic       string  `json:"topic"`
	Mastery     float64 `json:"mastery"`
	Interactions int32  `json:"interactions"`
	LastSeen    int64   `json:"last_seen"`
	FirstSeen   int64   `json:"first_seen"`
}

func (s *MemoryServiceServer) SaveKnowledgePoint(ctx context.Context, req *proto.SaveKnowledgePointRequest) (*proto.SaveKnowledgePointResponse, error) {
	if s.rdb == nil {
		return &proto.SaveKnowledgePointResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("learning:%s", req.UserId)
	topic := req.Point.Topic
	now := time.Now().Unix()

	if req.Merge {
		existing, err := s.rdb.HGet(ctx, key, topic).Result()
		if err == nil && existing != "" {
			var kp knowledgePointJSON
			if json.Unmarshal([]byte(existing), &kp) == nil {
				delta := req.Point.Mastery
				if delta < 0.05 {
					delta = 0.05
				} else if delta > 0.3 {
					delta = 0.3
				}
				newMastery := kp.Mastery*0.8 + (kp.Mastery+delta)*0.2
				if newMastery > 1.0 {
					newMastery = 1.0
				}
				if newMastery < 0 {
					newMastery = 0
				}

				kp.Mastery = newMastery
				kp.Interactions++
				kp.LastSeen = now
			}

			data, err := json.Marshal(kp)
			if err != nil {
				return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
			}
			if err := s.rdb.HSet(ctx, key, topic, data).Err(); err != nil {
				return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
			}
		} else {
			kp := knowledgePointJSON{
				Topic:       topic,
				Mastery:     clampMastery(req.Point.Mastery),
				Interactions: 1,
				LastSeen:    now,
				FirstSeen:   now,
			}
			data, err := json.Marshal(kp)
			if err != nil {
				return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
			}
			if err := s.rdb.HSet(ctx, key, topic, data).Err(); err != nil {
				return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
			}

			statsKey := fmt.Sprintf("learning_stats:%s", req.UserId)
			s.rdb.HSet(ctx, statsKey, "study_start_date", now)
		}
	} else {
		kp := knowledgePointJSON{
			Topic:       topic,
			Mastery:     clampMastery(req.Point.Mastery),
			Interactions: req.Point.Interactions,
			LastSeen:    now,
			FirstSeen:   req.Point.FirstSeen,
		}
		if kp.FirstSeen == 0 {
			kp.FirstSeen = now
		}
		data, err := json.Marshal(kp)
		if err != nil {
			return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
		}
		if err := s.rdb.HSet(ctx, key, topic, data).Err(); err != nil {
			return &proto.SaveKnowledgePointResponse{Success: false, Error: err.Error()}, nil
		}
	}

	if err := s.updateLearningStats(ctx, req.UserId); err != nil {
		log.Printf("Warning: failed to update learning stats: %v", err)
	}

	return &proto.SaveKnowledgePointResponse{Success: true, Topic: topic}, nil
}

func (s *MemoryServiceServer) GetKnowledgePoints(ctx context.Context, req *proto.GetKnowledgePointsRequest) (*proto.GetKnowledgePointsResponse, error) {
	if s.rdb == nil {
		return &proto.GetKnowledgePointsResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("learning:%s", req.UserId)
	all, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return &proto.GetKnowledgePointsResponse{Success: false, Error: err.Error()}, nil
	}

	points := make([]*proto.KnowledgePoint, 0, len(all))
	for _, v := range all {
		var kp knowledgePointJSON
		if json.Unmarshal([]byte(v), &kp) == nil {
			points = append(points, &proto.KnowledgePoint{
				Topic:       kp.Topic,
				Mastery:     kp.Mastery,
				Interactions: kp.Interactions,
				LastSeen:    kp.LastSeen,
				FirstSeen:   kp.FirstSeen,
			})
		}
	}

	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "last_seen"
	}
	sort.Slice(points, func(i, j int) bool {
		switch sortBy {
		case "mastery":
			return points[i].Mastery < points[j].Mastery
		case "interactions":
			return points[i].Interactions > points[j].Interactions
		case "last_seen":
			return points[i].LastSeen > points[j].LastSeen
		default:
			return points[i].LastSeen > points[j].LastSeen
		}
	})

	if req.Limit > 0 && int(req.Limit) < len(points) {
		points = points[:req.Limit]
	}

	return &proto.GetKnowledgePointsResponse{Success: true, Points: points}, nil
}

func (s *MemoryServiceServer) GetLearningStats(ctx context.Context, req *proto.GetLearningStatsRequest) (*proto.GetLearningStatsResponse, error) {
	if s.rdb == nil {
		return &proto.GetLearningStatsResponse{Success: false, Error: "redis not connected"}, nil
	}

	statsKey := fmt.Sprintf("learning_stats:%s", req.UserId)
	cached, err := s.rdb.HGetAll(ctx, statsKey).Result()
	if err != nil {
		return &proto.GetLearningStatsResponse{Success: false, Error: err.Error()}, nil
	}

	if len(cached) > 0 {
		return &proto.GetLearningStatsResponse{
			Success:          true,
			TotalTopics:      parseInt(cached["total_topics"]),
			AverageMastery:   parseFloat(cached["average_mastery"]),
			TotalInteractions: parseInt(cached["total_interactions"]),
			MasteredCount:    parseInt(cached["mastered_count"]),
			LearningCount:    parseInt(cached["learning_count"]),
			WeakCount:        parseInt(cached["weak_count"]),
			StudyStartDate:   parseInt64(cached["study_start_date"]),
		}, nil
	}

	if err := s.updateLearningStats(ctx, req.UserId); err != nil {
		return &proto.GetLearningStatsResponse{Success: false, Error: err.Error()}, nil
	}

	cached, err = s.rdb.HGetAll(ctx, statsKey).Result()
	if err != nil || len(cached) == 0 {
		return &proto.GetLearningStatsResponse{Success: true}, nil
	}

	return &proto.GetLearningStatsResponse{
		Success:          true,
		TotalTopics:      parseInt(cached["total_topics"]),
		AverageMastery:   parseFloat(cached["average_mastery"]),
		TotalInteractions: parseInt(cached["total_interactions"]),
		MasteredCount:    parseInt(cached["mastered_count"]),
		LearningCount:    parseInt(cached["learning_count"]),
		WeakCount:        parseInt(cached["weak_count"]),
		StudyStartDate:   parseInt64(cached["study_start_date"]),
	}, nil
}

func (s *MemoryServiceServer) updateLearningStats(ctx context.Context, userID string) error {
	key := fmt.Sprintf("learning:%s", userID)
	all, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}

	var totalMastery float64
	var totalInteractions int32
	var mastered, learning, weak int32
	var studyStartDate int64

	for _, v := range all {
		var kp knowledgePointJSON
		if json.Unmarshal([]byte(v), &kp) != nil {
			continue
		}
		totalMastery += kp.Mastery
		totalInteractions += kp.Interactions
		if kp.Mastery >= 0.8 {
			mastered++
		} else if kp.Mastery >= 0.3 {
			learning++
		} else {
			weak++
		}
		if studyStartDate == 0 || kp.FirstSeen < studyStartDate {
			studyStartDate = kp.FirstSeen
		}
	}

	totalTopics := int32(len(all))
	var avgMastery float64
	if totalTopics > 0 {
		avgMastery = totalMastery / float64(totalTopics)
	}

	statsKey := fmt.Sprintf("learning_stats:%s", userID)
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, statsKey,
		"total_topics", totalTopics,
		"average_mastery", fmt.Sprintf("%.4f", avgMastery),
		"total_interactions", totalInteractions,
		"mastered_count", mastered,
		"learning_count", learning,
		"weak_count", weak,
		"study_start_date", studyStartDate,
	)
	_, err = pipe.Exec(ctx)
	return err
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

func clampMastery(m float64) float64 {
	if m < 0 {
		return 0
	}
	if m > 1 {
		return 1
	}
	return m
}

func parseInt(s string) int32 {
	var v int32
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func main() {
	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
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
