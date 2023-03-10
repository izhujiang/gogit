package common

type NameHashPair struct {
	Oid  Hash
	Name string
	Mode FileMode
}

type NameHashPairs []*NameHashPair
type Change struct {
	From *NameHashPair
	To   *NameHashPair
}
type Changes struct {
	Create []*Change
	Remove []*Change
	Modify []*Change
}

// The tow arguments must be sorted by NameHashPair.Name
func CompareOrderedNameHashPairs(pairA NameHashPairs, pairB NameHashPairs) *Changes {
	// record the change from A to B
	changes := &Changes{}

	i, j := 0, 0

	for i < len(pairA) && j < len(pairB) {
		a := pairA[i].Name
		b := pairB[j].Name
		if a == b {
			if pairA[i].Oid != pairB[j].Oid || pairA[i].Mode != pairB[j].Mode {
				changes.Modify = append(
					changes.Modify,
					&Change{pairA[i], pairB[j]})
			}
			i++
			j++
		} else if a < b {
			changes.Remove = append(
				changes.Remove,
				&Change{pairA[i], nil})
			i++
		} else {
			changes.Create = append(
				changes.Create,
				&Change{nil, pairB[j]})
			j++
		}
	}

	for ; i < len(pairA); i++ {
		changes.Remove = append(
			changes.Remove,
			&Change{pairA[i], nil})
		i++
	}
	for ; j < len(pairB); j++ {
		changes.Create = append(
			changes.Create,
			&Change{nil, pairB[j]})
		j++
	}
	return changes
}
