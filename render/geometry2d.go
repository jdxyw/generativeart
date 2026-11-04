package render

import (
	"math"
	"sort"
)

// DelaunayTriangle represents a single triangle in a 2D Delaunay triangulation
// with vertices (Sites[A], Sites[B], Sites[C]) oriented counter-clockwise.
type DelaunayTriangle struct {
	A, B, C        int
	Circumcenter   Point
	CircumradiusSq float64
}

// VoronoiCell2D represents the dual bounded Voronoi polygon for Sites[SiteIndex],
// clipped against the domain bounding box [BoundsMin, BoundsMax].
type VoronoiCell2D struct {
	SiteIndex int
	Site      Point
	Vertices  []Point
	Area      float64
	Centroid  Point
	Neighbors []int
}

// DelaunayVoronoiMesh2D holds a complete 2D Bowyer-Watson Delaunay triangulation,
// its unique undirected edge graph, and its dual clipped Voronoi polygon diagram.
type DelaunayVoronoiMesh2D struct {
	Sites     []Point
	BoundsMin Point
	BoundsMax Point
	Triangles []DelaunayTriangle
	Cells     []VoronoiCell2D
	Edges     [][2]int
}

func makeDelaunayTriangle(pts []Point, a, b, c int) (DelaunayTriangle, bool) {
	pa, pb, pc := pts[a], pts[b], pts[c]
	// Ensure counter-clockwise orientation
	cross := (pb.X-pa.X)*(pc.Y-pa.Y) - (pb.Y-pa.Y)*(pc.X-pa.X)
	if math.Abs(cross) < 1e-11 {
		return DelaunayTriangle{}, false
	}
	if cross < 0 {
		b, c = c, b
		pb, pc = pc, pb
	}

	ax, ay := pa.X, pa.Y
	bx, by := pb.X, pb.Y
	cx, cy := pc.X, pc.Y
	d := 2.0 * (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by))
	if math.Abs(d) < 1e-11 {
		return DelaunayTriangle{}, false
	}
	aSq := ax*ax + ay*ay
	bSq := bx*bx + by*by
	cSq := cx*cx + cy*cy
	ux := (aSq*(by-cy) + bSq*(cy-ay) + cSq*(ay-by)) / d
	uy := (aSq*(cx-bx) + bSq*(ax-cx) + cSq*(bx-ax)) / d
	dx := ux - ax
	dy := uy - ay
	return DelaunayTriangle{
		A:              a,
		B:              b,
		C:              c,
		Circumcenter:   Pt(ux, uy),
		CircumradiusSq: dx*dx + dy*dy,
	}, true
}

// PolygonAreaAndCentroid2D computes the unsigned area and exact geometric centroid
// of a closed 2D polygon using Green's theorem (shoelace formula).
func PolygonAreaAndCentroid2D(poly []Point) (area float64, centroid Point) {
	n := len(poly)
	if n == 0 {
		return 0, Point{}
	}
	if n < 3 {
		var sx, sy float64
		for _, p := range poly {
			sx += p.X
			sy += p.Y
		}
		return 0, Pt(sx/float64(n), sy/float64(n))
	}

	var signedArea2, cx6, cy6 float64
	for i := 0; i < n; i++ {
		p0 := poly[i]
		p1 := poly[(i+1)%n]
		cross := p0.X*p1.Y - p1.X*p0.Y
		signedArea2 += cross
		cx6 += (p0.X + p1.X) * cross
		cy6 += (p0.Y + p1.Y) * cross
	}
	if math.Abs(signedArea2) < 1e-11 {
		var sx, sy float64
		for _, p := range poly {
			sx += p.X
			sy += p.Y
		}
		return 0, Pt(sx/float64(n), sy/float64(n))
	}
	inv3A := 1.0 / (3.0 * signedArea2)
	return 0.5 * math.Abs(signedArea2), Pt(cx6*inv3A, cy6*inv3A)
}

