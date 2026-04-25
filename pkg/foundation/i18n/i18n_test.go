package i18n

import (
	"context"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"en/email.yaml": {Data: []byte(`
subject:
  verification_code: "Verify Your Email"
TicketCreatedTitle: "Ticket Created"
status:
  open: "Pending"
`)},
		"zh_CN/email.yaml": {Data: []byte(`
subject:
  verification_code: "验证您的邮箱"
TicketCreatedTitle: "工单已创建"
status:
  open: "待处理"
`)},
	}
}

func TestInitAndTranslate(t *testing.T) {
	// 重置全局状态
	mu.Lock()
	global = nil
	mu.Unlock()

	if err := InitWithFS(Config{DefaultLocale: "zh_CN", FallbackLocale: "en"}, testFS(), "."); err != nil {
		t.Fatalf("InitWithFS 失败: %v", err)
	}

	// 检查 locale 列表
	locales := Locales()
	if len(locales) != 2 {
		t.Fatalf("期望 2 个 locale，实际 %d: %v", len(locales), locales)
	}

	// 中文翻译
	ctx := WithLocale(context.Background(), "zh_CN")
	got := T(ctx, "email.subject.verification_code", "Verify Your Email")
	if got != "验证您的邮箱" {
		t.Errorf("zh_CN 翻译错误: got %q", got)
	}

	got = T(ctx, "email.TicketCreatedTitle", "Ticket Created")
	if got != "工单已创建" {
		t.Errorf("zh_CN TicketCreatedTitle 错误: got %q", got)
	}

	got = TWithLocale("zh_CN", "email.status.open", "open")
	if got != "待处理" {
		t.Errorf("zh_CN status.open 错误: got %q", got)
	}

	// 英文翻译
	ctx2 := WithLocale(context.Background(), "en")
	got = T(ctx2, "email.subject.verification_code", "Verify Your Email")
	if got != "Verify Your Email" {
		t.Errorf("en 翻译错误: got %q", got)
	}
}

func TestFallback(t *testing.T) {
	mu.Lock()
	global = &i18n{
		locales: map[string]map[string]string{
			"en": {"email.foo": "English Foo"},
		},
		defaultLocale:  "zh_CN",
		fallbackLocale: "en",
	}
	mu.Unlock()

	// zh_CN 没有该 key，应 fallback 到 en
	ctx := WithLocale(context.Background(), "zh_CN")
	got := T(ctx, "email.foo", "default")
	if got != "English Foo" {
		t.Errorf("fallback 到 en 失败: got %q", got)
	}

	// 完全不存在时使用 defaultValue
	got = T(ctx, "email.nonexistent", "my default")
	if got != "my default" {
		t.Errorf("默认值 fallback 失败: got %q", got)
	}
}

func TestLocaleFromCtx(t *testing.T) {
	mu.Lock()
	global = &i18n{
		locales:        map[string]map[string]string{},
		defaultLocale:  "zh_CN",
		fallbackLocale: "en",
	}
	mu.Unlock()

	// 无 locale 时返回 defaultLocale
	got := Locale(context.Background())
	if got != "zh_CN" {
		t.Errorf("默认 locale 错误: got %q", got)
	}

	// 设置后返回设置的值
	ctx := WithLocale(context.Background(), "en")
	got = Locale(ctx)
	if got != "en" {
		t.Errorf("设置后 locale 错误: got %q", got)
	}
}

func TestResolveLocale(t *testing.T) {
	mu.Lock()
	global = &i18n{
		locales: map[string]map[string]string{
			"en":    {},
			"zh_CN": {},
		},
		defaultLocale:  "zh_CN",
		fallbackLocale: "en",
	}
	mu.Unlock()

	cases := []struct {
		header string
		want   string
	}{
		{"zh-CN,zh;q=0.9,en;q=0.8", "zh_CN"},
		{"en-US,en;q=0.9", "en"},
		{"en", "en"},
		{"zh_CN", "zh_CN"},
		{"", "zh_CN"},      // 无 header → defaultLocale
		{"fr-FR", "zh_CN"}, // 不存在 → defaultLocale
	}

	for _, c := range cases {
		got := resolveLocale(c.header)
		if got != c.want {
			t.Errorf("resolveLocale(%q) = %q, want %q", c.header, got, c.want)
		}
	}
}
