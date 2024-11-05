package mapuniq

import (
	"fmt"
	"testing"
)

func TestMapUniq(t *testing.T) {
	// 测试数据
	data1 := map[string]interface{}{
		"alertname": "cpu_high",
		"severity":  "critical",
		"instance":  "server1",
	}
	data2 := map[string]interface{}{
		"alertname": "cpu_high",
		"severity":  "critical",
		"instance":  "server2",
	}
	data3 := map[string]interface{}{
		"alertname": "cpu_high",
		"severity":  "critical",
		"instance":  "server1", // 与 data1 相同
	}

	// 使用 labelType "receive" 来添加标签
	labelType := "receive"

	// 清理测试数据
	labels[labelType] = nil
	hashSets[labelType] = nil

	// 异常处理
	if err := AddLabel(labelType, nil); err != nil {
		t.Errorf("Failed to add label: %v", err)
	}

	if err := AddLabel("", nil); err != nil {
		t.Errorf("Failed to add label: %v", err)
	}

	// 添加第一个标签
	if err := AddLabel(labelType, data1); err != nil {
		t.Errorf("Failed to add label: %v", err)
	}

	// 添加第二个标签
	if err := AddLabel(labelType, data2); err != nil {
		t.Errorf("Failed to add label: %v", err)
	}

	// 尝试添加重复的标签 (应该不会添加)
	if err := AddLabel(labelType, data3); err != nil {
		t.Errorf("Failed to add label: %v", err)
	}

	// 获取去重后的标签集合
	labelsReceived := GetLabels(labelType)
	if len(labelsReceived) != 3 {
		t.Errorf("Expected 2 unique labels, got %d", len(labelsReceived))
	}

	// 获取哈希集合
	hashSet, err := GetHashSet(labelType)
	if err != nil {
		t.Errorf("Failed to get hash set: %v", err)
	} else {
		fmt.Printf("Hash Set for '%s': %v\n", labelType, hashSet)
	}

	// 检查是否标签集合和哈希集合的长度一致
	if len(labelsReceived) != len(hashSet) {
		t.Errorf("Labels count mismatch with hash set, labels: %d, hash set: %d", len(labelsReceived), len(hashSet))
	}

	//
	fmt.Println(labels)

	// 清理测试数据
	labels[labelType] = nil
	hashSets[labelType] = nil

}
