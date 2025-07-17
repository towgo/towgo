package www

import "strings"

type PathIterator struct {
	parts []string // 分割后的路径片段
	index int      // 当前迭代位置
}

func NewPathIterator(path string) *PathIterator {
	split := strings.Split(path, "/")
	parts := make([]string, 0, len(split))

	// 清理空路径片段
	for _, p := range split {
		if p != "" {
			parts = append(parts, p)
		}
	}

	return &PathIterator{
		parts: parts,
		index: len(parts), // 初始位置设为路径深度
	}
}

func (it *PathIterator) Next() string {
	if it.index < 0 {
		return ""
	}
	defer func() { it.index-- }()

	if it.index == 0 {
		return "/"
	}

	return "/" + strings.Join(it.parts[:it.index], "/")
}
