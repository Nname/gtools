package mapuniq

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sync"
)

var (
	// 使用 map 存储不同 label 类型的标签集合
	labels = make(map[string][]map[string]interface{})
	// 每个 labelType 对应的锁
	labelsMutex = make(map[string]*sync.RWMutex)
	// 每个 labelType 对应的哈希集合
	hashSets = make(map[string]map[uint32][]string)
)

// mapToJSONString 将 map[string]interface{} 转换为 JSON 字符串
func mapToJSONString(m map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// hashString 计算字符串的哈希值
func hashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// getMutex 获取指定 labelType 对应的锁，若不存在则初始化
func getMutex(labelType string) *sync.RWMutex {
	if _, exists := labelsMutex[labelType]; !exists {
		labelsMutex[labelType] = &sync.RWMutex{}
	}
	return labelsMutex[labelType]
}

// AddLabel 通用函数：将 map 添加到指定的类型标签集合中，确保不重复
func AddLabel(labelType string, data map[string]interface{}) error {
	// 获取该 labelType 对应的读写锁
	labelsMutex := getMutex(labelType)

	// 使用写锁，因为我们要修改标签数据
	labelsMutex.Lock()
	defer labelsMutex.Unlock()

	// 如果 labels[labelType] 为 nil，初始化为空切片
	if labels[labelType] == nil {
		labels[labelType] = []map[string]interface{}{} // 初始化空的 map 切片
	}
	// 如果 hashSets[labelType] 为 nil，初始化为空 map
	if hashSets[labelType] == nil {
		hashSets[labelType] = make(map[uint32][]string) // 初始化空的 map
	}

	// 获取 labelType 对应的哈希集合和标签集合
	hashSet := hashSets[labelType]
	// 将 map 转为 JSON 字符串
	dataJSON, err := mapToJSONString(data)
	if err != nil {
		return fmt.Errorf("failed to convert map to JSON: %v", err)
	}

	// 计算 JSON 字符串的哈希值
	dataHash := hashString(dataJSON)

	// 检查哈希集
	if existingJSONs, found := hashSet[dataHash]; found {
		for _, existingJSON := range existingJSONs {
			if existingJSON == dataJSON {
				return nil // 已存在相同内容的 map，跳过
			}
		}
	}

	// 若没有重复项，则添加到数组和哈希集
	labels[labelType] = append(labels[labelType], data)
	hashSet[dataHash] = append(hashSet[dataHash], dataJSON)
	return nil
}

// GetLabels 获取指定类型的去重后的结果
func GetLabels(labelType string) []map[string]interface{} {
	// 获取该 labelType 对应的读写锁
	labelsMutex := getMutex(labelType)

	// 使用读锁，避免修改时阻塞读取操作
	labelsMutex.RLock()
	defer labelsMutex.RUnlock()

	return labels[labelType]
}

// GetHashSet 获取指定 labelType 的哈希集合
func GetHashSet(labelType string) (map[uint32][]string, error) {
	// 获取该 labelType 对应的读写锁
	labelsMutex := getMutex(labelType)

	// 使用读锁来避免修改时阻塞读取操作
	labelsMutex.RLock()
	defer labelsMutex.RUnlock()

	// 获取 hashSet，如果不存在则返回错误
	hashSet, exists := hashSets[labelType]
	if !exists {
		return nil, fmt.Errorf("hash set for labelType '%s' not found", labelType)
	}

	return hashSet, nil
}
