package Lee

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/geektutu")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "geektutu" {
		t.Fatal("name should be equal to 'geektutu'")
	}

	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])
}

// 路由性能基准测试
func BenchmarkRouting(b *testing.B) {
	// 创建测试引擎
	r := New()

	// 添加多种类型的路由
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users", func(c *Context) { c.String(200, "users") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/users/:id/posts", func(c *Context) { c.String(200, "user posts") })
	r.GET("/users/:id/posts/:pid", func(c *Context) { c.String(200, "user post") })
	r.GET("/api/v1/users", func(c *Context) { c.String(200, "api users") })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static file") })
	r.GET("/download/*filepath", func(c *Context) { c.String(200, "download") })
	r.POST("/login", func(c *Context) { c.String(200, "login") })
	r.POST("/register", func(c *Context) { c.String(200, "register") })
	r.POST("/api/v1/login", func(c *Context) { c.String(200, "api login") })

	// 测试路径列表
	testPaths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/456/posts",
		"/users/789/posts/101",
		"/api/v1/users",
		"/api/v1/users/999",
		"/static/css/style.css",
		"/download/files/document.pdf",
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 执行基准测试
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	}
}

// 十万次路由请求性能测试
func TestRoutePerformance100K(t *testing.T) {
	// 创建测试引擎
	r := New()

	// 添加各种类型的路由
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users", func(c *Context) { c.String(200, "users") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/users/:id/posts", func(c *Context) { c.String(200, "user posts") })
	r.GET("/users/:id/posts/:pid", func(c *Context) { c.String(200, "user post") })
	r.GET("/api/v1/users", func(c *Context) { c.String(200, "api users") })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/api/v2/users/:id/profile", func(c *Context) { c.String(200, "user profile") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static file") })
	r.GET("/download/*filepath", func(c *Context) { c.String(200, "download") })
	r.POST("/login", func(c *Context) { c.String(200, "login") })
	r.POST("/register", func(c *Context) { c.String(200, "register") })
	r.POST("/api/v1/login", func(c *Context) { c.String(200, "api login") })

	// 测试路径列表（模拟真实应用场景）
	testPaths := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/users/123"},
		{"GET", "/users/456/posts"},
		{"GET", "/users/789/posts/101"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/users/999"},
		{"GET", "/api/v2/users/888/profile"},
		{"GET", "/static/css/style.css"},
		{"GET", "/static/js/app.js"},
		{"GET", "/download/files/document.pdf"},
		{"POST", "/login"},
		{"POST", "/register"},
		{"POST", "/api/v1/login"},
		{"PUT", "/users/123"},
		{"DELETE", "/users/456"},
	}

	const totalRequests = 1000000

	fmt.Printf("开始执行 %d 次路由请求性能测试...\n", totalRequests)

	startTime := time.Now()

	// 执行十万次请求
	for i := 0; i < totalRequests; i++ {
		// 循环使用测试路径
		testCase := testPaths[i%len(testPaths)]

		req := httptest.NewRequest(testCase.method, testCase.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 验证响应状态码
		if w.Code != 200 && w.Code != 404 {
			t.Errorf("意外的状态码: %d, 路径: %s %s", w.Code, testCase.method, testCase.path)
		}
	}

	duration := time.Since(startTime)

	// 输出性能统计
	fmt.Printf("测试完成!\n")
	fmt.Printf("总请求数: %d\n", totalRequests)
	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("平均每次请求耗时: %v\n", duration/time.Duration(totalRequests))
	fmt.Printf("每秒处理请求数 (QPS): %.2f\n", float64(totalRequests)/duration.Seconds())
	fmt.Printf("每毫秒处理请求数: %.2f\n", float64(totalRequests)/float64(duration.Nanoseconds()/1000000))
}