// ClipPolygonHalfPlane2D clips a convex or concave polygon against the half-plane
// { p : (p - planePt) . inwardNormal >= 0 } using Sutherland-Hodgman clipping.
func ClipPolygonHalfPlane2D(poly []Point, planePt, inwardNormal Point) []Point {
	n := len(poly)
	if n < 3 {
		return nil
	}
	out := make([]Point, 0, n+2)
	eval := func(p Point) float64 {
		return (p.X-planePt.X)*inwardNormal.X + (p.Y-planePt.Y)*inwardNormal.Y
	}

	prev := poly[n-1]
	prevVal := eval(prev)
	for i := 0; i < n; i++ {
		cur := poly[i]
		curVal := eval(cur)
		if curVal >= -1e-9 {
			if prevVal < -1e-9 {
				t := prevVal / (prevVal - curVal)
				out = append(out, prev.Lerp(cur, t))
			}
			out = append(out, cur)
		} else if prevVal >= -1e-9 {
			t := prevVal / (prevVal - curVal)
			out = append(out, prev.Lerp(cur, t))
		}
		prev = cur
		prevVal = curVal
	}
	return out
}

// BuildDelaunayVoronoi2D constructs the Bowyer-Watson Delaunay triangulation and dual
// clipped Voronoi polygon mesh for sites inside [boundsMin, boundsMax].
func BuildDelaunayVoronoi2D(sites []Point, boundsMin, boundsMax Point) *DelaunayVoronoiMesh2D {
	n := len(sites)
	mesh := &DelaunayVoronoiMesh2D{
		Sites:     make([]Point, n),
		BoundsMin: boundsMin,
		BoundsMax: boundsMax,
		Cells:     make([]VoronoiCell2D, n),
	}
	copy(mesh.Sites, sites)
	if n == 0 {
		return mesh
	}

	minX, minY := boundsMin.X, boundsMin.Y
	maxX, maxY := boundsMax.X, boundsMax.Y
	for _, s := range sites {
		if s.X < minX {
			minX = s.X
		}
		if s.X > maxX {
			maxX = s.X
		}
		if s.Y < minY {
			minY = s.Y
		}
		if s.Y > maxY {
			maxY = s.Y
		}
	}
	span := math.Max(maxX-minX, maxY-minY)
	if span < 1.0 {
		span = 1.0
	}
	midX := 0.5 * (minX + maxX)
	midY := 0.5 * (minY + maxY)

	// Append 3 super-triangle vertices at indices n, n+1, n+2
	allPts := make([]Point, n+3)
	for i := 0; i < n; i++ {
		// Deterministic sub-ULP jitter prevents cocircular/collinear degeneracies
		epsX := float64((i*73856093)&1023) * 1e-9
		epsY := float64((i*19349663)&1023) * 1e-9
		allPts[i] = Pt(sites[i].X+epsX, sites[i].Y+epsY)
	}
	allPts[n] = Pt(midX-4.0*span, midY-3.0*span)
	allPts[n+1] = Pt(midX+4.0*span, midY-3.0*span)
	allPts[n+2] = Pt(midX, midY+5.0*span)

	var tris []DelaunayTriangle
	if st, ok := makeDelaunayTriangle(allPts, n, n+1, n+2); ok {
		tris = append(tris, st)
	}

	type edgeKey struct{ u, v int }
	for i := 0; i < n; i++ {
		pt := allPts[i]
		var badIdx []int
		for tIdx, tr := range tris {
			dx := pt.X - tr.Circumcenter.X
			dy := pt.Y - tr.Circumcenter.Y
			if dx*dx+dy*dy <= tr.CircumradiusSq*(1.0+1e-11) {
				badIdx = append(badIdx, tIdx)
			}
		}

		edgeCount := make(map[edgeKey]int, len(badIdx)*3)
		dirEdges := make([][2]int, 0, len(badIdx)*3)
		for _, bIdx := range badIdx {
			tr := tris[bIdx]
			triEdges := [3][2]int{{tr.A, tr.B}, {tr.B, tr.C}, {tr.C, tr.A}}
			for _, e := range triEdges {
				u, v := e[0], e[1]
				if u > v {
					u, v = v, u
				}
				k := edgeKey{u, v}
				edgeCount[k]++
				dirEdges = append(dirEdges, e)
			}
		}

		// Retain good triangles
		if len(badIdx) > 0 {
			keep := tris[:0]
			badPos := 0
			for tIdx, tr := range tris {
				if badPos < len(badIdx) && badIdx[badPos] == tIdx {
					badPos++
					continue
				}
				keep = append(keep, tr)
			}
			tris = keep
		}

		// Re-triangulate cavity boundary edges with vertex i
		for _, e := range dirEdges {
			u, v := e[0], e[1]
			if u > v {
				u, v = v, u
			}
			if edgeCount[edgeKey{u, v}] == 1 {
				if nt, ok := makeDelaunayTriangle(allPts, e[0], e[1], i); ok {
					tris = append(tris, nt)
				}
			}
		}
	}

	// Filter out triangles connected to the super-triangle and build adjacency
	neighborMap := make([]map[int]struct{}, n)
	for i := 0; i < n; i++ {
		neighborMap[i] = make(map[int]struct{})
	}
	uniqueEdges := make(map[edgeKey]struct{})

	finalTris := make([]DelaunayTriangle, 0, len(tris))
	for _, tr := range tris {
		if tr.A >= n || tr.B >= n || tr.C >= n {
			continue
		}
		finalTris = append(finalTris, tr)
		pairs := [3][2]int{{tr.A, tr.B}, {tr.B, tr.C}, {tr.C, tr.A}}
		for _, p := range pairs {
			u, v := p[0], p[1]
			neighborMap[u][v] = struct{}{}
			neighborMap[v][u] = struct{}{}
			if u > v {
				u, v = v, u
			}
			uniqueEdges[edgeKey{u, v}] = struct{}{}
		}
	}
	mesh.Triangles = finalTris

	mesh.Edges = make([][2]int, 0, len(uniqueEdges))
	for k := range uniqueEdges {
		mesh.Edges = append(mesh.Edges, [2]int{k.u, k.v})
	}
	sort.Slice(mesh.Edges, func(i, j int) bool {
		if mesh.Edges[i][0] != mesh.Edges[j][0] {
			return mesh.Edges[i][0] < mesh.Edges[j][0]
		}
		return mesh.Edges[i][1] < mesh.Edges[j][1]
	})

	// Construct dual clipped Voronoi polygons via half-plane bisector clipping
	baseBox := []Point{
		Pt(boundsMin.X, boundsMin.Y),
		Pt(boundsMax.X, boundsMin.Y),
		Pt(boundsMax.X, boundsMax.Y),
		Pt(boundsMin.X, boundsMax.Y),
	}

	for i := 0; i < n; i++ {
		si := sites[i]
		nbs := make([]int, 0, len(neighborMap[i]))
		for nb := range neighborMap[i] {
			nbs = append(nbs, nb)
		}
		sort.Slice(nbs, func(a, b int) bool {
			return si.Dist(sites[nbs[a]]) < si.Dist(sites[nbs[b]])
		})

		poly := make([]Point, len(baseBox))
		copy(poly, baseBox)

		// First clip against Delaunay neighbors (sorted nearest first)
		clippedSet := make(map[int]bool, len(nbs)+8)
		clippedSet[i] = true
		for _, j := range nbs {
			clippedSet[j] = true
			sj := sites[j]
			mid := si.Lerp(sj, 0.5)
			inward := si.Sub(sj)
			if inward.Len() < 1e-9 {
				continue
			}
			poly = ClipPolygonHalfPlane2D(poly, mid, inward)
			if len(poly) < 3 {
				break
			}
		}

		// Safety pass: for boundary cells where a super-triangle edge was removed, check any other sites within 2 * maxRadius
		if len(poly) >= 3 {
			maxR := 0.0
			for _, v := range poly {
				if d := si.Dist(v); d > maxR {
					maxR = d
				}
			}
			for j := 0; j < n; j++ {
				if clippedSet[j] {
					continue
				}
				sj := sites[j]
				if si.Dist(sj) > 2.0*maxR+1e-6 {
					continue
				}
				mid := si.Lerp(sj, 0.5)
				inward := si.Sub(sj)
				if inward.Len() < 1e-9 {
					continue
				}
				poly = ClipPolygonHalfPlane2D(poly, mid, inward)
				if len(poly) < 3 {
					break
				}
				maxR = 0.0
				for _, v := range poly {
					if d := si.Dist(v); d > maxR {
						maxR = d
					}
				}
			}
		}

		area, centroid := PolygonAreaAndCentroid2D(poly)
		if area <= 1e-9 {
			centroid = si
		}
		mesh.Cells[i] = VoronoiCell2D{
			SiteIndex: i,
			Site:      si,
			Vertices:  poly,
			Area:      area,
			Centroid:  centroid,
			Neighbors: nbs,
		}
	}

	return mesh
}

