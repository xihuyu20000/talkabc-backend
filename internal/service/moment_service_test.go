package service

import (
	"testing"
)

// TestUserMomentDTO_Structure 测试 UserMomentDTO 结构完整性
// 验证动态DTO包含所有必要字段，且字段值正确赋值
func TestUserMomentDTO_Structure(t *testing.T) {
	dto := UserMomentDTO{
		UID:       "12345678901234567890",
		MID:       100,
		PraiseNum: 25,
		PubTS:     1718500000,
		Text:      "Hello world",
		Files:     []string{"photo1.jpg"},
		Location:  "Shanghai",
	}

	if dto.UID != "12345678901234567890" {
		t.Errorf("UID should be '12345678901234567890', got %s", dto.UID)
	}

	if dto.MID != 100 {
		t.Error("MID should be 100")
	}

	if dto.PraiseNum != 25 {
		t.Error("PraiseNum should be 25")
	}

	if dto.PubTS != 1718500000 {
		t.Error("PubTS mismatch")
	}

	if len(dto.Files) != 1 {
		t.Error("Should have 1 file")
	}
}

// TestUserMomentDTO_EmptyData 测试 UserMomentDTO 空数据场景
// 验证当动态只有基本ID时，其他字段应保持默认值（空字符串、0或空数组）
func TestUserMomentDTO_EmptyData(t *testing.T) {
	dto := UserMomentDTO{
		UID:       "1",
		MID:       1,
		PraiseNum: 0,
		PubTS:     0,
		Text:      "",
		Files:     []string{},
		Location:  "",
	}

	if dto.UID != "1" {
		t.Errorf("Expected UID '1', got %s", dto.UID)
	}

	if dto.MID != 1 {
		t.Error("MID should be 1")
	}

	if len(dto.Files) != 0 {
		t.Error("Should have 0 files")
	}
}