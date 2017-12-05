package set

import (
	"sort"

	"github.com/davidwalter0/forwarder/pipe"
)

// Keys returns a slice of Keys from the map
func Keys(m *map[string]*pipe.Definition) (keys []string) {
	for k := range *m {
		keys = append(keys, k)
	}
	return
}

// Sort returns a slice of sorted keys from the map
func Sort(m *map[string]*pipe.Definition) (keys []string) {
	sort.Strings(Keys(m))
	return
}

// Difference between 2 maps returns left only (not in right), common
// to both, right only (not in left)
func Difference(lhs *map[string]*pipe.Definition, rhs *map[string]*pipe.Definition) (LOnly, Common, ROnly []string) {

	var set = make(map[string]bool)
	for k := range *lhs {
		if _, ok := (*rhs)[k]; !ok {
			LOnly = append(LOnly, k)
		} else {
			if _, ok := set[k]; !ok {
				Common = append(Common, k)
			}
		}
	}
	for k := range *rhs {
		if _, ok := (*lhs)[k]; !ok {
			ROnly = append(ROnly, k)
		} else {
			if _, ok := set[k]; !ok {
				Common = append(Common, k)
			}
		}
	}
	return
}

/////////// // Keys returns a slice of Keys from the map
/////////// func Keys(m *map[string]pipe.Definition) (keys []string) {
/////////// 	for k := range *m {
/////////// 		keys = append(keys, k)
/////////// 	}
/////////// 	return
/////////// }

/////////// // Sort returns a slice of sorted keys from the map
/////////// func Sort(m *map[string]pipe.Definition) (keys []string) {
/////////// 	sort.Strings(Keys(m))
/////////// 	return
/////////// }

/////////// // Difference between 2 maps returns left only (not in right), common
/////////// // to both, right only (not in left)
/////////// func Difference(lhs *map[string]pipe.Definition, rhs *map[string]pipe.Definition) (LOnly, Common, ROnly []string) {

/////////// 	var set = make(map[string]bool)
/////////// 	for k := range *lhs {
/////////// 		if _, ok := (*rhs)[k]; !ok {
/////////// 			LOnly = append(LOnly, k)
/////////// 		} else {
/////////// 			if _, ok := set[k]; !ok {
/////////// 				Common = append(Common, k)
/////////// 			}
/////////// 		}
/////////// 	}
/////////// 	for k := range *rhs {
/////////// 		if _, ok := (*lhs)[k]; !ok {
/////////// 			ROnly = append(ROnly, k)
/////////// 		} else {
/////////// 			if _, ok := set[k]; !ok {
/////////// 				Common = append(Common, k)
/////////// 			}
/////////// 		}
/////////// 	}
/////////// 	return
/////////// }
