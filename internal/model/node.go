package model

import (
	"encoding/json"
	"sync"

	"github.com/cespare/xxhash/v2"
)

const (
	Alive uint64 = 1 << 0
)

type Node struct {
	Raw       []byte
	UniqueKey uint64
	Info      *NodeInfo
}

type NodeInfo struct {
	Delay       uint16
	AliveStatus uint64
	Country     string
}

type UniqueKey struct {
	Server     string `yaml:"server"`
	Servername string `yaml:"servername"`
	Port       string `yaml:"port"`
	Type       string `yaml:"type"`
	Uuid       string `yaml:"uuid"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

func (u *UniqueKey) Gen() uint64 {
	bytes, _ := json.Marshal(u)
	return xxhash.Sum64(bytes)
}

func (i *NodeInfo) SetAliveStatus(status uint64, alive bool) {
	if alive {
		i.AliveStatus |= status
	} else {
		i.AliveStatus &^= status
	}
}

type NodePool struct {
	nodes []Node
	mu    sync.RWMutex
	exist map[uint64]bool
}

func NewNodePool(size int) *NodePool {
	return &NodePool{
		nodes: make([]Node, 0, size),
		exist: make(map[uint64]bool),
	}
}

func (p *NodePool) Add(node Node) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.exist[node.UniqueKey] {
		return false
	}

	p.nodes = append(p.nodes, node)
	p.exist[node.UniqueKey] = true
	return true
}

func (p *NodePool) GetAll() []Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]Node, len(p.nodes))
	copy(result, p.nodes)
	return result
}

func (p *NodePool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.nodes)
}
