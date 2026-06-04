package logic

import (
	"context"
	"fmt"
	"strings"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
)

type GenerateProblemSolutionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateProblemSolutionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateProblemSolutionLogic {
	return &GenerateProblemSolutionLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GenerateProblemSolutionLogic) GenerateProblemSolution(in *pb.GenerateProblemSolutionRequest) (*pb.GenerateProblemSolutionResponse, error) {
	problem := in.GetProblem()
	difficulty := in.GetDifficulty()
	lang := in.GetLanguage()

	approach, code, explanation := generateSolution(problem, difficulty, lang)
	return &pb.GenerateProblemSolutionResponse{
		Approach:    approach,
		Code:        code,
		Explanation: explanation,
		Success:     true,
	}, nil
}

// problemTemplates 常见算法问题的解题模板
var problemTemplates = map[string]struct {
	approach    string
	code        func(lang string) string
	explanation string
}{
	"two sum": {
		"使用哈希表在 O(n) 时间内找到两个和为 target 的数。遍历数组，对于每个元素，检查 target - num 是否已在哈希表中。",
		func(l string) string {
			switch strings.ToLower(l) {
			case "python":
				return "def twoSum(nums, target):\n    seen = {}\n    for i, num in enumerate(nums):\n        complement = target - num\n        if complement in seen:\n            return [seen[complement], i]\n        seen[num] = i\n    return []"
			case "go":
				return "func twoSum(nums []int, target int) []int {\n    seen := make(map[int]int)\n    for i, num := range nums {\n        if j, ok := seen[target-num]; ok {\n            return []int{j, i}\n        }\n        seen[num] = i\n    }\n    return nil\n}"
			default:
				return "function twoSum(nums, target) {\n    const seen = new Map();\n    for (let i = 0; i < nums.length; i++) {\n        const complement = target - nums[i];\n        if (seen.has(complement)) {\n            return [seen.get(complement), i];\n        }\n        seen.set(nums[i], i);\n    }\n    return [];\n}"
			}
		},
		"哈希表查找是 O(1)，整体时间复杂度 O(n)，空间复杂度 O(n)。遍历一次数组，边遍历边检查。",
	},
	"reverse": {
		"翻转问题通常使用双指针法，从两端向中间交换元素，O(n) 时间、O(1) 空间。",
		func(l string) string {
			return "func reverse(s []rune) {\n    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {\n        s[i], s[j] = s[j], s[i]\n    }\n}"
		},
		"双指针从两端交换，每轮交换后指针向中间移动一步。适用于字符串、数组翻转。",
	},
	"palindrome": {
		"判断回文用双指针从两端比较字符，相等则向中间移动，直到指针相遇。",
		func(l string) string {
			return "func isPalindrome(s string) bool {\n    i, j := 0, len(s)-1\n    for i < j {\n        if s[i] != s[j] {\n            return false\n        }\n        i++\n        j--\n    }\n    return true\n}"
		},
		"时间复杂度 O(n)，空间复杂度 O(1)。也可用反转后比较但需要 O(n) 额外空间。",
	},
	"fibonacci": {
		"斐波那契数列可用动态规划/迭代避免递归的指数级复杂度。使用两个变量存储前两个值。",
		func(l string) string {
			return "func fibonacci(n int) int {\n    if n <= 1 { return n }\n    a, b := 0, 1\n    for i := 2; i <= n; i++ {\n        a, b = b, a+b\n    }\n    return b\n}"
		},
		"迭代法 O(n) 时间、O(1) 空间。递归法有 O(2^n) 指数复杂度，需要记忆化优化。",
	},
	"sort": {
		"排序问题优先使用语言内置排序函数。自定义排序需提供比较器/排序键。",
		func(l string) string {
			switch strings.ToLower(l) {
			case "python":
				return "arr.sort()  # 升序\narr.sort(reverse=True)  # 降序\narr.sort(key=lambda x: x[1])  # 自定义键"
			case "go":
				return "sort.Ints(arr)\nsort.Slice(arr, func(i, j int) bool {\n    return arr[i] < arr[j]\n})"
			default:
				return "arr.sort((a, b) => a - b);  // 升序\narr.sort((a, b) => b - a);  // 降序"
			}
		},
		"标准库排序通常是 O(n log n)。如需实现排序算法，从冒泡排序开始理解，再到快速排序和归并排序。",
	},
	"binary search": {
		"二分查找将搜索空间每次减半，O(log n) 时间复杂度。需要数组有序。",
		func(l string) string {
			return "func binarySearch(arr []int, target int) int {\n    lo, hi := 0, len(arr)-1\n    for lo <= hi {\n        mid := lo + (hi-lo)/2\n        if arr[mid] == target { return mid }\n        if arr[mid] < target { lo = mid + 1 } else { hi = mid - 1 }\n    }\n    return -1\n}"
		},
		"注意 mid 计算用 lo+(hi-lo)/2 防止溢出。适用于有序数组、单调函数、峰值查找。",
	},
	"linked list": {
		"链表问题通常使用快慢指针（Floyd 判圈）、哑节点（简化头操作）、递归等技巧。",
		func(l string) string {
			return "type ListNode struct {\n    Val  int\n    Next *ListNode\n}\n\nfunc reverseList(head *ListNode) *ListNode {\n    var prev *ListNode\n    for head != nil {\n        next := head.Next\n        head.Next = prev\n        prev = head\n        head = next\n    }\n    return prev\n}"
		},
		"链表反转是基础操作。快慢指针可检测环、找中点。哑节点简化删除/插入操作。",
	},
	"tree": {
		"树问题常用 DFS（递归/栈）或 BFS（队列层序遍历）。二叉树遍历有前序、中序、后序。",
		func(l string) string {
			return "type TreeNode struct {\n    Val   int\n    Left  *TreeNode\n    Right *TreeNode\n}\n\nfunc inorder(root *TreeNode) {\n    if root == nil { return }\n    inorder(root.Left)\n    fmt.Println(root.Val)\n    inorder(root.Right)\n}"
		},
		"DFS 用递归简洁直观；BFS 用队列逐层处理。平衡树操作 O(log n)，退化链表 O(n)。",
	},
}

