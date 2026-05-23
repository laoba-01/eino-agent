package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"rag-eval/embedding"
)

type Document struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Query struct {
	ID          int    `json:"id"`
	Query       string `json:"query"`
	RelevantIDs []int  `json:"relevant_ids"`
	Category    string `json:"category"`
}

type evalResult struct {
	query        Query
	retrieved    []int
	similarities []float32
	recallAt     map[int]float64
	rr           float64
	ndcg5        float64
	ndcg10       float64
}

func main() {
	dataDir := resolveDataDir()
	docsPath := filepath.Join(dataDir, "data", "documents.json")
	queriesPath := filepath.Join(dataDir, "data", "queries.json")
	cachePath := filepath.Join(dataDir, "data", "embedding_cache.json")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║       RAG 召回率评估 (离线模式 - 余弦相似度)               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Load data
	docs, err := loadDocs(docsPath)
	if err != nil {
		fatalf("加载文档失败: %v", err)
	}
	fmt.Printf("║  已加载 %d 个知识文档                                          ║\n", len(docs))

	queries, err := loadQueries(queriesPath)
	if err != nil {
		fatalf("加载测试查询失败: %v", err)
	}
	fmt.Printf("║  已加载 %d 个测试查询                                          ║\n", len(queries))
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Embed documents (with cache)
	docVecs := embedDocs(docs, cachePath)

	// Embed queries
	fmt.Println("║  正在嵌入测试查询...                                          ║")
	queryVecs := make([][]float32, len(queries))
	for i, q := range queries {
		vec, err := embedding.Embed(q.Query)
		if err != nil {
			fatalf("嵌入查询 #%d 失败: %v", q.ID, err)
		}
		queryVecs[i] = vec
		fmt.Printf("\r║  查询嵌入进度: %d/%d", i+1, len(queries))
	}
	fmt.Println()
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Evaluate
	const maxK = 10

	results := make([]evalResult, 0, len(queries))
	for qi, qv := range queryVecs {
		q := queries[qi]

		type scoredDoc struct {
			id   int
			sim  float32
		}
		scored := make([]scoredDoc, len(docVecs))
		for di, dv := range docVecs {
			scored[di] = scoredDoc{id: docs[di].ID, sim: cosineSimilarity(qv, dv)}
		}
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].sim > scored[j].sim
		})

		limit := maxK
		if limit > len(scored) {
			limit = len(scored)
		}
		retrieved := make([]int, limit)
		sims := make([]float32, limit)
		for i := 0; i < limit; i++ {
			retrieved[i] = scored[i].id
			sims[i] = scored[i].sim
		}

		relevantSet := intSet(q.RelevantIDs)
		r := evalResult{
			query:        q,
			retrieved:    retrieved,
			similarities: sims,
			recallAt:     make(map[int]float64),
		}
		for _, k := range []int{1, 3, 5, 10} {
			r.recallAt[k] = recallAt(retrieved, relevantSet, k)
		}
		r.rr = reciprocalRank(retrieved, relevantSet)
		r.ndcg5 = ndcgAt(retrieved, relevantSet, 5)
		r.ndcg10 = ndcgAt(retrieved, relevantSet, 10)
		results = append(results, r)
	}

	// Print results
	printResults(results, maxK)
}