// RelaxVoronoiPolygonCVT performs exact geometric Lloyd's Centroidal Voronoi Tessellation
// relaxation over iterations steps by moving each site to the exact Green's-theorem centroid
// of its clipped Voronoi polygon.
func RelaxVoronoiPolygonCVT(sites []Point, boundsMin, boundsMax Point, iterations int) []Point {
	cur := make([]Point, len(sites))
	copy(cur, sites)
	if len(cur) <= 1 || iterations <= 0 {
		return cur
	}
	for iter := 0; iter < iterations; iter++ {
		mesh := BuildDelaunayVoronoi2D(cur, boundsMin, boundsMax)
		for i, cell := range mesh.Cells {
			if cell.Area > 1e-6 {
				cur[i] = cell.Centroid
			}
		}
	}
	return cur
}

// ExtractGabrielGraph filters the Delaunay edges to return the Gabriel Graph:
// all edges (i, j) whose open diametral circle contains no other site k.
func (m *DelaunayVoronoiMesh2D) ExtractGabrielGraph() [][2]int {
	if m == nil || len(m.Edges) == 0 {
		return nil
	}
	var gabriel [][2]int
	for _, e := range m.Edges {
		i, j := e[0], e[1]
		si, sj := m.Sites[i], m.Sites[j]
		mid := si.Lerp(sj, 0.5)
		rSq := 0.25 * ((si.X-sj.X)*(si.X-sj.X) + (si.Y-sj.Y)*(si.Y-sj.Y))

		empty := true
		// Check neighbors of i and j first (sufficient for Delaunay subgraph Gabriel test)
		for _, k := range m.Cells[i].Neighbors {
			if k == j {
				continue
			}
			sk := m.Sites[k]
			dx, dy := sk.X-mid.X, sk.Y-mid.Y
			if dx*dx+dy*dy < rSq*(1.0-1e-9) {
				empty = false
				break
			}
		}
		if empty {
			for _, k := range m.Cells[j].Neighbors {
				if k == i {
					continue
				}
				sk := m.Sites[k]
				dx, dy := sk.X-mid.X, sk.Y-mid.Y
				if dx*dx+dy*dy < rSq*(1.0-1e-9) {
					empty = false
					break
				}
			}
		}
		if empty {
			gabriel = append(gabriel, e)
		}
	}
	return gabriel
}

