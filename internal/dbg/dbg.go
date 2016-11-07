package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aurelien-rainone/assertgo"
	detour "github.com/aurelien-rainone/go-detour"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func main() {
	var (
		f    *os.File
		err  error
		mesh *detour.DtNavMesh
	)

	f, err = os.Open("testdata/navmesh.bin")
	check(err)
	defer f.Close()

	mesh, err = detour.Decode(f)
	check(err)
	if mesh == nil {
		fmt.Println("error loading mesh")
		return
	}
	fmt.Println("mesh loaded successfully")
	fmt.Printf("mesh params: %#v\n", mesh.Params)
	fmt.Println("Navigation Query")

	org := [3]float32{3, 0, 1}
	dst := [3]float32{50, 0, 30}

	// findPath ok with:
	//org := [3]float32{5, 0, 10}
	//dst := [3]float32{50, 0, 30}

	path, err := findPath(mesh, org, dst)
	if err != nil {
		log.Fatalln("findPath failed", err)
	}
	log.Println("findPath success, path:", path)
}

func findPath(mesh *detour.DtNavMesh, org, dst [3]float32) ([]detour.DtPolyRef, error) {
	var (
		orgRef, dstRef detour.DtPolyRef       // references of org/dst polygon refs
		query          *detour.DtNavMeshQuery // the query instance
		filter         *detour.DtQueryFilter  // filter to use for various queries
		extents        [3]float32             // search distance for polygon search (3 axis)
		nearestPt      [3]float32
		st             detour.DtStatus
		path           []detour.DtPolyRef
	)

	query, st = detour.NewDtNavMeshQuery(mesh, 1000)
	if detour.DtStatusFailed(st) {
		return path, fmt.Errorf("query creation failed with status 0x%x\n", st)
	}
	// define the extents vector for the nearest polygon query
	extents = [3]float32{0, 2, 0}

	// create a default query filter
	filter = detour.NewDtQueryFilter()

	// get org polygon reference
	st = query.FindNearestPoly(org[:], extents[:], filter, &orgRef, nearestPt[:])
	if detour.DtStatusFailed(st) {
		return path, fmt.Errorf("FindNearestPoly failed with 0x%x\n", st)
	} else if orgRef == 0 {
		return path, fmt.Errorf("org doesn't intersect any polygons")
	}
	assert.True(mesh.IsValidPolyRef(orgRef), "%d is not a valid poly ref")
	copy(org[:], nearestPt[:])
	log.Println("org is now", org)

	// get dst polygon reference
	st = query.FindNearestPoly(dst[:], extents[:], filter, &dstRef, nearestPt[:])
	if detour.DtStatusFailed(st) {
		return path, fmt.Errorf("FindNearestPoly failed with 0x%x\n", st)
	} else if dstRef == 0 {
		return path, fmt.Errorf("dst doesn't intersect any polygons")
	}
	assert.True(mesh.IsValidPolyRef(orgRef), "%d is not a valid poly ref")
	copy(dst[:], nearestPt[:])
	log.Println("dst is now", dst)

	// FindPath
	var (
		pathCount int32
	)
	path = make([]detour.DtPolyRef, 100)
	st = query.FindPath(orgRef, dstRef, org[:], dst[:], filter, &path, &pathCount, 100)
	if detour.DtStatusFailed(st) {
		return path, fmt.Errorf("query.FindPath failed with 0x%x\n", st)
	}
	return path[:pathCount], nil

	//fmt.Println("FindPath", "org:", org, "dst:", dst, "orgRef:", orgRef, "dstRef:", dstRef)
	//fmt.Println("FindPath set pathCount to", pathCount)
	//fmt.Println("path", path)
	//fmt.Println("actual path returned", path[:pathCount])

	//// If the end polygon cannot be reached through the navigation graph,
	//// the last polygon in the path will be the nearest the end polygon.
	//// check for that
	//if path[len(path)-1] == dstRef {
	//fmt.Println("no path found, as last poly in path in dstPoly")
	//} else {
	//fmt.Println("path found")
	//for _, polyRef := range path[:pathCount] {
	//fmt.Println("-poly ref", polyRef)
	//mesh.TileAndPolyByRefUnsafe(polyRef, &ptile, &ppoly)
	//polyIdx := mesh.DecodePolyIdPoly(polyRef)
	//poly := ptile.Polys[polyIdx]

	//centroid := make([]float32, 3)
	//detour.DtCalcPolyCenter(centroid, poly.Verts[:], int32(poly.VertCount), ptile.Verts)
	//fmt.Println("poly center: ", centroid)

	////for _, v := range poly.Verts[0:poly.VertCount] {
	////fmt.Println("poly vertex", ptile.Verts[3*v:3*v+3])
	////}
	//}
	//}
	return path, nil
}