func embedDocs(docs []Document, cachePath string) [][]float32 {
	cache := loadCache(cachePath)

	vecs := make([][]float32, 0, len(docs))
	needEmbed := make([]int, 0)

	for _, doc := range docs {
		key := fmt.Sprintf("%d", doc.ID)
		if v, ok := cache[key]; ok && len(v) > 0 {
			vecs = append(vecs, v)
		} else {
			vecs = append(vecs, nil)
			needEmbed = append(needEmbed, len(vecs)-1)
		}
	}

	if len(needEmbed) == 0 {
		fmt.Printf("║  文档向量已缓存 (%d 个)                                        ║\n", len(docs))
		return vecs
	}

	fmt.Printf("║  正在嵌入文档 (%d 个需处理)...                                ║\n", len(needEmbed))
	for i, idx := range needEmbed {
		doc := docs[idx]
		vec, err := embedding.Embed(doc.Content)
		if err != nil {
			fatalf("嵌入文档 #%d 失败: %v", doc.ID, err)
		}
		vecs[idx] = vec
		cache[fmt.Sprintf("%d", doc.ID)] = vec
		fmt.Printf("\r║  文档嵌入进度: %d/%d (dim=%d)", i+1, len(needEmbed), len(vec))

		if (i+1)%5 == 0 && i+1 < len(needEmbed) {
			fmt.Print("  [等待500ms]")
			time.Sleep(500 * time.Millisecond)
		}
	}
	fmt.Println()
	saveCache(cachePath, cache)
	return vecs
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

func recallAt(retrieved []int, relevant map[int]bool, k int) float64 {
	if len(relevant) == 0 {
		return 0
	}
	limit := k
	if limit > len(retrieved) {
		limit = len(retrieved)
	}
	hits := 0
	for i := 0; i < limit; i++ {
		if relevant[retrieved[i]] {
			hits++
		}
	}
	return float64(hits) / float64(len(relevant))
}

func reciprocalRank(retrieved []int, relevant map[int]bool) float64 {
	for i, id := range retrieved {
		if relevant[id] {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

func ndcgAt(retrieved []int, relevant map[int]bool, k int) float64 {
	limit := k
	if limit > len(retrieved) {
		limit = len(retrieved)
	}

	dcg := 0.0
	for i := 0; i < limit; i++ {
		rel := 0.0
		if relevant[retrieved[i]] {
			rel = 1.0
		}
		dcg += rel / math.Log2(float64(i+2))
	}

	idealCount := len(relevant)
	if idealCount > limit {
		idealCount = limit
	}
	idcg := 0.0
	for i := 0; i < idealCount; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

func printResults(results []evalResult, maxK int) {
	fmt.Println("║  各查询详细结果 (top-5):                                       ║")
	for _, r := range results {
		q := r.query
		var parts []string
		for j := 0; j < 5 && j < len(r.retrieved); j++ {
			marker := " "
			if intSet(q.RelevantIDs)[r.retrieved[j]] {
				marker = "+"
			}
			parts = append(parts, fmt.Sprintf("#%d%s(%.3f)", r.retrieved[j], marker, r.similarities[j]))
		}
		fmt.Printf("║  [%2d] %-28s R@1:%.1f R@5:%.1f MRR:%.2f\n",
			q.ID, truncate(q.Query, 26),
			r.recallAt[1], r.recallAt[5], r.rr)
		fmt.Printf("║       %s\n", strings.Join(parts, " "))
	}

	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  汇总指标:                                                    ║")

	avgRecall := map[int]float64{1: 0, 3: 0, 5: 0, 10: 0}
	var avgMRR, avgNDCG5, avgNDCG10 float64
	for _, r := range results {
		for _, k := range []int{1, 3, 5, 10} {
			avgRecall[k] += r.recallAt[k]
		}
		avgMRR += r.rr
		avgNDCG5 += r.ndcg5
		avgNDCG10 += r.ndcg10
	}
	n := float64(len(results))
	fmt.Printf("║  Recall@1:   %.4f  (%.1f%%)                                      ║\n", avgRecall[1]/n, avgRecall[1]/n*100)
	fmt.Printf("║  Recall@3:   %.4f  (%.1f%%)                                      ║\n", avgRecall[3]/n, avgRecall[3]/n*100)
	fmt.Printf("║  Recall@5:   %.4f  (%.1f%%)                                      ║\n", avgRecall[5]/n, avgRecall[5]/n*100)
	fmt.Printf("║  Recall@10:  %.4f  (%.1f%%)                                      ║\n", avgRecall[10]/n, avgRecall[10]/n*100)
	fmt.Printf("║  MRR:        %.4f                                             ║\n", avgMRR/n)
	fmt.Printf("║  NDCG@5:     %.4f                                             ║\n", avgNDCG5/n)
	fmt.Printf("║  NDCG@10:    %.4f                                             ║\n", avgNDCG10/n)

	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  分类 Recall@5:                                               ║")
	catRecall := make(map[string][]float64)
	for _, r := range results {
		catRecall[r.query.Category] = append(catRecall[r.query.Category], r.recallAt[5])
	}
	for cat, recs := range catRecall {
		sum := 0.0
		for _, v := range recs {
			sum += v
		}
		fmt.Printf("║    %-16s  %.4f  (%d 查询)                           ║\n",
			cat, sum/float64(len(recs)), len(recs))
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func loadDocs(path string) ([]Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var docs []Document
	return docs, json.Unmarshal(data, &docs)
}

func loadQueries(path string) ([]Query, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var qs []Query
	return qs, json.Unmarshal(data, &qs)
}

type embedCache map[string][]float32

func loadCache(path string) embedCache {
	data, err := os.ReadFile(path)
	if err != nil {
		return make(embedCache)
	}
	var c embedCache
	if json.Unmarshal(data, &c) != nil {
		return make(embedCache)
	}
	return c
}

func saveCache(path string, c embedCache) {
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
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

func fatalf(format string, args ...interface{}) {
	fmt.Printf("║  ERROR: "+format+"\n", args...)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	os.Exit(1)
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
