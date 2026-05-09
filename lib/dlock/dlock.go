package dlock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/towgo/towgo/dao/basedboperat"
)

// 锁存储结构
type DistributedLock struct {
	ID         int64     `xorm:"pk autoincr"`
	Name       string    `xorm:"varchar(255) notnull unique"`
	LockCode   string    `xorm:"varchar(36) notnull"`
	ExpireTime time.Time `xorm:"notnull"`
	CreatedAt  time.Time `xorm:"created notnull"`
}

// 锁管理器配置
type LockManagerConfig struct {
	RetryInterval      time.Duration // 重试间隔，默认 100ms
	DefaultAcquireWait time.Duration // 默认获取锁等待时间，默认 10s
	DefaultLockTTL     time.Duration // 默认锁生存时间，默认 30s
	CleanupInterval    time.Duration // 清理间隔，默认 1分钟
	MaxAcquireRetries  int           // 最大重试次数，0表示无限次
	CleanupBatchSize   int           // 批量清理数量，默认 100
}

// 分布式锁管理器
type DistributedLockManager struct {
	config    LockManagerConfig
	stopChan  chan struct{}
	waitGroup sync.WaitGroup
}

// 创建新的分布式锁管理器
func NewDistributedLockManager(config LockManagerConfig) (*DistributedLockManager, error) {
	// 应用默认配置
	if config.RetryInterval == 0 {
		config.RetryInterval = 100 * time.Millisecond
	}
	if config.DefaultAcquireWait == 0 {
		config.DefaultAcquireWait = 10 * time.Second
	}
	if config.DefaultLockTTL == 0 {
		config.DefaultLockTTL = 30 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Minute
	}
	if config.CleanupBatchSize == 0 {
		config.CleanupBatchSize = 100
	}

	dlm := &DistributedLockManager{

		config:   config,
		stopChan: make(chan struct{}),
	}

	// 启动后台清理协程
	dlm.waitGroup.Add(1)
	go dlm.backgroundCleanup()

	return dlm, nil
}

// 阻塞方式获取锁
func (dlm *DistributedLockManager) Lock(ctx context.Context, lockName string, opts ...LockOption) (string, error) {
	// 应用选项
	options := lockOptions{
		ttl:  dlm.config.DefaultLockTTL,
		wait: dlm.config.DefaultAcquireWait,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	// 设置上下文超时
	if options.wait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.wait)
		defer cancel()
	}

	lockCode := uuid.New().String()
	expireTime := time.Now().Add(options.ttl)

	retryCount := 0
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("lock acquire timeout: %w", ctx.Err())
		default:
			acquired, err := dlm.tryAcquireLock(lockName, lockCode, expireTime)
			if err != nil {
				return "", err
			}

			if acquired {
				return lockCode, nil
			}

			// 检查最大重试次数
			if dlm.config.MaxAcquireRetries > 0 && retryCount >= dlm.config.MaxAcquireRetries {
				return "", errors.New("exceeded maximum acquire retries")
			}
			retryCount++

			// 等待重试
			timer := time.NewTimer(dlm.config.RetryInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", fmt.Errorf("lock acquire timeout: %w", ctx.Err())
			case <-timer.C:
				// 重试前更新过期时间
				expireTime = time.Now().Add(options.ttl)
			}
		}
	}
}

// 尝试获取锁
func (dlm *DistributedLockManager) tryAcquireLock(lockName, lockCode string, expireTime time.Time) (bool, error) {
	session, err := basedboperat.NewTransaction()
	if err != nil {
		return false, err
	}
	if err := session.Begin(); err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 1. 检查是否有过期的锁
	var existing DistributedLock
	err = session.Get(&existing, nil, "name = ?", lockName)
	if err != nil {
		_ = session.Rollback()
		return false, fmt.Errorf("failed to check lock: %w", err)
	}

	var has bool

	if existing.ID > 0 {
		has = true
	}

	// 2. 处理锁不存在或已过期的情况
	if !has || (has && existing.ExpireTime.Before(time.Now())) {
		// 删除过期锁（如果存在）
		if has {

			if _, err := basedboperat.Delete(&DistributedLock{}, existing.ID, nil); err != nil {
				_ = session.Rollback()
				return false, fmt.Errorf("failed to remove expired lock: %w", err)
			}
		}

		// 创建新锁
		newLock := DistributedLock{
			Name:       lockName,
			LockCode:   lockCode,
			ExpireTime: expireTime,
		}

		if _, err := session.Create(&newLock); err != nil {
			_ = session.Rollback()
			return false, fmt.Errorf("failed to create new lock: %w", err)
		}

		if err := session.Commit(); err != nil {
			return false, fmt.Errorf("commit failed: %w", err)
		}

		return true, nil
	}

	_ = session.Rollback()
	return false, nil
}

