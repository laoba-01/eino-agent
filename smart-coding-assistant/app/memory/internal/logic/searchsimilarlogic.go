package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const defaultTopK = 10

type SearchSimilarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchSimilarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchSimilarLogic {
	return &SearchSimilarLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SearchSimilarLogic) SearchSimilar(in *pb.SearchSimilarRequest) (*pb.SearchSimilarResponse, error) {
	if l.svcCtx.Milvus == nil {
		return &pb.SearchSimilarResponse{Success: false, Error: "milvus not connected"}, nil
	}

	topK := int(in.TopK)
	if topK <= 0 {
		topK = defaultTopK
	}

	if err := l.svcCtx.Milvus.LoadCollection(l.ctx, in.Collection, false); err != nil {
		return &pb.SearchSimilarResponse{Success: false, Error: fmt.Sprintf("load collection: %v", err)}, nil
	}

	queryVec := make([]float32, len(in.QueryVector))
	for i, f := range in.QueryVector {
		queryVec[i] = f
	}

	sp, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		return &pb.SearchSimilarResponse{Success: false, Error: fmt.Sprintf("create search param: %v", err)}, nil
	}

	searchRes, err := l.svcCtx.Milvus.Search(
		l.ctx,
		in.Collection,
		[]string{},
		buildFilterExpr(in.Filter),
		[]string{idField, metaField},
		[]entity.Vector{entity.FloatVector(queryVec)},
		vectorField,
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return &pb.SearchSimilarResponse{Success: false, Error: err.Error()}, nil
	}

	results := make([]*pb.SearchResult, 0)
	if len(searchRes) > 0 {
		var metaCol *entity.ColumnJSONBytes
		for _, col := range searchRes[0].Fields {
			if col.Name() == metaField {
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
			sr := &pb.SearchResult{
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

	return &pb.SearchSimilarResponse{Success: true, Results: results}, nil
}