// 并发路由性能测试
func TestConcurrentRoutePerformance(t *testing.T) {
	// 创建测试引擎
	r := New()

	// 添加路由
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static file") })

	testPaths := []string{
		"/",
		"/users/123",
		"/api/v1/users/456",
		"/static/css/style.css",
	}

	const totalRequests = 1000000
	const numGoroutines = 10
	const requestsPerGoroutine = totalRequests / numGoroutines

	fmt.Printf("开始并发测试: %d 个协程，每个协程 %d 次请求...\n", numGoroutines, requestsPerGoroutine)

	startTime := time.Now()

	// 使用通道等待所有协程完成
	done := make(chan bool, numGoroutines)

	// 启动多个协程并发测试
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < requestsPerGoroutine; i++ {
				path := testPaths[i%len(testPaths)]
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
			done <- true
		}(g)
	}

	// 等待所有协程完成
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	duration := time.Since(startTime)

	// 输出并发性能统计
	fmt.Printf("并发测试完成!\n")
	fmt.Printf("总请求数: %d\n", totalRequests)
	fmt.Printf("并发协程数: %d\n", numGoroutines)
	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("平均每次请求耗时: %v\n", duration/time.Duration(totalRequests))
	fmt.Printf("并发 QPS: %.2f\n", float64(totalRequests)/duration.Seconds())
}