// ExtractMinimumSpanningTree computes the exact Euclidean Minimum Spanning Tree (EMST)
// in O(E log E) time by running Kruskal's algorithm over the Delaunay edges.
func (m *DelaunayVoronoiMesh2D) ExtractMinimumSpanningTree() [][2]int {
	if m == nil || len(m.Sites) <= 1 || len(m.Edges) == 0 {
		return nil
	}
	type weightedEdge struct {
		u, v   int
		distSq float64
	}
	wEdges := make([]weightedEdge, len(m.Edges))
	for idx, e := range m.Edges {
		si, sj := m.Sites[e[0]], m.Sites[e[1]]
		dx, dy := si.X-sj.X, si.Y-sj.Y
		wEdges[idx] = weightedEdge{u: e[0], v: e[1], distSq: dx*dx + dy*dy}
	}
	sort.Slice(wEdges, func(i, j int) bool {
		return wEdges[i].distSq < wEdges[j].distSq
	})

	parent := make([]int, len(m.Sites))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}

	mst := make([][2]int, 0, len(m.Sites)-1)
	for _, we := range wEdges {
		ru, rv := find(we.u), find(we.v)
		if ru != rv {
			parent[ru] = rv
			mst = append(mst, [2]int{we.u, we.v})
			if len(mst) == len(m.Sites)-1 {
				break
			}
		}
	}
	return mst
}
