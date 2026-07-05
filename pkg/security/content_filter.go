package security

import (
	"fmt"
	"regexp"
	"strings"
)

// ==================== 敏感词库 ====================

// sensitiveWords 包含暴力、色情、政治敏感、诈骗等违规词汇
var sensitiveWords = []string{
	// 暴力犯罪
	"暴力", "犯罪", "枪支", "爆炸", "杀人", "抢劫", "绑架", "偷窃", "盗窃",
	"贩毒", "毒品", "吸毒", "走私", "赌博", "嫖娼", "卖淫", "强奸", "乱伦",
	// 政治敏感
	"邪教", "反动", "分裂", "颠覆", "叛国", "台独", "港独", "疆独",
	// 色情低俗
	"色情", "淫秽", "裸露", "性器官", "性交", "做爱", "自慰", "变态", "AV",
	"三级片", "无码", "露点", "波霸", "奶子", "操你", "妈逼", "傻逼", "狗日",
	// 辱骂攻击
	"去死", "滚蛋", "垃圾", "废物", "脑残", "智障", "神经病", "傻逼", "混蛋",
	// 诈骗传销
	"诈骗", "传销", "非法集资", "庞氏骗局", "诈骗集团", "钓鱼网站",
	// 网络攻击
	"黑客", "病毒", "木马", "破解", "入侵", "攻击", "DDOS", "端口扫描",
	// 其他违规
	"违法", "违规", "举报", "投诉", "维权", "上访",
}

// ==================== 危险模式正则 ====================

// urlPattern 匹配超链接（http/https/www）
var urlPattern = regexp.MustCompile(`(?i)(https?://[^\s]+|www\.[^\s]+)`)

// htmlPattern 匹配HTML标签
var htmlPattern = regexp.MustCompile(`<[^>]*>`)

// scriptPattern 匹配JavaScript代码
var scriptPattern = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>|javascript:|eval\(|alert\(|confirm\(|prompt\(`)

// sqlPattern 匹配SQL注入模式
var sqlPattern = regexp.MustCompile(`(?i)(select\s+|insert\s+|update\s+|delete\s+|drop\s+|union\s+|exec\s+|execute\s+|xp_|sp_|--\s*|/\*.*?\*/|\b(or|and)\s+\d+\s*=\s*\d+)`)

// xssPattern 匹配XSS攻击模式
var xssPattern = regexp.MustCompile(`(?i)(on\w+\s*=\s*["'].*?["']|<img\s+|onerror\s*=|onload\s*=|onclick\s*=)`)

// ==================== 安全检查函数 ====================

// ContainsSensitiveWord 检查内容是否包含敏感词
// 参数：content - 待检查内容
// 返回：bool - 是否包含敏感词，error - 错误信息（包含具体违规词）
func ContainsSensitiveWord(content string) (bool, error) {
	if content == "" {
		return false, nil
	}

	lowerContent := strings.ToLower(content)
	for _, word := range sensitiveWords {
		if strings.Contains(lowerContent, strings.ToLower(word)) {
			return true, fmt.Errorf("内容包含违规词汇：%s", word)
		}
	}

	return false, nil
}

// ContainsURL 检查内容是否包含超链接
// 参数：content - 待检查内容
// 返回：bool - 是否包含URL
func ContainsURL(content string) bool {
	return urlPattern.MatchString(content)
}

// ContainsHTML 检查内容是否包含HTML标签
// 参数：content - 待检查内容
// 返回：bool - 是否包含HTML标签
func ContainsHTML(content string) bool {
	return htmlPattern.MatchString(content)
}

// ContainsScript 检查内容是否包含JavaScript代码
// 参数：content - 待检查内容
// 返回：bool - 是否包含脚本代码
func ContainsScript(content string) bool {
	return scriptPattern.MatchString(content)
}

// ContainsSQLInjection 检查内容是否包含SQL注入模式
// 参数：content - 待检查内容
// 返回：bool - 是否包含SQL注入模式
func ContainsSQLInjection(content string) bool {
	return sqlPattern.MatchString(content)
}

// ContainsXSS 检查内容是否包含XSS攻击模式
// 参数：content - 待检查内容
// 返回：bool - 是否包含XSS模式
func ContainsXSS(content string) bool {
	return xssPattern.MatchString(content)
}

// ==================== 综合验证函数 ====================

// ValidateNickname 验证昵称安全性
// 安全规则：
//   1. 长度限制：2-20个字符
//   2. 字符限制：仅允许中文、英文、数字、下划线、连字符
//   3. 敏感词过滤：禁止包含敏感词汇
//   4. URL过滤：禁止包含超链接
//   5. HTML过滤：禁止包含HTML标签
//   6. XSS过滤：禁止包含XSS攻击代码
func ValidateNickname(nickname string) error {
	if nickname == "" {
		return nil
	}

	// 【昵称安全规则1】长度检查
	if len(nickname) < 2 || len(nickname) > 20 {
		return fmt.Errorf("昵称长度必须在2-20个字符之间")
	}

	// 【昵称安全规则2】字符类型检查（仅允许中文、英文、数字、下划线、连字符）
	charPattern := regexp.MustCompile(`^[\p{Han}a-zA-Z0-9_-]+$`)
	if !charPattern.MatchString(nickname) {
		return fmt.Errorf("昵称只能包含中文、英文、数字、下划线和连字符")
	}

	// 【昵称安全规则3】敏感词检查
	if hasSensitive, err := ContainsSensitiveWord(nickname); hasSensitive {
		return fmt.Errorf("昵称包含违规内容：%w", err)
	}

	// 【昵称安全规则4】URL检查
	if ContainsURL(nickname) {
		return fmt.Errorf("昵称不能包含网址链接")
	}

	// 【昵称安全规则5】HTML检查
	if ContainsHTML(nickname) {
		return fmt.Errorf("昵称不能包含HTML标签")
	}

	// 【昵称安全规则6】XSS检查
	if ContainsXSS(nickname) {
		return fmt.Errorf("昵称包含非法内容")
	}

	return nil
}

// ValidateSignText 验证个性签名安全性
// 安全规则：
//   1. 长度限制：最大200字符
//   2. 敏感词过滤：禁止包含敏感词汇
//   3. URL过滤：禁止包含超链接
//   4. HTML过滤：禁止包含HTML标签
//   5. JavaScript过滤：禁止包含脚本代码
//   6. SQL注入过滤：禁止包含SQL注入代码
//   7. XSS过滤：禁止包含XSS攻击代码
func ValidateSignText(signText string) error {
	if signText == "" {
		return nil
	}

	// 【签名安全规则1】长度检查
	if len(signText) > 200 {
		return fmt.Errorf("签名长度不能超过200字符")
	}

	// 【签名安全规则2】敏感词检查
	if hasSensitive, err := ContainsSensitiveWord(signText); hasSensitive {
		return fmt.Errorf("签名包含违规内容：%w", err)
	}

	// 【签名安全规则3】URL检查
	if ContainsURL(signText) {
		return fmt.Errorf("签名不能包含网址链接")
	}

	// 【签名安全规则4】HTML检查
	if ContainsHTML(signText) {
		return fmt.Errorf("签名不能包含HTML标签")
	}

	// 【签名安全规则5】JavaScript检查
	if ContainsScript(signText) {
		return fmt.Errorf("签名不能包含脚本代码")
	}

	// 【签名安全规则6】SQL注入检查
	if ContainsSQLInjection(signText) {
		return fmt.Errorf("签名包含非法内容")
	}

	// 【签名安全规则7】XSS检查
	if ContainsXSS(signText) {
		return fmt.Errorf("签名包含非法内容")
	}

	return nil
}