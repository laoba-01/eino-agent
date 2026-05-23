package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"rag-eval/embedding"

	proto "eino/memory-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Query struct {
	ID          int      `json:"id"`
	Query       string   `json:"query"`
	RelevantIDs []int    `json:"relevant_ids"`
	Category    string   `json:"category"`
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

var memoryAddr = getEnv("MEMORY_SERVICE_ADDR", "localhost:50053")

type queryResult struct {
	query        Query
	retrievedIDs []int64
	scores       []float32
	recallAtK    map[int]float64
	rr           float64
	ndcgAt5      float64
	ndcgAt10     float64
}

func main() {
	dataDir := resolveDataDir()
	queriesPath := filepath.Join(dataDir, "data", "queries.json")

	queries, err := loadQueries(queriesPath)
	if err != nil {
		log.Fatalf("加载测试查询失败: %v", err)
	}
	log.Printf("已加载 %d 个测试查询", len(queries))

	conn, err := grpc.Dial(memoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接 memory-service 失败: %v", err)
	}
	defer conn.Close()
	client := proto.NewMemoryServiceClient(conn)

	const collection = "rag_knowledge"
	const maxK = 10

	results := make([]queryResult, 0, len(queries))

	for i, q := range queries {
		log.Printf("[%d/%d] 评估查询: %s", i+1, len(queries), q.Query)

		vec, err := embedding.Embed(q.Query)
		if err != nil {
			log.Fatalf("嵌入查询 #%d 失败: %v", q.ID, err)
		}

		resp, err := client.SearchSimilar(context.Background(), &proto.SearchSimilarRequest{
			Collection:  collection,
			QueryVector: vec,
			TopK:        maxK,
		})
		if err != nil {
			log.Fatalf("检索查询 #%d 失败: %v", q.ID, err)
		}
		if !resp.Success {
			log.Fatalf("检索查询 #%d 返回失败: %s", q.ID, resp.Error)
		}

		retrievedIDs := make([]int64, 0, len(resp.Results))
		retrievedScores := make([]float32, 0, len(resp.Results))
		for _, r := range resp.Results {
			retrievedIDs = append(retrievedIDs, r.Id)
			retrievedScores = append(retrievedScores, r.Score)
		}

		relevantSet := intSet(q.RelevantIDs)

		qr := queryResult{
			query:        q,
			retrievedIDs: retrievedIDs,
			scores:       retrievedScores,
			recallAtK:    make(map[int]float64),
		}

		for _, k := range []int{1, 3, 5, 10} {
			qr.recallAtK[k] = computeRecall(retrievedIDs, relevantSet, k)
		}
		qr.rr = computeReciprocalRank(retrievedIDs, relevantSet)
		qr.ndcgAt5 = computeNDCG(retrievedIDs, relevantSet, 5)
		qr.ndcgAt10 = computeNDCG(retrievedIDs, relevantSet, 10)

		results = append(results, qr)
	}

	printResults(results, maxK)
}

func computeRecall(retrieved []int64, relevant map[int]bool, k int) float64 {
	if len(relevant) == 0 {
		return 0
	}
	limit := k
	if limit > len(retrieved) {
		limit = len(retrieved)
	}
	hits := 0
	for i := 0; i < limit; i++ {
		if relevant[int(retrieved[i])] {
			hits++
		}
	}
	return float64(hits) / float64(len(relevant))
}

func computeReciprocalRank(retrieved []int64, relevant map[int]bool) float64 {
	for i, id := range retrieved {
		if relevant[int(id)] {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

func computeNDCG(retrieved []int64, relevant map[int]bool, k int) float64 {
	limit := k
	if limit > len(retrieved) {
		limit = len(retrieved)
	}

	dcg := 0.0
	for i := 0; i < limit; i++ {
		rel := 0.0
		if relevant[int(retrieved[i])] {
			rel = 1.0
		}
		dcg += rel / math.Log2(float64(i+2))
	}

	idealRelCount := len(relevant)
	if idealRelCount > limit {
		idealRelCount = limit
	}
	idcg := 0.0
	for i := 0; i < idealRelCount; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

func printResults(results []queryResult, maxK int) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    RAG 召回率评估报告                               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")

	fmt.Printf("║  总查询数: %-3d                                                       ║\n", len(results))
	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")

	fmt.Println("║  各查询详细结果:                                                     ║")
	for _, r := range results {
		q := r.query
		hitInfo := ""
		for j, id := range r.retrievedIDs {
			if j >= 5 {
				break
			}
			marker := " "
			if intSet(q.RelevantIDs)[int(id)] {
				marker = "✓"
			}
			hitInfo += fmt.Sprintf(" #%d%s", id, marker)
		}
		fmt.Printf("║  [%2d] %-30s R@1:%.0f R@3:%.0f R@5:%.0f MRR:%.2f%s║\n",
			q.ID, truncate(q.Query, 28),
			r.recallAtK[1], r.recallAtK[3], r.recallAtK[5],
			r.rr, hitInfo)
	}

	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  汇总指标:                                                           ║")

	avgRecall := map[int]float64{1: 0, 3: 0, 5: 0, 10: 0}
	var avgMRR, avgNDCG5, avgNDCG10 float64
	for _, r := range results {
		for _, k := range []int{1, 3, 5, 10} {
			avgRecall[k] += r.recallAtK[k]
		}
		avgMRR += r.rr
		avgNDCG5 += r.ndcgAt5
		avgNDCG10 += r.ndcgAt10
	}
	n := float64(len(results))
	for _, k := range []int{1, 3, 5, 10} {
		avgRecall[k] /= n
		fmt.Printf("║  Recall@%-2d:  %.4f (%.1f%%)                                          ║\n",
			k, avgRecall[k], avgRecall[k]*100)
	}
	avgMRR /= n
	avgNDCG5 /= n
	avgNDCG10 /= n
	fmt.Printf("║  MRR:       %.4f                                                ║\n", avgMRR)
	fmt.Printf("║  NDCG@5:    %.4f                                                ║\n", avgNDCG5)
	fmt.Printf("║  NDCG@10:   %.4f                                                ║\n", avgNDCG10)

	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")

	// Category breakdown
	catRecall := make(map[string][]float64)
	for _, r := range results {
		cat := r.query.Category
		catRecall[cat] = append(catRecall[cat], r.recallAtK[5])
	}
	fmt.Println("║  分类 Recall@5:                                                      ║")
	for cat, recs := range catRecall {
		sum := 0.0
		for _, v := range recs {
			sum += v
		}
		fmt.Printf("║    %-15s: %.4f (%d 查询)                                  ║\n",
			cat, sum/float64(len(recs)), len(recs))
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
}

func loadQueries(path string) ([]Query, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var queries []Query
	if err := json.Unmarshal(data, &queries); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return queries, nil
}

func intSet(ids []int) map[int]bool {
	s := make(map[int]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func resolveDataDir() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	dir := filepath.Dir(f)
	for {
		if _, err := os.Stat(filepath.Join(dir, "data")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

var _ = sort.Ints
