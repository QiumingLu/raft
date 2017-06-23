func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}
