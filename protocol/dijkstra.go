package protocol

import (
	"math"
)

type Graph struct {
	adjList map[int][]node
	nodes   map[int]bool
}

type node struct {
	n      int
	weight int
}

func NewGraph() *Graph {
	g := &Graph{}
	g.adjList = make(map[int][]node)
	g.nodes = make(map[int]bool)

	return g
}

func (g *Graph) AddEdge(x, y, weight int) {
	n := node{n: y, weight: weight}

	if g.adjList[x] == nil {
		g.adjList[x] = make([]node, 0)
	}

	g.adjList[x] = append(g.adjList[x], n)
	g.nodes[x] = true
	g.nodes[y] = true
}

func (g *Graph) RemoveNode(x int) {
	delete(g.adjList, x)

	for i := range g.adjList {
		newList := make([]node, 0, len(g.adjList[i]))

		for _, node := range g.adjList[i] {
			if node.n == x {
				continue
			}

			newList = append(newList, node)
		}

		g.adjList[i] = newList
	}
}

func nextNodeExists(dist map[int]int) bool {
	for _, d := range dist {
		if d == math.MaxInt {
			continue
		}

		if d == -1 {
			continue
		}

		return true
	}

	return false
}

func minNode(dist map[int]int, visited map[int]bool) (node int) {
	minDist := math.MaxInt
	node = -1

	for i, d := range dist {
		if visited[i] {
			continue
		}

		if d >= minDist {
			continue
		}

		node = i
		minDist = d
	}

	return
}

func (g *Graph) Dijkstra(source int) (next map[int]int, dist map[int]int) {
	next = make(map[int]int, len(g.adjList))
	dist = make(map[int]int, len(g.adjList))
	visited := make(map[int]bool, len(g.adjList))

	for x := range g.nodes {
		dist[x] = math.MaxInt
		next[x] = -1
	}

	dist[source] = 0
	next[source] = source
	visited[source] = true

	for nextNodeExists(dist) {
		// update distances
		for x := range visited {
			for _, node := range g.adjList[x] {
				if visited[node.n] {
					continue
				}

				if dist[x]+node.weight > dist[node.n] {
					continue
				}

				dist[node.n] = dist[x] + node.weight
				next[node.n] = next[x]

				if x == source {
					next[node.n] = node.n
				}
			}
		}

		// add next node
		nextNode := minNode(dist, visited)
		if nextNode == -1 {
			return
		}

		visited[nextNode] = true
	}

	return
}