func generateSolution(problem, difficulty, lang string) (string, string, string) {
	problemLower := strings.ToLower(problem)

	// 尝试匹配问题模板
	for key, tmpl := range problemTemplates {
		if strings.Contains(problemLower, key) || strings.Contains(key, problemLower) {
			diffNote := ""
			if difficulty != "" {
				diffNote = fmt.Sprintf("\n\n**难度**: %s", difficulty)
			}
			return tmpl.approach + diffNote, tmpl.code(lang), tmpl.explanation
		}
	}

	// 未匹配到模板，给出通用方案
	approach := fmt.Sprintf("### 解题思路\n\n对于「%s」，建议按以下步骤分析：\n\n", problem)
	approach += fmt.Sprintf("1. **理解问题**: 明确输入输出格式和约束条件\n")
	approach += fmt.Sprintf("2. **分析复杂度**: 根据数据规模选择合适的算法（O(n), O(n log n), O(n^2) 等）\n")
	approach += "3. **选择数据结构和算法**: 常见选择包括哈希表（快速查找）、双指针（数组问题）、DP（最优子结构）、BFS/DFS（图/树遍历）\n"
	approach += "4. **边界条件**: 空值、单元素、极大/极小值等\n"
	if difficulty != "" {
		approach += fmt.Sprintf("\n**难度**: %s\n", difficulty)
	}

	code := fmt.Sprintf("// %s 解法 (%s)\n// 难度: %s\n// 请根据具体问题调整实现\n\nfunc solve(input) output {\n    // TODO: 实现具体逻辑\n    // 1. 处理边界条件\n    // 2. 核心算法逻辑\n    // 3. 返回结果\n    return result\n}", problem, lang, difficulty)

	explanation := "### 代码说明\n\n"
	explanation += fmt.Sprintf("以上代码框架提供了一个 %s 语言的解题起点。\n", lang)
	explanation += "请根据具体问题填充算法逻辑。\n"
	explanation += "\n**常用优化技巧**:\n"
	explanation += "- 使用哈希表减少查找时间 O(1)\n"
	explanation += "- 双指针避免 O(n²) 嵌套循环\n"
	explanation += "- 记忆化递归避免重复计算\n"
	explanation += "- 前缀和/差分数组优化区间查询\n"

	return approach, code, explanation
}
