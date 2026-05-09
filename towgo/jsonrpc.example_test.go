/*
jsonrpc 中间件和错误处理示例
by:liangliangit

本文件展示如何：
1. 使用中间件系统
2. 在 handler 中返回 result 和 error
3. 在中间件中统一处理 result/error 并调用 Write()
*/
package towgo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/towgo/towgo/towgo"
)

// ============================================================
// 示例: 使用中间件记录请求日志
// ============================================================

func TestMiddleware_Logging(t *testing.T) {
	// 注册日志中间件
	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
		fmt.Printf("收到请求: %s\n", conn.GetRpcRequest().Method)

		// 调用下一个中间件或 handler
		conn.Next()

		// 后置处理：检查是否有错误
		if err := conn.GetError(); err != nil {
			fmt.Printf("请求 %s 处理出错: %v\n", conn.GetRpcRequest().Method, err)
		} else {
			fmt.Printf("请求 %s 处理成功，结果: %v\n", conn.GetRpcRequest().Method, conn.GetResult())
		}
	})
}

// ============================================================
// 示例: 统一响应处理中间件（放在最后，负责 Write）
// ============================================================

func TestMiddleware_ResponseWriter(t *testing.T) {
	// 注册响应处理中间件（通常放在最后）
	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
		conn.Next()

		// 后置处理：根据 result/error 生成响应
		if err := conn.GetError(); err != nil {
			// 根据错误类型返回不同的响应
			if gErr := gerror.Current(err); gErr != nil {
				conn.WriteError(500, err.Error())
			} else {
				conn.WriteError(500, err.Error())
			}
		} else {
			// 正常结果直接 Write
			conn.Write()
		}
	})
}

// ============================================================
// 示例: 在 handler 中使用 SetResult/WriteResult/WithError
// ============================================================

func TestHandler_SetResult(t *testing.T) {
	// 注册 handler
	towgo.Handle("user.get", func(conn towgo.JsonRpcConnection) {
		var req struct {
			ID int64 `json:"id"`
		}

		// 解析参数
		if err := conn.ReadParams(&req); err != nil {
			conn.WithError(err)
			return
		}

		// 业务逻辑
		if req.ID <= 0 {
			conn.WithError(gerror.NewCode(gcode.CodeInvalidParameter, "id必须大于0"))
			return
		}

		// 方式1: WriteResult 立即发送（简单场景）
		conn.WriteResult(map[string]interface{}{
			"id":   req.ID,
			"name": "张三",
		})

		// 方式2: SetResult 只设置，由中间件决定何时 Write
		// conn.SetResult(map[string]interface{}{"id": req.ID})
	})
}

// ============================================================
// 示例: 认证中间件
// ============================================================

func TestMiddleware_Auth(t *testing.T) {
	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
		// 公开接口不需要认证
		method := conn.GetRpcRequest().Method
		if method == "public.login" || method == "public.register" {
			conn.Next()
			return
		}

		// 检查 session/token
		session := conn.GetRpcRequest().Session
		if session == "" {
			conn.WithError(gerror.New("未登录"))

			return
		}

		// 验证通过，继续执行
		conn.Next()
	})
}

// ============================================================
// 示例: 使用 BindObject 注册带 context 的 handler
// ============================================================

type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, req *UserGetReq) (*UserGetRes, error) {
	// 注意：ctx 是 context.Context
	// req 会自动从请求参数解析
	// 返回 error 时会自动设置到 conn

	if req.ID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id必须大于0")
	}

	return &UserGetRes{
		ID:   req.ID,
		Name: "张三",
	}, nil
}

type UserGetReq struct {
	ID int64 `json:"id"`
}

type UserGetRes struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func TestBindObject(t *testing.T) {
	// BindObject 会自动处理 context 和错误
	towgo.BindObject("user.", &UserService{})
}

// ============================================================
// 示例: 洋葱模型执行顺序
// ============================================================

func TestMiddleware_ExecutionOrder(t *testing.T) {
	var order []string

	// 先清空已有中间件
	// 注意：这会影响其他测试，生产环境中不要这样做
	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
		order = append(order, "1-前")

		conn.Next()
		order = append(order, "5-后")
	})

	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
		order = append(order, "2-前")
		conn.Next()
		order = append(order, "4-后")
	})

	// 注册测试 handler
	towgo.Handle("test.order", func(conn towgo.JsonRpcConnection) {
		order = append(order, "3-Handler")
		// 立即发送响应
		conn.WriteResult(nil)
	})

	// 使用 LocalRpcConnection 模拟请求
	req := towgo.NewJsonrpcrequest()
	req.Method = "test.order"
	req.Id = "test-123"

	conn := towgo.NewLocalRpcConnection(req, nil)
	towgo.TestExec(conn)

	// 预期执行顺序: 1前 -> 2前 -> Handler -> 2后 -> 1后
	expected := []string{"1-前", "2-前", "3-Handler", "4-后", "5-后"}

	t.Logf("执行顺序: %v", order)

	if len(order) != len(expected) {
		t.Fatalf("执行顺序长度不匹配，期望 %d 个，实际 %d 个", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("第 %d 个执行顺序错误，期望 %q，实际 %q", i, v, order[i])
		}
	}
}
