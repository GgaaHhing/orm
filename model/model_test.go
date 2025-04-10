package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnderscoreCase(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"空字符串", "", ""},
		{"单个小写", "user", "user"},
		{"单个大写", "User", "user"},
		{"标准驼峰", "userName", "user_name"},
		{"大写驼峰", "UserName", "user_name"},
		{"连续大写", "UserID", "user_id"},
		{"多个单词", "FirstMiddleName", "first_middle_name"},
		{"全大写", "ID", "id"},
		{"混合大小写", "MyXMLParser", "my_xml_parser"},
		{"数字", "User123Name", "user123_name"},
		{"特殊字符", "User_Name", "user_name"},
		{"复杂组合", "ThisIsALongVariableName", "this_is_a_long_variable_name"},
		{"首字母缩写", "APIResponse", "api_response"},
		{"数字混合", "Get2ndValue", "get2nd_value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := underscoreCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