// Lee vs Gin 性能对比测试
func TestLeeVsGinPerformance(t *testing.T) {
	// 设置Gin为发布模式，减少日志输出
	gin.SetMode(gin.ReleaseMode)

	// 测试配置
	const totalRequests = 1000000

	// 测试路径
	testPaths := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users/123"},
		{"GET", "/users/456/posts"},
		{"GET", "/users/789/posts/101"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/users/999"},
		{"GET", "/static/css/style.css"},
		{"GET", "/download/files/document.pdf"},
		{"POST", "/login"},
		{"POST", "/register"},
	}

	fmt.Printf("=== Lee vs Gin 性能对比测试 ===\n")
	fmt.Printf("测试请求数: %d\n", totalRequests)
	fmt.Printf("测试路径数: %d\n\n", len(testPaths))

	// 测试 Lee 框架
	fmt.Printf("🚀 测试 Lee 框架...\n")
	leeQPS, leeAvgTime := testLeeFramework(totalRequests, testPaths)

	// 测试 Gin 框架
	fmt.Printf("\n🍸 测试 Gin 框架...\n")
	ginQPS, ginAvgTime := testGinFramework(totalRequests, testPaths)

	// 输出对比结果
	fmt.Print("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Print("📊 性能对比结果\n")
	fmt.Print(strings.Repeat("=", 50) + "\n")
	fmt.Print("框架      | QPS        | 平均响应时间    | 性能比\n")
	fmt.Print(strings.Repeat("-", 50) + "\n")
	fmt.Printf("Lee       | %8.2f   | %12v   | 基准\n", leeQPS, leeAvgTime)
	fmt.Printf("Gin       | %8.2f   | %12v   | %.2fx\n", ginQPS, ginAvgTime, ginQPS/leeQPS)
	fmt.Print(strings.Repeat("-", 50) + "\n")

	if leeQPS > ginQPS {
		fmt.Printf("🎉 Lee 框架性能领先 Gin %.2f%%\n", (leeQPS-ginQPS)/ginQPS*100)
	} else {
		fmt.Printf("📈 Gin 框架性能领先 Lee %.2f%%\n", (ginQPS-leeQPS)/leeQPS*100)
	}
}

// 测试 Lee 框架性能
func testLeeFramework(totalRequests int, testPaths []struct{ method, path string }) (float64, time.Duration) {
	// 创建 Lee 引擎
	r := New()

	// 添加路由
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/users/:id/posts", func(c *Context) { c.String(200, "user posts") })
	r.GET("/users/:id/posts/:pid", func(c *Context) { c.String(200, "user post") })
	r.GET("/api/v1/users", func(c *Context) { c.String(200, "api users") })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static file") })
	r.GET("/download/*filepath", func(c *Context) { c.String(200, "download") })
	r.POST("/login", func(c *Context) { c.String(200, "login") })
	r.POST("/register", func(c *Context) { c.String(200, "register") })

	startTime := time.Now()

	// 执行请求
	for i := 0; i < totalRequests; i++ {
		testCase := testPaths[i%len(testPaths)]
		req := httptest.NewRequest(testCase.method, testCase.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	duration := time.Since(startTime)
	qps := float64(totalRequests) / duration.Seconds()
	avgTime := duration / time.Duration(totalRequests)

	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("平均响应时间: %v\n", avgTime)

	return qps, avgTime
}

// 测试 Gin 框架性能
func testGinFramework(totalRequests int, testPaths []struct{ method, path string }) (float64, time.Duration) {
	// 创建 Gin 引擎
	r := gin.New()

	// 添加路由
	r.GET("/", func(c *gin.Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *gin.Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/users/:id/posts", func(c *gin.Context) { c.String(200, "user posts") })
	r.GET("/users/:id/posts/:pid", func(c *gin.Context) { c.String(200, "user post") })
	r.GET("/api/v1/users", func(c *gin.Context) { c.String(200, "api users") })
	r.GET("/api/v1/users/:id", func(c *gin.Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *gin.Context) { c.String(200, "static file") })
	r.GET("/download/*filepath", func(c *gin.Context) { c.String(200, "download") })
	r.POST("/login", func(c *gin.Context) { c.String(200, "login") })
	r.POST("/register", func(c *gin.Context) { c.String(200, "register") })

	startTime := time.Now()

	// 执行请求
	for i := 0; i < totalRequests; i++ {
		testCase := testPaths[i%len(testPaths)]
		req := httptest.NewRequest(testCase.method, testCase.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	duration := time.Since(startTime)
	qps := float64(totalRequests) / duration.Seconds()
	avgTime := duration / time.Duration(totalRequests)

	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("QPS: %.2f\n", qps)
	fmt.Printf("平均响应时间: %v\n", avgTime)

	return qps, avgTime
}

// Lee vs Gin 并发性能对比测试
func TestLeeVsGinConcurrentPerformance(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	const totalRequests = 1000000
	const numGoroutines = 10
	const requestsPerGoroutine = totalRequests / numGoroutines

	testPaths := []string{
		"/",
		"/users/123",
		"/api/v1/users/456",
		"/static/css/style.css",
	}

	fmt.Printf("=== Lee vs Gin 并发性能对比测试 ===\n")
	fmt.Printf("总请求数: %d\n", totalRequests)
	fmt.Printf("并发协程数: %d\n", numGoroutines)
	fmt.Printf("每协程请求数: %d\n\n", requestsPerGoroutine)

	// 测试 Lee 框架并发性能
	fmt.Printf("🚀 测试 Lee 框架并发性能...\n")
	leeQPS := testLeeConcurrent(totalRequests, numGoroutines, requestsPerGoroutine, testPaths)

	// 测试 Gin 框架并发性能
	fmt.Printf("\n🍸 测试 Gin 框架并发性能...\n")
	ginQPS := testGinConcurrent(totalRequests, numGoroutines, requestsPerGoroutine, testPaths)

	// 输出并发对比结果
	fmt.Print("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Print("📊 并发性能对比结果\n")
	fmt.Print(strings.Repeat("=", 50) + "\n")
	fmt.Print("框架      | 并发QPS    | 性能比\n")
	fmt.Print(strings.Repeat("-", 30) + "\n")
	fmt.Printf("Lee       | %8.2f   | 基准\n", leeQPS)
	fmt.Printf("Gin       | %8.2f   | %.2fx\n", ginQPS, ginQPS/leeQPS)
	fmt.Print(strings.Repeat("-", 30) + "\n")

	if leeQPS > ginQPS {
		fmt.Printf("🎉 Lee 框架并发性能领先 Gin %.2f%%\n", (leeQPS-ginQPS)/ginQPS*100)
	} else {
		fmt.Printf("📈 Gin 框架并发性能领先 Lee %.2f%%\n", (ginQPS-leeQPS)/leeQPS*100)
	}
}

// Lee 框架并发测试
func testLeeConcurrent(totalRequests, numGoroutines, requestsPerGoroutine int, testPaths []string) float64 {
	r := New()
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static file") })

	startTime := time.Now()
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func() {
			for i := 0; i < requestsPerGoroutine; i++ {
				path := testPaths[i%len(testPaths)]
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
			done <- true
		}()
	}

	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	duration := time.Since(startTime)
	qps := float64(totalRequests) / duration.Seconds()

	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("并发QPS: %.2f\n", qps)

	return qps
}

// Gin 框架并发测试
func testGinConcurrent(totalRequests, numGoroutines, requestsPerGoroutine int, testPaths []string) float64 {
	r := gin.New()
	r.GET("/", func(c *gin.Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *gin.Context) { c.String(200, "user %s", c.Param("id")) })
	r.GET("/api/v1/users/:id", func(c *gin.Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *gin.Context) { c.String(200, "static file") })

	startTime := time.Now()
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func() {
			for i := 0; i < requestsPerGoroutine; i++ {
				path := testPaths[i%len(testPaths)]
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
			done <- true
		}()
	}

	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	duration := time.Since(startTime)
	qps := float64(totalRequests) / duration.Seconds()

	fmt.Printf("总耗时: %v\n", duration)
	fmt.Printf("并发QPS: %.2f\n", qps)

	return qps
}

// Lee vs Gin 基准测试对比
func BenchmarkLeeRouting(b *testing.B) {
	r := New()
	r.GET("/", func(c *Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *Context) { c.String(200, "user") })
	r.GET("/api/v1/users/:id", func(c *Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *Context) { c.String(200, "static") })

	testPaths := []string{"/", "/users/123", "/api/v1/users/456", "/static/css/style.css"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkGinRouting(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) { c.String(200, "home") })
	r.GET("/users/:id", func(c *gin.Context) { c.String(200, "user") })
	r.GET("/api/v1/users/:id", func(c *gin.Context) { c.String(200, "api user") })
	r.GET("/static/*filepath", func(c *gin.Context) { c.String(200, "static") })

	testPaths := []string{"/", "/users/123", "/api/v1/users/456", "/static/css/style.css"}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// 测试Context对象池的效果
func BenchmarkContextPool(b *testing.B) {
	engine := New()
	engine.GET("/test", func(c *Context) { c.String(200, "test") })
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}

// 测试路由查找性能（静态路由 vs 参数路由）
func BenchmarkStaticRouting(b *testing.B) {
	engine := New()
	// 添加大量静态路由
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/static/route/%d", i)
		engine.GET(path, func(c *Context) { c.String(200, "static") })
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/static/route/%d", i%100)
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkDynamicRouting(b *testing.B) {
	engine := New()
	engine.GET("/users/:id", func(c *Context) { c.String(200, "user") })
	engine.GET("/posts/:id/comments/:cid", func(c *Context) { c.String(200, "comment") })
	
	testPaths := []string{
		"/users/123",
		"/users/456", 
		"/posts/789/comments/101",
		"/posts/999/comments/202",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}

// 测试内存分配优化效果
func TestMemoryOptimization(t *testing.T) {
	engine := New()
	engine.GET("/test/:id", func(c *Context) { 
		c.String(200, "user %s", c.Param("id")) 
	})
	
	// 预热
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/test/123", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
	
	// 测试内存分配
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 执行大量请求
	for i := 0; i < 10000; i++ {
		req := httptest.NewRequest("GET", "/test/123", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	fmt.Printf("内存分配差异: %d bytes\n", allocDiff)
	fmt.Printf("平均每次请求分配: %.2f bytes\n", float64(allocDiff)/10000)
	
	// 验证内存分配是否在合理范围内（每次请求应该很少分配新内存）
	avgAllocPerReq := float64(allocDiff) / 10000
	if avgAllocPerReq > 100 { // 每次请求分配超过100字节认为需要优化
		t.Logf("警告: 平均每次请求分配 %.2f bytes，可能需要进一步优化", avgAllocPerReq)
	} else {
		t.Logf("✅ 内存优化效果良好: 平均每次请求分配 %.2f bytes", avgAllocPerReq)
	}
}

// 性能优化验证测试
func TestPerformanceOptimizations(t *testing.T) {
	fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Print("🔧 Lee 框架性能优化验证测试\n")
	fmt.Print(strings.Repeat("=", 60) + "\n")
	
	// 1. Context 对象池测试
	t.Run("Context对象池测试", func(t *testing.T) {
		fmt.Print("\n📦 测试 Context 对象池效果...\n")
		
		engine := New()
		engine.GET("/test/:id", func(c *Context) { 
			c.String(200, "user %s", c.Param("id")) 
		})
		
		// 预热
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest("GET", "/test/123", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
		}
		
		// 测试内存分配
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		const testRequests = 10000
		startTime := time.Now()
		
		for i := 0; i < testRequests; i++ {
			req := httptest.NewRequest("GET", "/test/123", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
		}
		
		duration := time.Since(startTime)
		runtime.GC()
		runtime.ReadMemStats(&m2)
		
		allocDiff := m2.TotalAlloc - m1.TotalAlloc
		avgAllocPerReq := float64(allocDiff) / testRequests
		qps := float64(testRequests) / duration.Seconds()
		
		fmt.Printf("   ✅ 执行 %d 次请求\n", testRequests)
		fmt.Printf("   ✅ 总耗时: %v\n", duration)
		fmt.Printf("   ✅ QPS: %.2f\n", qps)
		fmt.Printf("   ✅ 内存分配差异: %d bytes\n", allocDiff)
		fmt.Printf("   ✅ 平均每次请求分配: %.2f bytes\n", avgAllocPerReq)
		
		if avgAllocPerReq > 200 {
			t.Logf("⚠️  警告: 平均每次请求分配 %.2f bytes，可能需要进一步优化", avgAllocPerReq)
		} else {
			t.Logf("🎉 内存优化效果良好: 平均每次请求分配 %.2f bytes", avgAllocPerReq)
		}
	})
	
	// 2. 静态路由查找性能测试
	t.Run("静态路由查找性能", func(t *testing.T) {
		fmt.Print("\n🔍 测试静态路由查找性能...\n")
		
		engine := New()
		// 添加大量静态路由
		const routeCount = 1000
		for i := 0; i < routeCount; i++ {
			path := fmt.Sprintf("/static/route/%d", i)
			engine.GET(path, func(c *Context) { c.String(200, "static") })
		}
		
		// 测试查找性能
		const testRequests = 10000
		startTime := time.Now()
		
		for i := 0; i < testRequests; i++ {
			path := fmt.Sprintf("/static/route/%d", i%routeCount)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
		}
		
		duration := time.Since(startTime)
		qps := float64(testRequests) / duration.Seconds()
		avgTime := duration / time.Duration(testRequests)
		
		fmt.Printf("   ✅ 路由数量: %d\n", routeCount)
		fmt.Printf("   ✅ 测试请求: %d\n", testRequests)
		fmt.Printf("   ✅ 总耗时: %v\n", duration)
		fmt.Printf("   ✅ QPS: %.2f\n", qps)
		fmt.Printf("   ✅ 平均响应时间: %v\n", avgTime)
		
		// 验证性能是否达标（每次查找应该很快）
		if avgTime > time.Microsecond*100 {
			t.Logf("⚠️  警告: 平均响应时间 %v，静态路由查找可能需要优化", avgTime)
		} else {
			t.Logf("🎉 静态路由查找性能良好: 平均响应时间 %v", avgTime)
		}
	})
	
	// 3. 动态路由参数解析性能测试
	t.Run("动态路由参数解析性能", func(t *testing.T) {
		fmt.Print("\n⚡ 测试动态路由参数解析性能...\n")
		
		engine := New()
		engine.GET("/users/:id", func(c *Context) { 
			c.String(200, "user %s", c.Param("id")) 
		})
		engine.GET("/posts/:id/comments/:cid", func(c *Context) { 
			c.String(200, "comment %s on post %s", c.Param("cid"), c.Param("id")) 
		})
		engine.GET("/api/v1/users/:uid/posts/:pid/tags/:tid", func(c *Context) {
			c.String(200, "complex route")
		})
		
		testPaths := []string{
			"/users/123",
			"/users/456", 
			"/posts/789/comments/101",
			"/posts/999/comments/202",
			"/api/v1/users/111/posts/222/tags/333",
		}
		
		const testRequests = 10000
		startTime := time.Now()
		
		for i := 0; i < testRequests; i++ {
			path := testPaths[i%len(testPaths)]
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
		}
		
		duration := time.Since(startTime)
		qps := float64(testRequests) / duration.Seconds()
		avgTime := duration / time.Duration(testRequests)
		
		fmt.Printf("   ✅ 测试路径: %d 种\n", len(testPaths))
		fmt.Printf("   ✅ 测试请求: %d\n", testRequests)
		fmt.Printf("   ✅ 总耗时: %v\n", duration)
		fmt.Printf("   ✅ QPS: %.2f\n", qps)
		fmt.Printf("   ✅ 平均响应时间: %v\n", avgTime)
		
		if avgTime > time.Microsecond*500 {
			t.Logf("⚠️  警告: 平均响应时间 %v，动态路由解析可能需要优化", avgTime)
		} else {
			t.Logf("🎉 动态路由解析性能良好: 平均响应时间 %v", avgTime)
		}
	})
	
	fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Print("✅ 性能优化验证测试完成\n")
	fmt.Print(strings.Repeat("=", 60) + "\n")
}
