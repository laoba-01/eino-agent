package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	defaultDim  = 128
	vectorField = "vector"
	metaField   = "metadata"
	idField     = "id"
)

type SaveVectorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveVectorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveVectorLogic {
	return &SaveVectorLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SaveVectorLogic) SaveVector(in *pb.SaveVectorRequest) (*pb.SaveVectorResponse, error) {
	if l.svcCtx.Milvus == nil {
		return &pb.SaveVectorResponse{Success: false, Error: "milvus not connected"}, nil
	}

	dim := defaultDim
	if len(in.Vectors) > 0 && len(in.Vectors[0].Vector) > 0 {
		dim = len(in.Vectors[0].Vector)
	}

	// 懒加载 + 缓存: 仅首次插入时检查/创建 collection
	if err := ensureCollectionCached(l.ctx, l.svcCtx, in.Collection, dim); err != nil {
		return &pb.SaveVectorResponse{Success: false, Error: err.Error()}, nil
	}

	ids := make([]int64, 0, len(in.Vectors))
	idCol := make([]int64, 0, len(in.Vectors))
	vecCol := make([][]float32, 0, len(in.Vectors))
	metaCol := make([][]byte, 0, len(in.Vectors))

	for _, v := range in.Vectors {
		idCol = append(idCol, v.Id)
		ids = append(ids, v.Id)
		vec := make([]float32, len(v.Vector))
		for i, f := range v.Vector {
			vec[i] = f
		}
		vecCol = append(vecCol, vec)

		jsonBytes, err := json.Marshal(v.Metadata)
		if err != nil {
			return &pb.SaveVectorResponse{Success: false, Error: fmt.Sprintf("marshal metadata for id %d: %v", v.Id, err)}, nil
		}
		metaCol = append(metaCol, jsonBytes)
	}

	columns := []entity.Column{
		entity.NewColumnInt64(idField, idCol),
		entity.NewColumnFloatVector(vectorField, dim, vecCol),
		entity.NewColumnJSONBytes(metaField, metaCol),
	}

	if _, err := l.svcCtx.Milvus.Insert(l.ctx, in.Collection, "", columns...); err != nil {
		return &pb.SaveVectorResponse{Success: false, Error: err.Error()}, nil
	}

	if err := l.svcCtx.Milvus.Flush(l.ctx, in.Collection, false); err != nil {
		log.Printf("Warning: flush failed: %v", err)
	}

	return &pb.SaveVectorResponse{Success: true, InsertedIds: ids}, nil
}

// ensureCollectionCached 懒加载 + 缓存: 仅首次调用时检查/创建 collection
// 后续调用直接跳过，避免每次插入都发 HasCollection RPC
func ensureCollectionCached(ctx context.Context, svcCtx *svc.ServiceContext, collectionName string, dim int) error {
	// 快速路径: collection 已确认存在
	if _, loaded := svcCtx.LoadedCols.Load(collectionName); loaded {
		return nil
	}

	// 慢路径: 首次检查 + 创建
	has, err := svcCtx.Milvus.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection failed: %w", err)
	}

	if !has {
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
					Name:     metaField,
					DataType: entity.FieldTypeJSON,
				},
			},
		}

		if err := svcCtx.Milvus.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
			return fmt.Errorf("create collection failed: %w", err)
		}

		idx, err := entity.NewIndexAUTOINDEX(entity.L2)
		if err != nil {
			return fmt.Errorf("create index config failed: %w", err)
		}
		if err := svcCtx.Milvus.CreateIndex(ctx, collectionName, vectorField, idx, false); err != nil {
			return fmt.Errorf("create index failed: %w", err)
		}

		log.Printf("Created Milvus collection %q (dim=%d)", collectionName, dim)
	}

	// 标记为已加载，后续调用走快速路径
	svcCtx.LoadedCols.Store(collectionName, true)
	return nil
}