// 解锁
func (dlm *DistributedLockManager) Unlock(lockName, lockCode string) error {
	session, err := basedboperat.NewTransaction()
	if err != nil {
		return err
	}

	if err := session.Begin(); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 获取当前锁
	var lock DistributedLock
	err = session.Get(&lock, nil, "name = ?", lockName)
	if err != nil {
		_ = session.Rollback()
		return fmt.Errorf("failed to get lock: %w", err)
	}
	var has bool

	if lock.ID > 0 {
		has = true
	}

	// 验证锁的存在和有效性
	if !has {
		return errors.New("lock not found")
	}
	if lock.LockCode != lockCode {
		_ = session.Rollback()
		return errors.New("invalid unlock code")
	}

	// 删除锁
	if _, err := session.Delete(&DistributedLock{}, lock.ID, nil); err != nil {
		_ = session.Rollback()
		return fmt.Errorf("failed to delete lock: %w", err)
	}

	if err := session.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// 续期锁
func (dlm *DistributedLockManager) Renew(ctx context.Context, lockName, lockCode string, ttl time.Duration) error {
	newExpireTime := time.Now().Add(ttl)

	session, err := basedboperat.NewTransaction()
	if err != nil {
		return err
	}

	if err := session.Begin(); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 获取当前锁
	var lock DistributedLock
	err = session.Get(&lock, nil, "name = ? AND lock_code = ?", lockName, lockCode)
	if err != nil {
		_ = session.Rollback()
		return fmt.Errorf("failed to get lock: %w", err)
	}

	if lock.ID == 0 {
		_ = session.Rollback()
		return errors.New("lock not found or code mismatch")
	}

	// 更新过期时间
	lock.ExpireTime = newExpireTime
	if err := session.Update(&lock, nil, "id=?", lock.ID); err != nil {
		_ = session.Rollback()
		return fmt.Errorf("failed to update lock: %w", err)
	}

	if err := session.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// 后台清理过期的锁
func (dlm *DistributedLockManager) backgroundCleanup() {
	defer dlm.waitGroup.Done()

	ticker := time.NewTicker(dlm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dlm.cleanupExpiredLocks()
		case <-dlm.stopChan:
			return
		}
	}
}

// 清理过期锁
func (dlm *DistributedLockManager) cleanupExpiredLocks() {
	// 使用批量清理，避免一次性删除过多记录
	for {
		// 获取过期锁
		var expiredLocks []DistributedLock
		var expiredLock DistributedLock
		var list basedboperat.List
		list.Limit = -1
		list.Where = append(list.Where, basedboperat.Condition{
			Field:    "expire_time",
			Operator: "<",
			Value:    time.Now(),
		})
		basedboperat.ListScan(&list, &expiredLock, &expiredLocks)

		if list.Error != nil {
			log.Print(list.Error.Error())
			return
		}

		if len(expiredLocks) == 0 {
			break
		}

		// 删除过期锁
		ids := make([]int64, 0, len(expiredLocks))
		for _, lock := range expiredLocks {
			ids = append(ids, lock.ID)
		}

		if _, err := basedboperat.Delete(&DistributedLock{}, nil, "id in (?)", ids); err != nil {
			break
		}

		if len(expiredLocks) < dlm.config.CleanupBatchSize {
			break
		}
	}
}

// 关闭锁管理器
func (dlm *DistributedLockManager) Close() {
	close(dlm.stopChan)
	dlm.waitGroup.Wait()
}

// 锁选项类型
type lockOptions struct {
	ttl  time.Duration
	wait time.Duration
}

// 锁选项接口
type LockOption interface {
	apply(*lockOptions)
}

// 锁TTL选项
type WithTTL time.Duration

func (w WithTTL) apply(o *lockOptions) {
	o.ttl = time.Duration(w)
}

// 锁等待时间选项
type WithWait time.Duration

func (w WithWait) apply(o *lockOptions) {
	o.wait = time.Duration(w)
}

// 实用选项函数
func WithTTLOption(ttl time.Duration) LockOption {
	return WithTTL(ttl)
}

func WithWaitOption(wait time.Duration) LockOption {
	return WithWait(wait)
}
